package util

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var listenerPort = strconv.Itoa(defaultHealthCheckPort)

var ReadyEndpoint = "http://localhost:" + listenerPort + "/ready"
var LiveEndpoint = "http://localhost:" + listenerPort + "/live"

var VerboseReadyEndpoint = "http://localhost:" + listenerPort + "/ready?full=1"
var VerboseLiveEndpoint = "http://localhost:" + listenerPort + "/live?full=1"

type HealthJSON struct {
	Ready     string `json:"ready,omitempty"`
	Listening string `json:"listening,omitempty"`
}

func getHealth(endpoint string) (*HealthJSON, error) {
	webClient := http.Client{
		Timeout: time.Second * 2,
	}

	request, requestErr := http.NewRequest(http.MethodGet, endpoint, nil)
	if requestErr != nil {
		return nil, requestErr
	}

	response, responseErr := webClient.Do(request)
	if responseErr != nil {
		return nil, responseErr
	}

	body, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return nil, readErr
	}

	healthJSON := &HealthJSON{}
	unmarshalErr := json.Unmarshal(body, healthJSON)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		err := fmt.Errorf("response Status: %d: %s",
			response.StatusCode,
			http.StatusText(response.StatusCode))
		return healthJSON, err
	}

	fmt.Println(healthJSON)

	return healthJSON, nil
}

func assertHealthStatusCodeIsBad(endpoint string, t *testing.T) {
	_, err := getHealth(endpoint)

	assert.Error(t, err)
	errorMsg := "response Status: 503: Service Unavailable"
	assert.EqualError(t, err, errorMsg)
}

func assertHealthStatusCodeIsGood(endpoint string, t *testing.T) {
	_, err := getHealth(endpoint)

	assert.NoError(t, err)
}

func assertReadyJSONIsBad(err error, healthJSON *HealthJSON, t *testing.T) {
	assert.Error(t, err)
	errorMsg := "response Status: 503: Service Unavailable"
	assert.EqualError(t, err, errorMsg)

	assert.NotNil(t, healthJSON)

	assert.Equal(t, "secretless is not ready", (*healthJSON).Ready)
}

func assertReadyJSONIsGood(err error, healthJSON *HealthJSON, t *testing.T) {
	assert.NoError(t, err)
	assert.NotNil(t, healthJSON)
	assert.Equal(t, "OK", (*healthJSON).Ready)
}

func assertReadyJSONIsNotPresent(healthJSON *HealthJSON, t *testing.T) {
	assert.NotNil(t, healthJSON)
	assert.Equal(t, "", (*healthJSON).Ready)
}

func assertListeningJSONIsBad(err error, healthJSON *HealthJSON, t *testing.T) {
	assert.Error(t, err)
	errorMsg := "response Status: 503: Service Unavailable"
	assert.EqualError(t, err, errorMsg)

	assert.NotNil(t, healthJSON)

	assert.Equal(t, "secretless is not listening", (*healthJSON).Listening)
}

func assertListeningJSONIsGood(err error, healthJSON *HealthJSON, t *testing.T) {
	assert.NoError(t, err)
	assert.NotNil(t, healthJSON)
	assert.Equal(t, "OK", (*healthJSON).Listening)
}

func callEnableHealthCheck() {
	enableHealthCheck()

	// Server can be slow to come up :(
	time.Sleep(250 * time.Millisecond)
}

func enableAndReadyHealthCheck() {
	callEnableHealthCheck()
	SetAppInitializedFlag()
}

func enableReadyAndLivenHealthCheck() {
	enableAndReadyHealthCheck()
	SetAppIsLive(true)
}

func Test_Health(t *testing.T) {
	t.Run("When nothing is proactively done", func(t *testing.T) {
		t.Run("Shows not ready", func(t *testing.T) {
			callEnableHealthCheck()
			assertHealthStatusCodeIsBad(ReadyEndpoint, t)
		})

		t.Run("Shows not live", func(t *testing.T) {
			callEnableHealthCheck()
			assertHealthStatusCodeIsBad(LiveEndpoint, t)
		})

		t.Run("Shows expected not ready JSON", func(t *testing.T) {
			callEnableHealthCheck()

			health, err := getHealth(VerboseReadyEndpoint)
			assertReadyJSONIsBad(err, health, t)
			assertListeningJSONIsBad(err, health, t)
		})

		t.Run("Shows expected not live JSON", func(t *testing.T) {
			callEnableHealthCheck()

			health, err := getHealth(VerboseLiveEndpoint)
			assertReadyJSONIsNotPresent(health, t)
			assertListeningJSONIsBad(err, health, t)
		})

		t.Cleanup(func() {
			disableHealthCheck()
		})
	})

	t.Run("When readniness flag is turned on but liveliness is not", func(t *testing.T) {
		t.Run("Shows not ready", func(t *testing.T) {
			enableAndReadyHealthCheck()
			assertHealthStatusCodeIsBad(ReadyEndpoint, t)
		})

		t.Run("Shows not live", func(t *testing.T) {
			enableAndReadyHealthCheck()
			assertHealthStatusCodeIsBad(LiveEndpoint, t)
		})

		t.Run("Shows expected ready JSON but bad listening status", func(t *testing.T) {
			enableAndReadyHealthCheck()

			health, err := getHealth(VerboseReadyEndpoint)

			// We ignore error manually since we know it will be checked below
			assertReadyJSONIsGood(nil, health, t)

			assertListeningJSONIsBad(err, health, t)
		})

		t.Run("Shows expected not live JSON", func(t *testing.T) {
			enableAndReadyHealthCheck()

			health, err := getHealth(VerboseLiveEndpoint)
			assertReadyJSONIsNotPresent(health, t)
			assertListeningJSONIsBad(err, health, t)
		})

		t.Cleanup(func() {
			disableHealthCheck()
		})
	})

	t.Run("When app is ready and listening", func(t *testing.T) {
		t.Run("Shows ready", func(t *testing.T) {
			enableReadyAndLivenHealthCheck()
			assertHealthStatusCodeIsGood(ReadyEndpoint, t)
		})

		t.Run("Shows live", func(t *testing.T) {
			enableReadyAndLivenHealthCheck()
			assertHealthStatusCodeIsGood(LiveEndpoint, t)
		})

		t.Run("Shows expected ready JSON", func(t *testing.T) {
			enableReadyAndLivenHealthCheck()

			health, err := getHealth(VerboseReadyEndpoint)

			assertReadyJSONIsGood(err, health, t)
			assertListeningJSONIsGood(err, health, t)
		})

		t.Run("Shows expected live JSON", func(t *testing.T) {
			enableReadyAndLivenHealthCheck()

			health, err := getHealth(VerboseLiveEndpoint)
			assertReadyJSONIsNotPresent(health, t)
			assertListeningJSONIsGood(err, health, t)
		})

		t.Cleanup(func() {
			disableHealthCheck()
		})
	})

	t.Run("When app is ready and listening is dynamically changed", func(t *testing.T) {
		t.Run("Updates ready status", func(t *testing.T) {
			enableReadyAndLivenHealthCheck()
			assertHealthStatusCodeIsGood(ReadyEndpoint, t)

			SetAppIsLive(false)
			assertHealthStatusCodeIsBad(ReadyEndpoint, t)

			SetAppIsLive(true)
			assertHealthStatusCodeIsGood(ReadyEndpoint, t)
		})

		t.Run("Updates live status", func(t *testing.T) {
			enableReadyAndLivenHealthCheck()
			assertHealthStatusCodeIsGood(LiveEndpoint, t)

			SetAppIsLive(false)
			assertHealthStatusCodeIsBad(LiveEndpoint, t)

			SetAppIsLive(true)
			assertHealthStatusCodeIsGood(LiveEndpoint, t)
		})

		t.Run("Updates expected ready JSON", func(t *testing.T) {
			enableReadyAndLivenHealthCheck()

			health, err := getHealth(VerboseReadyEndpoint)
			assertReadyJSONIsGood(err, health, t)
			assertListeningJSONIsGood(err, health, t)

			SetAppIsLive(false)
			health, err = getHealth(VerboseReadyEndpoint)
			assertReadyJSONIsGood(nil, health, t)
			assertListeningJSONIsBad(err, health, t)

			SetAppIsLive(true)
			health, err = getHealth(VerboseReadyEndpoint)
			assertReadyJSONIsGood(err, health, t)
			assertListeningJSONIsGood(err, health, t)
		})

		t.Run("Updates expected live JSON", func(t *testing.T) {
			enableReadyAndLivenHealthCheck()

			health, err := getHealth(VerboseLiveEndpoint)
			assertReadyJSONIsNotPresent(health, t)
			assertListeningJSONIsGood(err, health, t)

			SetAppIsLive(false)
			health, err = getHealth(VerboseLiveEndpoint)
			assertReadyJSONIsNotPresent(health, t)
			assertListeningJSONIsBad(err, health, t)

			SetAppIsLive(true)
			health, err = getHealth(VerboseLiveEndpoint)
			assertReadyJSONIsNotPresent(health, t)
			assertListeningJSONIsGood(err, health, t)
		})

		t.Cleanup(func() {
			disableHealthCheck()
		})
	})
}
