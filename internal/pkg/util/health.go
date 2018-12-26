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

const DEFAULT_HEALTH_CHECK_PORT = 5335

type HealthWatcher struct {
	handler        healthcheck.Handler
	isAppReady     bool
	isAppLive      bool
	serverInstance *http.Server
}

var _healthWatcher *HealthWatcher
var _healthWatcherSyncOnce = sync.Once{}

func (healthWatcher *HealthWatcher) registerLivelinessCheck() {
	checkFunc := func() error {
		if healthWatcher.isAppLive {
			return nil
		}

		// If we are not ready, we return a failed liveliness check
		return fmt.Errorf("Secretless is not listening")
	}

	healthWatcher.handler.AddLivenessCheck("listening", checkFunc)
}

func (healthWatcher *HealthWatcher) registerReadinessCheck() {
	checkFunc := func() error {
		if healthWatcher.isAppReady {
			return nil
		}

		// If we are not ready, we return a failed readiness check
		return fmt.Errorf("Secretless is not ready")
	}

	healthWatcher.handler.AddReadinessCheck("ready", checkFunc)
}

func (healthWatcher *HealthWatcher) enable(port int) {
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

func EnableHealthCheck() {
	_healthWatcherSyncOnce.Do(func() {
		_healthWatcher = &HealthWatcher{}
		_healthWatcher.enable(DEFAULT_HEALTH_CHECK_PORT)
	})
}

func DisableHealthCheck() {
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

func SetAppInitializedFlag() {
	EnableHealthCheck()
	_healthWatcher.isAppReady = true
}

func SetAppIsLive(isLive bool) {
	EnableHealthCheck()
	_healthWatcher.isAppLive = isLive
}
