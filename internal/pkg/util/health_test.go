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

var listenerPort = strconv.Itoa(DEFAULT_HEALTH_CHECK_PORT)

var READY_ENDPOINT = "http://localhost:" + listenerPort + "/ready"
var LIVE_ENDPOINT = "http://localhost:" + listenerPort + "/live"

var VERBOSE_READY_ENDPOINT = "http://localhost:" + listenerPort + "/ready?full=1"
var VERBOSE_LIVE_ENDPOINT = "http://localhost:" + listenerPort + "/live?full=1"

type HealthJson struct {
	Ready     string `json:"ready,omitempty"`
	Listening string `json:"listening,omitempty"`
}

func getHealth(endpoint string) (error, *HealthJson) {
	webClient := http.Client{
		Timeout: time.Second * 2,
	}

	request, requestErr := http.NewRequest(http.MethodGet, endpoint, nil)
	if requestErr != nil {
		return requestErr, nil
	}

	response, responseErr := webClient.Do(request)
	if responseErr != nil {
		return responseErr, nil
	}

	body, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return readErr, nil
	}

	healthJson := &HealthJson{}
	unmarshalErr := json.Unmarshal(body, healthJson)
	if unmarshalErr != nil {
		return unmarshalErr, nil
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		err := fmt.Errorf("Response Status: %d: %s",
			response.StatusCode,
			http.StatusText(response.StatusCode))
		return err, healthJson
	}

	fmt.Println(healthJson)

	return nil, healthJson
}

func assertHealthStatusCodeIsBad(endpoint string) {
	err, _ := getHealth(endpoint)

	So(err, ShouldNotBeNil)
	errorMsg := "Response Status: 503: Service Unavailable"
	So(err.Error(), ShouldEqual, errorMsg)
}

func assertHealthStatusCodeIsGood(endpoint string) {
	err, _ := getHealth(endpoint)

	So(err, ShouldBeNil)
}

func assertReadyJSONIsBad(err error, healthJson *HealthJson) {
	So(err, ShouldNotBeNil)
	errorMsg := "Response Status: 503: Service Unavailable"
	So(err.Error(), ShouldEqual, errorMsg)

	So(healthJson, ShouldNotBeNil)

	So((*healthJson).Ready, ShouldEqual, "Secretless is not ready")
}

func assertReadyJSONIsGood(err error, healthJson *HealthJson) {
	So(err, ShouldBeNil)
	So(healthJson, ShouldNotBeNil)
	So((*healthJson).Ready, ShouldEqual, "OK")
}

func assertReadyJSONIsNotPresent(healthJson *HealthJson) {
	So(healthJson, ShouldNotBeNil)
	So((*healthJson).Ready, ShouldEqual, "")
}

func assertListeningJSONIsBad(err error, healthJson *HealthJson) {
	So(err, ShouldNotBeNil)
	errorMsg := "Response Status: 503: Service Unavailable"
	So(err.Error(), ShouldEqual, errorMsg)

	So(healthJson, ShouldNotBeNil)

	So((*healthJson).Listening, ShouldEqual, "Secretless is not listening")
}

func assertListeningJSONIsGood(err error, healthJson *HealthJson) {
	So(err, ShouldBeNil)
	So(healthJson, ShouldNotBeNil)
	So((*healthJson).Listening, ShouldEqual, "OK")
}

func enableHealthCheck() {
	EnableHealthCheck()

	// Server can be slow to come up :(
	time.Sleep(200 * time.Millisecond)
}

func enableAndReadyHealthCheck() {
	enableHealthCheck()
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
				enableHealthCheck()
				assertHealthStatusCodeIsBad(READY_ENDPOINT)
			})

			Convey("Shows not live", func() {
				enableHealthCheck()
				assertHealthStatusCodeIsBad(LIVE_ENDPOINT)
			})

			Convey("Shows expected not ready JSON", func() {
				enableHealthCheck()

				err, health := getHealth(VERBOSE_READY_ENDPOINT)
				assertReadyJSONIsBad(err, health)
				assertListeningJSONIsBad(err, health)
			})

			Convey("Shows expected not live JSON", func() {
				enableHealthCheck()

				err, health := getHealth(VERBOSE_LIVE_ENDPOINT)
				assertReadyJSONIsNotPresent(health)
				assertListeningJSONIsBad(err, health)
			})

			Reset(func() {
				DisableHealthCheck()
			})
		})

		Convey("When readniness flag is turned on but liveliness is not", func() {
			Convey("Shows not ready", func() {
				enableAndReadyHealthCheck()
				assertHealthStatusCodeIsBad(READY_ENDPOINT)
			})

			Convey("Shows not live", func() {
				enableAndReadyHealthCheck()
				assertHealthStatusCodeIsBad(LIVE_ENDPOINT)
			})

			Convey("Shows expected ready JSON but bad listening status", func() {
				enableAndReadyHealthCheck()

				err, health := getHealth(VERBOSE_READY_ENDPOINT)

				// We ignore error manually since we know it will be checked below
				assertReadyJSONIsGood(nil, health)

				assertListeningJSONIsBad(err, health)
			})

			Convey("Shows expected not live JSON", func() {
				enableAndReadyHealthCheck()

				err, health := getHealth(VERBOSE_LIVE_ENDPOINT)
				assertReadyJSONIsNotPresent(health)
				assertListeningJSONIsBad(err, health)
			})

			Reset(func() {
				DisableHealthCheck()
			})
		})

		Convey("When app is ready and listening", func() {
			Convey("Shows ready", func() {
				enableReadyAndLivenHealthCheck()
				assertHealthStatusCodeIsGood(READY_ENDPOINT)
			})

			Convey("Shows live", func() {
				enableReadyAndLivenHealthCheck()
				assertHealthStatusCodeIsGood(LIVE_ENDPOINT)
			})

			Convey("Shows expected ready JSON", func() {
				enableReadyAndLivenHealthCheck()

				err, health := getHealth(VERBOSE_READY_ENDPOINT)

				assertReadyJSONIsGood(err, health)
				assertListeningJSONIsGood(err, health)
			})

			Convey("Shows expected live JSON", func() {
				enableReadyAndLivenHealthCheck()

				err, health := getHealth(VERBOSE_LIVE_ENDPOINT)
				assertReadyJSONIsNotPresent(health)
				assertListeningJSONIsGood(err, health)
			})

			Reset(func() {
				DisableHealthCheck()
			})
		})

		Convey("When app is ready and listening is dynamically changed", func() {
			Convey("Updates ready status", func() {
				enableReadyAndLivenHealthCheck()
				assertHealthStatusCodeIsGood(READY_ENDPOINT)

				SetAppIsLive(false)
				assertHealthStatusCodeIsBad(READY_ENDPOINT)

				SetAppIsLive(true)
				assertHealthStatusCodeIsGood(READY_ENDPOINT)
			})

			Convey("Updates live status", func() {
				enableReadyAndLivenHealthCheck()
				assertHealthStatusCodeIsGood(LIVE_ENDPOINT)

				SetAppIsLive(false)
				assertHealthStatusCodeIsBad(LIVE_ENDPOINT)

				SetAppIsLive(true)
				assertHealthStatusCodeIsGood(LIVE_ENDPOINT)
			})

			Convey("Updates expected ready JSON", func() {
				enableReadyAndLivenHealthCheck()

				err, health := getHealth(VERBOSE_READY_ENDPOINT)
				assertReadyJSONIsGood(err, health)
				assertListeningJSONIsGood(err, health)

				SetAppIsLive(false)
				err, health = getHealth(VERBOSE_READY_ENDPOINT)
				assertReadyJSONIsGood(nil, health)
				assertListeningJSONIsBad(err, health)

				SetAppIsLive(true)
				err, health = getHealth(VERBOSE_READY_ENDPOINT)
				assertReadyJSONIsGood(err, health)
				assertListeningJSONIsGood(err, health)
			})

			Convey("Updates expected live JSON", func() {
				enableReadyAndLivenHealthCheck()

				err, health := getHealth(VERBOSE_LIVE_ENDPOINT)
				assertReadyJSONIsNotPresent(health)
				assertListeningJSONIsGood(err, health)

				SetAppIsLive(false)
				err, health = getHealth(VERBOSE_LIVE_ENDPOINT)
				assertReadyJSONIsNotPresent(health)
				assertListeningJSONIsBad(err, health)

				SetAppIsLive(true)
				err, health = getHealth(VERBOSE_LIVE_ENDPOINT)
				assertReadyJSONIsNotPresent(health)
				assertListeningJSONIsGood(err, health)
			})

			Reset(func() {
				DisableHealthCheck()
			})
		})
	})
}
