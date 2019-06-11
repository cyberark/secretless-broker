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
	formatter_api "github.com/cyberark/secretless-broker/bin/juxtaposer/formatter/api"
	tester_api "github.com/cyberark/secretless-broker/bin/juxtaposer/tester/api"
	"github.com/cyberark/secretless-broker/bin/juxtaposer/tester/db"
	"github.com/cyberark/secretless-broker/bin/juxtaposer/util"
)

const zeroDuration = 0 * time.Second

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
		return zeroDuration, err
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

	requestedDuration := zeroDuration
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

	if requestedDuration != zeroDuration {
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
	aggregatedTimings map[string]formatter_api.BackendTiming) error {

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

func updateTimingData(round int, config *conf.Config, aggregatedTimings map[string]formatter_api.BackendTiming,
	backendName string, singleTestRunDuration time.Duration, testError error, baselineTestDuration time.Duration) {

	timingInfo := aggregatedTimings[backendName]
	timingInfo.Count++

	if testError != nil {
		log.Printf("[%.3d/%s] %-35s=> %v", round, config.Comparison.Rounds, backendName, testError)
		timingInfo.Errors = append(timingInfo.Errors,
			formatter_api.TestRunError{
				Error: testError,
				Round: round,
			})
		aggregatedTimings[backendName] = timingInfo
		return
	}

	timingInfo.Duration = timingInfo.Duration + singleTestRunDuration

	if timingInfo.MinimumDuration == zeroDuration {
		timingInfo.MinimumDuration = timingInfo.Duration
	}

	if singleTestRunDuration > timingInfo.MaximumDuration {
		timingInfo.MaximumDuration = singleTestRunDuration
	}

	if singleTestRunDuration < timingInfo.MinimumDuration {
		timingInfo.MinimumDuration = singleTestRunDuration
	}

	baselineDivergencePercent := 100
	if baselineTestDuration != zeroDuration {
		baselineDivergencePercent = int(float32(singleTestRunDuration) /
			float32(baselineTestDuration) * 100.0)
	}

	if !config.Comparison.Silent {
		log.Printf("[%d/%s], %-35s=>%15v, %3d%%", round, config.Comparison.Rounds,
			backendName, singleTestRunDuration, baselineDivergencePercent)
	} else {
		if round%1000 == 0 {
			fmt.Printf(".")
		}
	}

	timingInfo.BaselineDivergencePercent[baselineDivergencePercent]++

	aggregatedTimings[backendName] = timingInfo
}

func runMainTestingLoop(config *conf.Config, backendNames *[]string,
	backendInstances map[string]tester_api.DriverManager,
	baselineBackendName string,
	rounds int,
	shutdownChannel <-chan bool) (map[string]formatter_api.BackendTiming, error) {

	aggregatedTimings := map[string]formatter_api.BackendTiming{}
	for _, backendName := range *backendNames {
		aggregatedTimings[backendName] = formatter_api.NewBackendTiming()
	}

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

			updateTimingData(round, config, aggregatedTimings, backendName, singleTestRunDuration,
				testErr, baselineTestDuration)
		}
	}

	return aggregatedTimings, nil
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
