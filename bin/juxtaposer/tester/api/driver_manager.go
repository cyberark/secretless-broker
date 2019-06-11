package api

import (
	"time"
)

type DriverManager interface {
	RunSingleTest() (time.Duration, error)
	RotatePassword(string) error
	Shutdown() error
}
