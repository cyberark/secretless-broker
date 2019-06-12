package api

import (
	"time"
)

type DriverManager interface {
	GetName() string
	RunSingleTest() (time.Duration, error)
	RotatePassword(string) error
	Shutdown() error
}
