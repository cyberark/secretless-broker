package db

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/cyberark/secretless-broker/bin/juxtaposer/tester/api"
	mysql "github.com/cyberark/secretless-broker/bin/juxtaposer/tester/db/mysql"
	postgres "github.com/cyberark/secretless-broker/bin/juxtaposer/tester/db/postgres"
)

type DriverManager struct {
	Options  *api.DbTesterOptions
	Tester   api.DbTester
	TestType string
}

var DbTesterImplementatons = map[string]func() (api.DbTester, error){
	"mysql-5.7": mysql.NewTester,
	"postgres":  postgres.NewTester,
}

const ZeroDuration = 0 * time.Second

const DefaultDatabaseName = "mydb"
const DefaultTableName = "mytable"

const SampleDataRowCount = 100
const NameFieldPrefix = "person #"

const CreateTableStatement = `
    name         TEXT,
    id           INTEGER,
    birth_date   DATE,
    result       DECIMAL,
    passed       BOOLEAN
`

var QueryTypes = map[string]string{
	"dropTable": fmt.Sprintf("DROP TABLE IF EXISTS %s;", DefaultTableName),
	"createTable": fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);",
		DefaultTableName,
		CreateTableStatement),
	"insertItem": fmt.Sprintf(`INSERT INTO %s (name, id, birth_date, result, passed)
		VALUES `, DefaultTableName),
	"select": fmt.Sprintf("SELECT name FROM %s;", DefaultTableName),
}

func (manager *DriverManager) ensureWantedDbDataState() error {
	err := manager.Tester.Connect(*manager.Options)
	if err != nil {
		log.Printf("ERROR! Connect failed!")
		return err
	}
	defer manager.Tester.Shutdown()

	err = manager.Tester.Query(QueryTypes["dropTable"])
	if err != nil {
		return err
	}

	err = manager.Tester.Query(QueryTypes["createTable"])
	if err != nil {
		return err
	}

	if manager.Options.Debug {
		log.Printf("Table created")
	}

	for itemIndex := 0; itemIndex < SampleDataRowCount; itemIndex++ {
		insertItemStatement := QueryTypes["insertItem"] +
			fmt.Sprintf("(%s)", manager.Tester.GetQueryMarkers(5))

		err = manager.Tester.Query(insertItemStatement,
			fmt.Sprintf("%s%d", NameFieldPrefix, itemIndex),
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
		return fmt.Errorf("ERROR: DB driver type not supported: %s!", driverName)
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

func validateOptions(options api.DbTesterOptions) error {
	if options.Host == "" && options.Socket == "" {
		return fmt.Errorf("ERROR: Neither Host nor Socket specified!")
	}

	if options.DatabaseName == "" {
		return fmt.Errorf("ERROR: Database name not specified!")
	}

	return nil
}

func ensureCorrectReturnedData(rows []string) error {
	if len(rows) != SampleDataRowCount {
		return fmt.Errorf("ERROR: Expected %d returned rows but got %d",
			SampleDataRowCount,
			len(rows))
	}

	for _, row := range rows {
		if row[:len(NameFieldPrefix)] != NameFieldPrefix {
			return fmt.Errorf("ERROR: Row '%s' did not have expected prefix '%s'",
				row,
				NameFieldPrefix)
		}
	}

	return nil
}

func (manager *DriverManager) RunSingleTest() (time.Duration, error) {
	startTime := time.Now()

	rows, err := manager.Tester.QueryRows("name", QueryTypes[manager.TestType])
	if err != nil {
		log.Printf("ERROR! Query failed!")
		return ZeroDuration, err
	}

	err = ensureCorrectReturnedData(rows)
	if err != nil {
		return ZeroDuration, err
	}

	testDuration := time.Now().Sub(startTime)

	if manager.Options.Debug {
		log.Printf("WARN: TODO: Compare returned data!")
		log.Printf("DB query: OK")
	}

	return testDuration, nil
}

func (manager *DriverManager) RotatePassword(newPassword string) error {
	return fmt.Errorf("ERROR: Rotating passwords is not yet implemented!")
}

func (manager *DriverManager) Shutdown() error {
	return manager.Tester.Shutdown()
}

func NewTestDriver(driver string, testType string, options api.DbTesterOptions) (api.DriverManager, error) {
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
