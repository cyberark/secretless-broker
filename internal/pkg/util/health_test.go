package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
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

	body, readErr := ioutil.ReadAll(response.Body)
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

func assertHealthStatusCodeIsBad(endpoint string) {
	_, err := getHealth(endpoint)

	So(err, ShouldNotBeNil)
	errorMsg := "response Status: 503: Service Unavailable"
	So(err.Error(), ShouldEqual, errorMsg)
}

func assertHealthStatusCodeIsGood(endpoint string) {
	_, err := getHealth(endpoint)

	So(err, ShouldBeNil)
}

func assertReadyJSONIsBad(err error, healthJSON *HealthJSON) {
	So(err, ShouldNotBeNil)
	errorMsg := "response Status: 503: Service Unavailable"
	So(err.Error(), ShouldEqual, errorMsg)

	So(healthJSON, ShouldNotBeNil)

	So((*healthJSON).Ready, ShouldEqual, "secretless is not ready")
}

func assertReadyJSONIsGood(err error, healthJSON *HealthJSON) {
	So(err, ShouldBeNil)
	So(healthJSON, ShouldNotBeNil)
	So((*healthJSON).Ready, ShouldEqual, "OK")
}

func assertReadyJSONIsNotPresent(healthJSON *HealthJSON) {
	So(healthJSON, ShouldNotBeNil)
	So((*healthJSON).Ready, ShouldEqual, "")
}

func assertListeningJSONIsBad(err error, healthJSON *HealthJSON) {
	So(err, ShouldNotBeNil)
	errorMsg := "response Status: 503: Service Unavailable"
	So(err.Error(), ShouldEqual, errorMsg)

	So(healthJSON, ShouldNotBeNil)

	So((*healthJSON).Listening, ShouldEqual, "secretless is not listening")
}

func assertListeningJSONIsGood(err error, healthJSON *HealthJSON) {
	So(err, ShouldBeNil)
	So(healthJSON, ShouldNotBeNil)
	So((*healthJSON).Listening, ShouldEqual, "OK")
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
	Convey("Health", t, func() {
		Convey("When nothing is proactively done", func() {
			Convey("Shows not ready", func() {
				callEnableHealthCheck()
				assertHealthStatusCodeIsBad(ReadyEndpoint)
			})

			Convey("Shows not live", func() {
				callEnableHealthCheck()
				assertHealthStatusCodeIsBad(LiveEndpoint)
			})

			Convey("Shows expected not ready JSON", func() {
				callEnableHealthCheck()

				health, err := getHealth(VerboseReadyEndpoint)
				assertReadyJSONIsBad(err, health)
				assertListeningJSONIsBad(err, health)
			})

			Convey("Shows expected not live JSON", func() {
				callEnableHealthCheck()

				health, err := getHealth(VerboseLiveEndpoint)
				assertReadyJSONIsNotPresent(health)
				assertListeningJSONIsBad(err, health)
			})

			Reset(func() {
				disableHealthCheck()
			})
		})

		Convey("When readniness flag is turned on but liveliness is not", func() {
			Convey("Shows not ready", func() {
				enableAndReadyHealthCheck()
				assertHealthStatusCodeIsBad(ReadyEndpoint)
			})

			Convey("Shows not live", func() {
				enableAndReadyHealthCheck()
				assertHealthStatusCodeIsBad(LiveEndpoint)
			})

			Convey("Shows expected ready JSON but bad listening status", func() {
				enableAndReadyHealthCheck()

				health, err := getHealth(VerboseReadyEndpoint)

				// We ignore error manually since we know it will be checked below
				assertReadyJSONIsGood(nil, health)

				assertListeningJSONIsBad(err, health)
			})

			Convey("Shows expected not live JSON", func() {
				enableAndReadyHealthCheck()

				health, err := getHealth(VerboseLiveEndpoint)
				assertReadyJSONIsNotPresent(health)
				assertListeningJSONIsBad(err, health)
			})

			Reset(func() {
				disableHealthCheck()
			})
		})

		Convey("When app is ready and listening", func() {
			Convey("Shows ready", func() {
				enableReadyAndLivenHealthCheck()
				assertHealthStatusCodeIsGood(ReadyEndpoint)
			})

			Convey("Shows live", func() {
				enableReadyAndLivenHealthCheck()
				assertHealthStatusCodeIsGood(LiveEndpoint)
			})

			Convey("Shows expected ready JSON", func() {
				enableReadyAndLivenHealthCheck()

				health, err := getHealth(VerboseReadyEndpoint)

				assertReadyJSONIsGood(err, health)
				assertListeningJSONIsGood(err, health)
			})

			Convey("Shows expected live JSON", func() {
				enableReadyAndLivenHealthCheck()

				health, err := getHealth(VerboseLiveEndpoint)
				assertReadyJSONIsNotPresent(health)
				assertListeningJSONIsGood(err, health)
			})

			Reset(func() {
				disableHealthCheck()
			})
		})

		Convey("When app is ready and listening is dynamically changed", func() {
			Convey("Updates ready status", func() {
				enableReadyAndLivenHealthCheck()
				assertHealthStatusCodeIsGood(ReadyEndpoint)

				SetAppIsLive(false)
				assertHealthStatusCodeIsBad(ReadyEndpoint)

				SetAppIsLive(true)
				assertHealthStatusCodeIsGood(ReadyEndpoint)
			})

			Convey("Updates live status", func() {
				enableReadyAndLivenHealthCheck()
				assertHealthStatusCodeIsGood(LiveEndpoint)

				SetAppIsLive(false)
				assertHealthStatusCodeIsBad(LiveEndpoint)

				SetAppIsLive(true)
				assertHealthStatusCodeIsGood(LiveEndpoint)
			})

			Convey("Updates expected ready JSON", func() {
				enableReadyAndLivenHealthCheck()

				health, err := getHealth(VerboseReadyEndpoint)
				assertReadyJSONIsGood(err, health)
				assertListeningJSONIsGood(err, health)

				SetAppIsLive(false)
				health, err = getHealth(VerboseReadyEndpoint)
				assertReadyJSONIsGood(nil, health)
				assertListeningJSONIsBad(err, health)

				SetAppIsLive(true)
				health, err = getHealth(VerboseReadyEndpoint)
				assertReadyJSONIsGood(err, health)
				assertListeningJSONIsGood(err, health)
			})

			Convey("Updates expected live JSON", func() {
				enableReadyAndLivenHealthCheck()

				health, err := getHealth(VerboseLiveEndpoint)
				assertReadyJSONIsNotPresent(health)
				assertListeningJSONIsGood(err, health)

				SetAppIsLive(false)
				health, err = getHealth(VerboseLiveEndpoint)
				assertReadyJSONIsNotPresent(health)
				assertListeningJSONIsBad(err, health)

				SetAppIsLive(true)
				health, err = getHealth(VerboseLiveEndpoint)
				assertReadyJSONIsNotPresent(health)
				assertListeningJSONIsGood(err, health)
			})

			Reset(func() {
				disableHealthCheck()
			})
		})
	})
}
