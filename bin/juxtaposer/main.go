package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/cyberark/secretless-broker/bin/juxtaposer/formatter"
	formatter_api "github.com/cyberark/secretless-broker/bin/juxtaposer/formatter/api"
	tester_api "github.com/cyberark/secretless-broker/bin/juxtaposer/tester/api"
	"github.com/cyberark/secretless-broker/bin/juxtaposer/tester/db"
)

type Backend struct {
	Database    string `yaml:"database"`
	Debug       bool   `yaml:"debug"`
	Description string `yaml:"description"`
	Host        string `yaml:"host"`
	Ignore      bool   `yaml:"ignore"`
	Password    string `yaml:"password"`
	Port        string `yaml:"port"`
	SslMode     string `yaml:"sslmode"`
	Socket      string `yaml:"socket"`
	Username    string `yaml:"username`
}

type Comparison struct {
	BaselineBackend             string `yaml:"baselineBackend"`
	BaselineMaxThresholdPercent int    `yaml:"baselineMaxThresholdPercent"`
	Rounds                      string `yaml:"rounds"`
	Style                       string `yaml:"style"`
	Type                        string `yaml:"type"`
}

type Config struct {
	Backends   map[string]Backend                        `yaml:"backends"`
	Comparison Comparison                                `yaml:"comparison"`
	Driver     string                                    `yaml:"driver"`
	Formatters map[string]formatter_api.FormatterOptions `yaml:"formatters"`
}

const ZeroDuration = 0 * time.Second

func verifyConfiguration(config *Config) error {
	if config.Comparison.Type != "sql" {
		return fmt.Errorf("ERROR: Comparison type supported: %s", config.Comparison.Type)
	}

	if config.Comparison.Style != "select" {
		return fmt.Errorf("ERROR: Comparison style supported: %s", config.Comparison.Style)
	}

	if len(config.Formatters) == 0 {
		return fmt.Errorf("ERROR: No formatters defined!")
	}

	baselineBackend := config.Comparison.BaselineBackend
	if baselineBackend == "" {
		return fmt.Errorf("ERROR: Comparison baselineBackend must be specified!")
	}

	if _, ok := config.Backends[baselineBackend]; !ok {
		return fmt.Errorf("ERROR: Comparison baseline backend '%s' not found!",
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
		Comparison: Comparison{
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
	filteredBackends := map[string]Backend{}
	for backendName, backendConfig := range config.Backends {
		if backendConfig.Ignore == false {
			filteredBackends[backendName] = backendConfig
		}
	}

	config.Backends = filteredBackends

	return &config, nil
}

func registerShutdownSignalHandlers(shutdownChannel chan<- bool) {
	log.Println("Registering shutdown signal listeners...")
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGABRT,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	go func() {
		exitSignal := <-signalChannel
		log.Printf("Intercepted exit signal '%v'. Waiting for loop to finish...", exitSignal)
		shutdownChannel <- true
	}()
}

func performInvocation(backendName string, backendTestManager tester_api.DriverManager,
	backendConfig Backend) (time.Duration, error) {

	if backendConfig.Debug {
		fmt.Printf("%s %s %s\n",
			strings.Repeat("v", 35),
			backendName,
			strings.Repeat("v", 35))
	}

	testDuration, err := backendTestManager.RunSingleTest()
	if err != nil {
		return ZeroDuration, err
	}

	if backendConfig.Debug {
		log.Println("Run completed")
		fmt.Printf("%s\n", strings.Repeat("^", 85))
	}

	return testDuration, nil
}

func main() {
	log.Println("Juxtaposer starting up...")

	configFile := flag.String("f", "juxtaposer.yml", "Location of the configuration file.")
	flag.Parse()

	log.Printf("Using configuration: %s", *configFile)

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

	// Keys in a map are not guaranteed to be retrieved in the same order
	// each time so we have a separate array that guarantees it
	backendNames := []string{}
	backendInstances := map[string]tester_api.DriverManager{}
	baselineBackendName := config.Comparison.BaselineBackend

	log.Println("Backends:", len(config.Backends))
	for backendName, backendConfig := range config.Backends {
		backendNames = append(backendNames, backendName)

		log.Printf("Setting up backend: %s", backendName)

		if config.Comparison.Type != "sql" {
			log.Fatalf("ERROR: Comparison type supported: %s", config.Comparison.Type)
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
			log.Fatalln(err)
		}

		backendInstances[backendName] = backendTestManager
	}

	// Sort backendNames for consistent output
	sort.Strings(backendNames)

	// Place baseline backend first
	backendBaselineNameIndex := sort.SearchStrings(backendNames, baselineBackendName)
	backendNames = append(backendNames[:backendBaselineNameIndex], backendNames[backendBaselineNameIndex+1:]...)
	backendNames = append([]string{baselineBackendName}, backendNames...)

	aggregatedTimings := map[string]formatter_api.BackendTiming{}
	for _, backendName := range backendNames {
		aggregatedTimings[backendName] = formatter_api.BackendTiming{
			BaselineDivergencePercent: map[int]int{},
			Count:                     0,
			Duration:                  ZeroDuration,
			MinimumDuration:           ZeroDuration,
			MaximumDuration:           ZeroDuration,
			Errors:                    []formatter_api.TestRunError{},
		}
	}

	log.Println("-------------------------")
	log.Println("Starting juxtaposition...")
	log.Println("-------------------------")

	shutdownChannel := make(chan bool, 1)
	registerShutdownSignalHandlers(shutdownChannel)

	rounds := -1
	if config.Comparison.Rounds != "infinity" {
		rounds, err = strconv.Atoi(config.Comparison.Rounds)
		if err != nil {
			log.Fatalln(err)
		}
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

		round = round + 1
		if rounds != -1 && round > rounds {
			break
		}

		for _, backendName := range backendNames {
			singleTestRunDuration, err := performInvocation(backendName, backendInstances[backendName],
				config.Backends[backendName])

			timingInfo := aggregatedTimings[backendName]
			timingInfo.Count = timingInfo.Count + 1
			if err != nil {
				log.Printf("[%.3d/%s] %-20s=> %v", round, config.Comparison.Rounds, backendName, err)
				timingInfo.Errors = append(timingInfo.Errors,
					formatter_api.TestRunError{
						Error: err,
						Round: round,
					})
				aggregatedTimings[backendName] = timingInfo
				continue
			}

			timingInfo.Duration = timingInfo.Duration + singleTestRunDuration

			if backendName == baselineBackendName {
				baselineTestDuration = singleTestRunDuration
			}

			if timingInfo.MinimumDuration == ZeroDuration {
				timingInfo.MinimumDuration = timingInfo.Duration
			}

			if singleTestRunDuration > timingInfo.MaximumDuration {
				timingInfo.MaximumDuration = singleTestRunDuration
			}

			if singleTestRunDuration < timingInfo.MinimumDuration {
				timingInfo.MinimumDuration = singleTestRunDuration
			}

			baselineDivergencePercent := 100
			if backendName != baselineBackendName {
				baselineDivergencePercent = int(float32(singleTestRunDuration) /
					float32(baselineTestDuration) * 100.0)
			}

			log.Printf("[%d/%s], %-20s=>%15v, %3d%%", round, config.Comparison.Rounds,
				backendName, singleTestRunDuration, baselineDivergencePercent)

			timingInfo.BaselineDivergencePercent[baselineDivergencePercent] += 1

			aggregatedTimings[backendName] = timingInfo
		}
	}

	for formatterName, formatterOptions := range config.Formatters {
		log.Printf("Processing output formatter '%s'...", formatterName)

		formatterType := formatterOptions["type"]
		if formatterType == "" {
			formatterType = formatterName
		}

		formatter, err := formatter.GetFormatter(formatterType, formatterOptions)
		if err != nil {
			log.Fatalln(err)
		}

		formatter.ProcessResults(backendNames, aggregatedTimings,
			config.Comparison.BaselineMaxThresholdPercent)
	}
}
