package main

import (
	"flag"
	"fmt"
	"log"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	conf "github.com/cyberark/secretless-broker/bin/juxtaposer/config"
	"github.com/cyberark/secretless-broker/bin/juxtaposer/formatter"
	formatter_api "github.com/cyberark/secretless-broker/bin/juxtaposer/formatter/api"
	tester_api "github.com/cyberark/secretless-broker/bin/juxtaposer/tester/api"
	"github.com/cyberark/secretless-broker/bin/juxtaposer/tester/db"
	"github.com/cyberark/secretless-broker/bin/juxtaposer/timing"
	"github.com/cyberark/secretless-broker/bin/juxtaposer/util"
)

type TestManager struct {
	BackendNames               []string
	BackendInstances           map[string]tester_api.DriverManager
	BaselineBackendName        string
	IsShuttingDown             bool
	LatestBaselineTestDuration time.Duration
	MaxRounds                  int
	RoundsCompleted            map[string]int
	RoundsCompletedLock        sync.RWMutex
	RoundsMaxReachedWaitGroup  sync.WaitGroup
	ShutdownChannel            chan bool
	ThreadRunnersWaitGroup     sync.WaitGroup
	Threads                    int
}

type TestRunOptions struct {
	BackendName string
	Round       int
}

func createBackendTesters(config *conf.Config,
	baselineBackendName string) ([]string, map[string]tester_api.DriverManager, error) {

	// Keys in a map are not guaranteed to be retrieved in the same order
	// each time so we have a separate array that guarantees it
	backendNames := []string{}
	backendInstances := map[string]tester_api.DriverManager{}

	for backendName, backendConfig := range config.Backends {
		backendNames = append(backendNames, backendName)

		log.Printf("Setting up backend: %s", backendName)

		connectionType := "persistent"
		if config.Comparison.RecreateConnections {
			connectionType = "recreate"
		}

		options := tester_api.DbTesterOptions{
			ConnectionType: connectionType,
			DatabaseName:   backendConfig.Database,
			Debug:          backendConfig.Debug,
			Host:           backendConfig.Host,
			Password:       backendConfig.Password,
			Port:           backendConfig.Port,
			SslMode:        backendConfig.SslMode,
			Socket:         backendConfig.Socket,
			Username:       backendConfig.Username,
		}

		if backendConfig.Debug {
			fmt.Printf("%s %s %s\n",
				strings.Repeat("v", 35),
				backendName,
				strings.Repeat("v", 35))
		}

		backendTestManager, err := db.NewTestDriver(backendName, config.Driver,
			config.Comparison.SqlStatementType, options)

		if backendConfig.Debug {
			fmt.Printf("%s\n", strings.Repeat("^", 85))
		}

		if err != nil {
			return nil, nil, err
		}

		backendInstances[backendName] = backendTestManager
	}

	// Sort backendNames for consistent output
	sort.Strings(backendNames)

	// Place baseline backend first
	backendBaselineNameIndex := sort.SearchStrings(backendNames, baselineBackendName)
	backendNames = append(backendNames[:backendBaselineNameIndex], backendNames[backendBaselineNameIndex+1:]...)
	backendNames = append([]string{baselineBackendName}, backendNames...)

	return backendNames, backendInstances, nil
}

func applyExitConditions(config *conf.Config, requestedDurationString string,
	shutdownChannel chan<- bool) (int, error) {

	var err error

	requestedDuration := timing.ZeroDuration
	if requestedDurationString != "" {
		requestedDuration, err = time.ParseDuration(requestedDurationString)
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("Using test duration of %v (overriding any 'rounds' from configfile)",
			requestedDuration)

		config.Comparison.Rounds = "infinity"
	}

	rounds := -1
	if config.Comparison.Rounds != "infinity" {
		rounds, err = strconv.Atoi(config.Comparison.Rounds)
		if err != nil {
			return -1, err
		}
	}

	if requestedDuration != timing.ZeroDuration {
		go func() {
			time.Sleep(requestedDuration)
			log.Println("Timeout reached!")
			log.Println("Sending shutdown signal!")
			shutdownChannel <- true
		}()
	}

	return rounds, nil
}

func processAllResults(backendNames []string, formatters map[string]formatter_api.FormatterOptions,
	aggregateTimings *timing.AggregateTimings) error {

	for formatterName, formatterOptions := range formatters {
		log.Printf("Processing output formatter '%s'...", formatterName)

		formatterType := formatterOptions["type"]
		if formatterType == "" {
			formatterType = formatterName
		}

		formatter, err := formatter.GetFormatter(formatterType, formatterOptions)
		if err != nil {
			return err
		}

		formatter.ProcessResults(backendNames, aggregateTimings.Timings,
			aggregateTimings.BaselineMaxThresholdPercent)
	}

	return nil
}

func runTest(backendName string, threadIndex int, aggregateTimings *timing.AggregateTimings,
	testManager *TestManager, round int) time.Duration {

	backendInstance := testManager.BackendInstances[backendName]
	singleTestRunDuration, testErr := backendInstance.RunSingleTest()

	if backendInstance.GetName() == testManager.BaselineBackendName {
		testManager.LatestBaselineTestDuration = singleTestRunDuration
	}

	aggregateTimings.AddTimingData(&timing.SingleRunTiming{
		BaselineTestDuration: testManager.LatestBaselineTestDuration,
		BackendName:          backendInstance.GetName(),
		Duration:             singleTestRunDuration,
		Round:                round,
		TestError:            testErr,
		Thread:               threadIndex,
	})

	return singleTestRunDuration
}

func createThreadedRunner(backendName string, threadIndex int, testManager *TestManager,
	aggregateTimings *timing.AggregateTimings) {

	var once sync.Once

	go func() {
		for {
			if testManager.IsShuttingDown {
				testManager.ThreadRunnersWaitGroup.Done()
				return
			}

			testManager.RoundsCompletedLock.Lock()

			testManager.RoundsCompleted[backendName]++
			round := testManager.RoundsCompleted[backendName]

			testManager.RoundsCompletedLock.Unlock()

			if testManager.MaxRounds != -1 && round > testManager.MaxRounds {
				once.Do(func() {
					testManager.RoundsMaxReachedWaitGroup.Done()
				})
			}

			testDuration := runTest(backendName, threadIndex, aggregateTimings, testManager,
				round)

			if backendName == testManager.BaselineBackendName {
				testManager.LatestBaselineTestDuration = testDuration
			}
		}
	}()
}

func runMainTestingLoop(config *conf.Config, testManager *TestManager) (*timing.AggregateTimings, error) {
	aggregateTimings := timing.NewAggregateTimings(&timing.AggregateTimingOptions{
		BackendNames:                testManager.BackendNames,
		BaselineBackendName:         testManager.BaselineBackendName,
		BaselineMaxThresholdPercent: config.Comparison.BaselineMaxThresholdPercent,
		MaxRounds:                   testManager.MaxRounds,
		Threads:                     config.Comparison.Threads,
		Silent:                      config.Comparison.Silent,
	})

	go func() {
		<-testManager.ShutdownChannel
		log.Println("async: Setting thread exit flag...")
		testManager.IsShuttingDown = true
	}()

	// Initialize a backendBaseline duration to avoid race condition on first run
	initialBaselineTestDuration, testErr := testManager.
		BackendInstances[testManager.BaselineBackendName].RunSingleTest()
	if testErr != nil {
		return nil, testErr
	}
	testManager.LatestBaselineTestDuration = initialBaselineTestDuration

	for _, backendName := range testManager.BackendNames {
		for threadIndex := 0; threadIndex < testManager.Threads; threadIndex++ {
			log.Printf("Creating thread %s[%d]", backendName, threadIndex)
			createThreadedRunner(backendName, threadIndex, testManager,
				&aggregateTimings)

			testManager.ThreadRunnersWaitGroup.Add(1)
			testManager.RoundsMaxReachedWaitGroup.Add(1)
		}
	}

	go func() {
		if testManager.MaxRounds != -1 {
			testManager.RoundsMaxReachedWaitGroup.Wait()
			log.Println("async: Max rounds reached on all backends. Sending shutdown signal...")
			testManager.ShutdownChannel <- true
		}
	}()

	log.Println("async: Waiting for all tests to complete...")
	testManager.ThreadRunnersWaitGroup.Wait()

	aggregateTimings.Process()

	return &aggregateTimings, nil
}

func main() {
	log.Println("Juxtaposer starting up...")

	configFile := flag.String("f", "juxtaposer.yml", "Location of the configuration file.")
	continueRunningAfterExit := flag.Bool("c", false, "Continue running after exit")
	requestedDurationString := flag.String("t", "",
		"Duration of test (ignores 'rounds' field in the configuration.")
	flag.Parse()

	log.Printf("Using configuration: %s", *configFile)

	config, err := conf.NewConfiguration(*configFile)
	if err != nil {
		log.Fatalf("ERROR: Could not load config file '%s': %v", *configFile, err)
	}
	log.Println("Config loaded!")

	shutdownChannel := make(chan bool, 1)
	util.RegisterShutdownSignalCallback(shutdownChannel)

	maxRounds, err := applyExitConditions(config, *requestedDurationString,
		shutdownChannel)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Driver:", config.Driver)
	log.Println("Recreate connections:", config.Comparison.RecreateConnections)
	log.Println("Backends:", len(config.Backends))
	log.Println("Threads:", config.Comparison.Threads)
	log.Println("Rounds:", config.Comparison.Rounds)
	log.Println("Duration:", *requestedDurationString)

	baselineBackendName := config.Comparison.BaselineBackend
	backendNames, backendInstances, err := createBackendTesters(config, baselineBackendName)
	if err != nil {
		log.Fatalln(err)
	}

	testManager := TestManager{
		BackendNames:               backendNames,
		BackendInstances:           backendInstances,
		BaselineBackendName:        baselineBackendName,
		IsShuttingDown:             false,
		LatestBaselineTestDuration: timing.ZeroDuration,
		MaxRounds:                  maxRounds,
		RoundsCompleted:            map[string]int{},
		RoundsCompletedLock:        sync.RWMutex{},
		RoundsMaxReachedWaitGroup:  sync.WaitGroup{},
		ThreadRunnersWaitGroup:     sync.WaitGroup{},
		ShutdownChannel:            shutdownChannel,
		Threads:                    config.Comparison.Threads,
	}

	log.Println("Running tests...")

	aggregatedTimings, err := runMainTestingLoop(config, &testManager)
	if err != nil {
		log.Fatalln(err)
	}

	err = processAllResults(backendNames, config.Formatters, aggregatedTimings)
	if err != nil {
		log.Fatalln(err)
	}

	if *continueRunningAfterExit == true {
		log.Println("Continuing to run after tests requested. Sleeping forever...")
		log.Println("You can exit this process by sending it SIGTERM/SIGKILL/SIGINT")
		signal.Reset()
		select {}
	}
}
