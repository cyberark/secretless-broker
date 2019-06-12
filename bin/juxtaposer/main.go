package main

import (
	"flag"
	"fmt"
	"log"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"time"

	conf "github.com/cyberark/secretless-broker/bin/juxtaposer/config"
	"github.com/cyberark/secretless-broker/bin/juxtaposer/formatter"
	tester_api "github.com/cyberark/secretless-broker/bin/juxtaposer/tester/api"
	"github.com/cyberark/secretless-broker/bin/juxtaposer/tester/db"
	"github.com/cyberark/secretless-broker/bin/juxtaposer/timing"
	"github.com/cyberark/secretless-broker/bin/juxtaposer/util"
)

func performInvocation(backendName string, backendTestManager tester_api.DriverManager,
	backendConfig conf.Backend) (time.Duration, error) {

	if backendConfig.Debug {
		fmt.Printf("%s %s %s\n",
			strings.Repeat("v", 35),
			backendName,
			strings.Repeat("v", 35))
	}

	testDuration, err := backendTestManager.RunSingleTest()
	if err != nil {
		return timing.ZeroDuration, err
	}

	if backendConfig.Debug {
		log.Println("Run completed")
		fmt.Printf("%s\n", strings.Repeat("^", 85))
	}

	return testDuration, nil
}

func createBackendTesters(config *conf.Config,
	baselineBackendName string) ([]string, map[string]tester_api.DriverManager, error) {

	// Keys in a map are not guaranteed to be retrieved in the same order
	// each time so we have a separate array that guarantees it
	backendNames := []string{}
	backendInstances := map[string]tester_api.DriverManager{}

	log.Println("Backends:", len(config.Backends))
	for backendName, backendConfig := range config.Backends {
		backendNames = append(backendNames, backendName)

		log.Printf("Setting up backend: %s", backendName)

		// Sanity check
		if !strings.HasPrefix(config.Comparison.Type, "sql/") {
			return nil, nil, fmt.Errorf("ERROR: Comparison type not supported: %s", config.Comparison.Type)
		}

		// TODO: Make this more robust
		connectionType := config.Comparison.Type[4:]

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

		backendTestManager, err := db.NewTestDriver(config.Driver, config.Comparison.Style, options)

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

func processAllResults(backendNames []string, config *conf.Config,
	aggregatedTimings map[string]timing.BackendTiming) error {

	for formatterName, formatterOptions := range config.Formatters {
		log.Printf("Processing output formatter '%s'...", formatterName)

		formatterType := formatterOptions["type"]
		if formatterType == "" {
			formatterType = formatterName
		}

		formatter, err := formatter.GetFormatter(formatterType, formatterOptions)
		if err != nil {
			return err
		}

		formatter.ProcessResults(backendNames, aggregatedTimings,
			config.Comparison.BaselineMaxThresholdPercent)
	}

	return nil
}

func runMainTestingLoop(config *conf.Config, backendNames *[]string,
	backendInstances map[string]tester_api.DriverManager,
	baselineBackendName string,
	rounds int,
	shutdownChannel <-chan bool) (map[string]timing.BackendTiming, error) {

	aggregateTimings := timing.NewAggregateTimings(backendNames, baselineBackendName,
		config.Comparison.Silent)

	round := 0
	shuttingDown := false

	var baselineTestDuration time.Duration
	for {
		select {
		case _ = <-shutdownChannel:
			shuttingDown = true
		default:
		}

		if shuttingDown {
			break
		}

		round++
		if rounds != -1 && round > rounds {
			break
		}

		for _, backendName := range *backendNames {
			singleTestRunDuration, testErr := performInvocation(backendName, backendInstances[backendName],
				config.Backends[backendName])

			if backendName == baselineBackendName {
				baselineTestDuration = singleTestRunDuration
			}

			aggregateTimings.AddTimingData(&timing.SingleRunTiming{
				BaselineTestDuration: baselineTestDuration,
				BackendName:          backendName,
				Duration:             singleTestRunDuration,
				MaxRounds:            rounds,
				Round:                round,
				TestError:            testErr,
			})
		}

	}

	aggregateTimings.Process()

	return aggregateTimings.Timings, nil
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

	log.Println("Driver:", config.Driver)
	log.Println("Comparison type:", config.Comparison.Type)

	baselineBackendName := config.Comparison.BaselineBackend
	backendNames, backendInstances, err := createBackendTesters(config, baselineBackendName)
	if err != nil {
		log.Fatalln(err)
	}

	shutdownChannel := make(chan bool, 1)
	util.RegisterShutdownSignalCallback(shutdownChannel)

	rounds, err := applyExitConditions(config, *requestedDurationString, shutdownChannel)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Running tests...")
	aggregatedTimings, err := runMainTestingLoop(config,
		&backendNames,
		backendInstances,
		baselineBackendName,
		rounds,
		shutdownChannel)
	if err != nil {
		log.Fatalln(err)
	}

	err = processAllResults(backendNames, config, aggregatedTimings)
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
