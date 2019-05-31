package api

import (
	"time"
)

type DriverManager interface {
	RunSingleTest() (time.Duration, error)
	Shutdown() error
}
