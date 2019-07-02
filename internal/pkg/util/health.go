package util

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/heptiolabs/healthcheck"
)

const defaultHealthCheckPort = 5335

type healthWatcher struct {
	handler        healthcheck.Handler
	isAppReady     bool
	isAppLive      bool
	serverInstance *http.Server
}

var _healthWatcher *healthWatcher
var _healthWatcherSyncOnce = sync.Once{}

func (healthWatcher *healthWatcher) registerLivelinessCheck() {
	checkFunc := func() error {
		if healthWatcher.isAppLive {
			return nil
		}

		// If we are not ready, we return a failed liveliness check
		return fmt.Errorf("secretless is not listening")
	}

	healthWatcher.handler.AddLivenessCheck("listening", checkFunc)
}

func (healthWatcher *healthWatcher) registerReadinessCheck() {
	checkFunc := func() error {
		if healthWatcher.isAppReady {
			return nil
		}

		// If we are not ready, we return a failed readiness check
		return fmt.Errorf("secretless is not ready")
	}

	healthWatcher.handler.AddReadinessCheck("ready", checkFunc)
}

func (healthWatcher *healthWatcher) enable(port int) {
	log.Printf("Initializing health check on :%d...", port)

	healthWatcher.handler = healthcheck.NewHandler()

	healthWatcher.registerLivelinessCheck()
	healthWatcher.registerReadinessCheck()

	healthWatcher.serverInstance = &http.Server{
		Addr:           ":" + strconv.Itoa(port),
		Handler:        healthWatcher.handler,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go healthWatcher.serverInstance.ListenAndServe()

	log.Printf("Initialization of health check done. " +
		"You can access the endpoint at `/live` and `/ready`.")
}

func enableHealthCheck() {
	_healthWatcherSyncOnce.Do(func() {
		_healthWatcher = &healthWatcher{}
		_healthWatcher.enable(defaultHealthCheckPort)
	})
}

func disableHealthCheck() {
	if _healthWatcher == nil || _healthWatcher.serverInstance == nil {
		return
	}

	// Clean up everything as best we can to ensure prompt GC
	if err := _healthWatcher.serverInstance.Shutdown(context.Background()); err != nil {
		panic(err)
	}

	_healthWatcher.serverInstance = nil
	_healthWatcher.handler = nil
	_healthWatcher = nil
	_healthWatcherSyncOnce = sync.Once{}
}

// SetAppInitializedFlag enables health check and sets the ready flag.
func SetAppInitializedFlag() {
	enableHealthCheck()
	_healthWatcher.isAppReady = true
}

// SetAppIsLive enables health check and marks the app as live.
func SetAppIsLive(isLive bool) {
	enableHealthCheck()
	_healthWatcher.isAppLive = isLive
}
