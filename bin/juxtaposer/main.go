package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/cyberark/secretless-broker/bin/juxtaposer/formatter"
	formatter_api "github.com/cyberark/secretless-broker/bin/juxtaposer/formatter/api"
	tester_api "github.com/cyberark/secretless-broker/bin/juxtaposer/tester/api"
	"github.com/cyberark/secretless-broker/bin/juxtaposer/tester/db"
	"github.com/cyberark/secretless-broker/bin/juxtaposer/util"
)

type backend struct {
	Database    string `yaml:"database"`
	Debug       bool   `yaml:"debug"`
	Description string `yaml:"description"`
	Host        string `yaml:"host"`
	Ignore      bool   `yaml:"ignore"`
	Password    string `yaml:"password"`
	Port        string `yaml:"port"`
	SslMode     string `yaml:"sslmode"`
	Socket      string `yaml:"socket"`
	Username    string `yaml:"username"`
}

type comparison struct {
	BaselineBackend             string `yaml:"baselineBackend"`
	BaselineMaxThresholdPercent int    `yaml:"baselineMaxThresholdPercent"`
	Rounds                      string `yaml:"rounds"`
	Silent                      bool   `yaml:"silent"`
	Style                       string `yaml:"style"`
	Type                        string `yaml:"type"`
}

// Config is the main structure used to define the perfagent parameters
type Config struct {
	Backends   map[string]backend                        `yaml:"backends"`
	Comparison comparison                                `yaml:"comparison"`
	Driver     string                                    `yaml:"driver"`
	Formatters map[string]formatter_api.FormatterOptions `yaml:"formatters"`
}

const zeroDuration = 0 * time.Second

func verifyConfiguration(config *Config) error {
	if config.Comparison.Type != "sql" {
		return fmt.Errorf("ERROR: Comparison type supported: %s", config.Comparison.Type)
	}

	if config.Comparison.Style != "select" {
		return fmt.Errorf("ERROR: Comparison style supported: %s", config.Comparison.Style)
	}

	if len(config.Formatters) == 0 {
		return fmt.Errorf("ERROR: No formatters defined")
	}

	baselineBackend := config.Comparison.BaselineBackend
	if baselineBackend == "" {
		return fmt.Errorf("ERROR: Comparison baselineBackend must be specified")
	}

	if _, ok := config.Backends[baselineBackend]; !ok {
		return fmt.Errorf("ERROR: Comparison baseline backend '%s' not found",
			baselineBackend)
	}

	return nil
}

func readConfiguration(configFile string) (*Config, error) {
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	// Default options
	config := Config{
		Comparison: comparison{
			BaselineMaxThresholdPercent: 120,
			Rounds:                      "1000",
			Style:                       "select",
			Type:                        "sql",
		},
		Formatters: map[string]formatter_api.FormatterOptions{
			"stdout": formatter_api.FormatterOptions{},
		},
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}

	// Slice out any backends which are ignored
	filteredBackends := map[string]backend{}
	for backendName, backendConfig := range config.Backends {
		if backendConfig.Ignore == false {
			filteredBackends[backendName] = backendConfig
		}
	}

	config.Backends = filteredBackends

	return &config, nil
}

func performInvocation(backendName string, backendTestManager tester_api.DriverManager,
	backendConfig backend) (time.Duration, error) {

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

func createBackendTesters(config *Config,
	baselineBackendName string) ([]string, map[string]tester_api.DriverManager, error) {

	// Keys in a map are not guaranteed to be retrieved in the same order
	// each time so we have a separate array that guarantees it
	backendNames := []string{}
	backendInstances := map[string]tester_api.DriverManager{}

	log.Println("Backends:", len(config.Backends))
	for backendName, backendConfig := range config.Backends {
		backendNames = append(backendNames, backendName)

		log.Printf("Setting up backend: %s", backendName)

		if config.Comparison.Type != "sql" {
			return nil, nil, fmt.Errorf("ERROR: Comparison type supported: %s", config.Comparison.Type)
		}

		options := tester_api.DbTesterOptions{
			DatabaseName: backendConfig.Database,
			Debug:        backendConfig.Debug,
			Host:         backendConfig.Host,
			Password:     backendConfig.Password,
			Port:         backendConfig.Port,
			SslMode:      backendConfig.SslMode,
			Socket:       backendConfig.Socket,
			Username:     backendConfig.Username,
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

func applyExitConditions(config *Config, requestedDurationString string,
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

func processAllResults(backendNames []string, config *Config,
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

func updateTimingData(round int, config *Config, aggregatedTimings map[string]formatter_api.BackendTiming,
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

func runMainTestingLoop(config *Config, backendNames *[]string,
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

	var err error
	config, err := readConfiguration(*configFile)
	if err != nil {
		log.Fatalf("ERROR: Could not read config file '%s': %v", *configFile, err)
	}

	err = verifyConfiguration(config)
	if err != nil {
		log.Fatalln(err)
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
