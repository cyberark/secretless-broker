package db

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/cyberark/secretless-broker/bin/juxtaposer/testers/api"
)

type DbTesterOptions struct {
	Address      string
	DatabaseName string
	Debug        bool
	Password     string
	Socket       string
	Username     string
}

type DbTester interface {
	Connect(DbTesterOptions) error
	Query(string, ...interface{}) ([]byte, error)
	Shutdown() error
}

type DriverManager struct {
	Options  *DbTesterOptions
	Tester   DbTester
	TestType string
}

var DbTesterImplementatons = map[string]func() (DbTester, error){
	"mysql-5.7": NewMysqlTester,
}

const ZeroDuration = 0 * time.Second
const DefaultDatabaseName = "mydb"
const DefaultTableName = "mytable"
const CreateTableStatement = `
    name         TEXT,
    id           INTEGER,
    birth_date   DATE,
    result       DECIMAL,
    passed       BOOLEAN
`

var QueryTypes = map[string]string{
	"dropTable": fmt.Sprintf("DROP TABLE %s;", DefaultTableName),
	"createTable": fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);",
		DefaultTableName,
		CreateTableStatement),
	"insertItem": fmt.Sprintf(`INSERT INTO %s (name, id, birth_date, result, passed)
		VALUES (?, ?, ?, ?, ?);`, DefaultTableName),
	"select": fmt.Sprintf("SELECT * FROM %s;", DefaultTableName),
}

func (manager *DriverManager) ensureWantedDbDataState() error {
	err := manager.Tester.Connect(*manager.Options)
	if err != nil {
		log.Printf("ERROR! Connect failed!")
		return err
	}
	defer manager.Tester.Shutdown()

	_, err = manager.Tester.Query(QueryTypes["dropTable"])
	if err != nil {
		return err
	}

	_, err = manager.Tester.Query(QueryTypes["createTable"])
	if err != nil {
		return err
	}

	if manager.Options.Debug {
		log.Printf("Table created")
	}

	for itemIndex := 0; itemIndex < 10; itemIndex++ {
		_, err = manager.Tester.Query(QueryTypes["insertItem"],
			fmt.Sprintf("person #%d", itemIndex),
			itemIndex,
			time.Now().AddDate(0, 0, itemIndex),
			float32(itemIndex)*10,
			rand.Int31()&(1<<30) == 0)

		if err != nil {
			log.Printf("ERROR! Could not insert canned values into DB!")
			manager.Tester.Shutdown()
			return err
		}
	}

	return nil
}

func (manager *DriverManager) instantiateNewDbDriverTester(driverName string) error {
	testerConstructor, ok := DbTesterImplementatons[driverName]
	if !ok {
		return fmt.Errorf("ERROR: DB tester type not supported: %s!", driverName)
	}

	tester, err := testerConstructor()
	if err != nil {
		return err
	}
	manager.Tester = tester

	err = manager.ensureWantedDbDataState()
	if err != nil {
		return err
	}

	return nil
}

func validateOptions(options DbTesterOptions) error {
	if options.Address == "" && options.Socket == "" {
		return fmt.Errorf("ERROR: Neither Address nor Socket specified!")
	}

	if options.DatabaseName == "" {
		return fmt.Errorf("ERROR: Database name not specified!")
	}

	return nil
}

func (manager *DriverManager) RunSingleTest() (time.Duration, error) {
	startTime := time.Now()

	_, err := manager.Tester.Query(QueryTypes[manager.TestType])
	if err != nil {
		log.Printf("ERROR! Query failed!")
		return ZeroDuration, err
	}

	testDuration := time.Now().Sub(startTime)

	if manager.Options.Debug {
		log.Printf("WARN: TODO: Compare returned data!")
		log.Printf("DB query: OK")
	}

	return testDuration, nil
}

func (manager *DriverManager) Shutdown() error {
	return manager.Tester.Shutdown()
}

func NewTestDriver(driver string, testType string, options DbTesterOptions) (api.DriverManager, error) {
	if options.DatabaseName == "" {
		options.DatabaseName = DefaultDatabaseName
	}

	err := validateOptions(options)
	if err != nil {
		return nil, err
	}

	manager := DriverManager{
		Options:  &options,
		TestType: testType,
	}

	err = manager.instantiateNewDbDriverTester(driver)
	if err != nil {
		return nil, err
	}

	if options.Debug {
		log.Printf("Tester creation: OK")
	}

	manager.Tester.Connect(*manager.Options)
	if options.Debug {
		log.Printf("Tester connection: OK")
	}

	return &manager, nil
}
