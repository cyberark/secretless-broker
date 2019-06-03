package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/cyberark/secretless-broker/bin/juxtaposer/testers/api"
	"github.com/cyberark/secretless-broker/bin/juxtaposer/testers/db"
)

type Backend struct {
	Address     string `yaml:"address"`
	Debug       bool   `yaml:"debug"`
	Description string `yaml:"description"`
	Password    string `yaml:"password"`
	Socket      string `yaml:"socket"`
	Username    string `yaml:"username`
}

type BackendTiming struct {
	Count    int
	Duration time.Duration
	Errors   []error
}

type Comparison struct {
	Rounds string `yaml:"rounds"`
	Style  string `yaml:"style"`
	Type   string `yaml:"type"`
}

type Config struct {
	Backends   map[string]Backend `yaml:"backends"`
	Comparison Comparison         `yaml:"comparison"`
	Driver     string             `yaml:"driver"`
}

const ZeroDuration = 0 * time.Second

func verifyConfiguration(config *Config) error {
	if config.Driver != "mysql-5.7" {
		err := fmt.Errorf("ERROR: Driver not supported: %s", config.Driver)
		return err
	}

	if config.Comparison.Type != "sql" {
		err := fmt.Errorf("ERROR: Comparison type supported: %s", config.Comparison.Type)
		return err
	}

	if config.Comparison.Style != "select" {
		err := fmt.Errorf("ERROR: Comparison style supported: %s", config.Comparison.Style)
		return err
	}

	return nil
}

func readConfiguration(configFile string) (*Config, error) {
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	config := Config{
		Comparison: Comparison{
			Rounds: "1000",
			Style:  "select",
			Type:   "sql",
		},
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}

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

func printResults(backendNames []string, aggregatedTimings map[string]BackendTiming) {
	fmt.Println()
	fmt.Println("Calculating averages...")
	fmt.Println()
	fmt.Printf("%-20s|%15s|%8s|%8s|%13s|%15s|\n",
		"Name",
		"Avg Duration",
		"Runs",
		"Errors",
		"Success(%)",
		"Total Duration")

	dividerString := strings.Repeat("-", 85)
	fmt.Printf("%s\n", dividerString)

	for _, backendName := range backendNames {
		timingInfo := aggregatedTimings[backendName]

		successfulRuns := timingInfo.Count - len(timingInfo.Errors)

		averageDuration := ZeroDuration
		if successfulRuns > 0 {
			averageDuration = time.Duration(int64(timingInfo.Duration) /
				int64(successfulRuns))
		}

		fmt.Printf("%-20s %15v %8d %8d %13.0f %15v \n",
			backendName,
			averageDuration,
			timingInfo.Count,
			len(timingInfo.Errors),
			(float32(successfulRuns)/float32(timingInfo.Count))*100,
			timingInfo.Duration)
	}
}

func performInvocation(backendName string, backendTestManager api.DriverManager,
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
		log.Printf("ERROR: Could not read config file: %v", err)
		os.Exit(1)
	}

	err = verifyConfiguration(config)
	if err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}

	log.Println("Config loaded!")

	log.Println("Driver:", config.Driver)
	log.Println("Comparison type:", config.Comparison.Type)

	// Keys in a map are not guaranteed to be retrieved in the same order
	// each time so we have a separate array that guarantees it
	backendNames := []string{}
	backendInstances := map[string]api.DriverManager{}

	log.Println("Backends:", len(config.Backends))
	for backendName, backendConfig := range config.Backends {
		backendNames = append(backendNames, backendName)

		log.Printf("Setting up backend: %s", backendName)

		if config.Comparison.Type != "sql" {
			err := fmt.Errorf("ERROR: Comparison type supported: %s", config.Comparison.Type)
			log.Printf("%v", err)
			os.Exit(1)
		}

		options := db.DbTesterOptions{
			Address:  backendConfig.Address,
			Debug:    backendConfig.Debug,
			Password: backendConfig.Password,
			Socket:   backendConfig.Socket,
			Username: backendConfig.Username,
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
			log.Printf("%v", err)
			os.Exit(1)
		}

		backendInstances[backendName] = backendTestManager
	}

	aggregatedTimings := map[string]BackendTiming{}
	for _, backendName := range backendNames {
		aggregatedTimings[backendName] = BackendTiming{
			Count:    0,
			Duration: 0 * time.Second,
			Errors:   []error{},
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
			log.Printf("%v", err)
			os.Exit(1)
		}
	}

	shuttingDown := false
	round := 0
	//	for {
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
				timingInfo.Errors = append(timingInfo.Errors, err)
				aggregatedTimings[backendName] = timingInfo
				continue
			}

			log.Printf("[%.3d/%s], %-20s=>%15v", round, config.Comparison.Rounds,
				backendName, singleTestRunDuration)
			timingInfo.Duration = timingInfo.Duration + singleTestRunDuration

			aggregatedTimings[backendName] = timingInfo
		}
	}

	printResults(backendNames, aggregatedTimings)
}
