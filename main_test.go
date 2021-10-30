package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

var a App

func TestMain(m *testing.M) {
	a = App{}
	a.Initialize("./requesty_test.db")

	clearTable()
	fillDBWithDummyData()

	code := m.Run()
	os.Exit(code)
}

func TestStatisticsLastMinuteRequest(t *testing.T) {
	req, _ := http.NewRequest("GET", "/domains/statistics/last-minute", nil)
	response := executeRequest(req)

	if strings.Contains(response.Body.String(), "afterroundminute.com") || strings.Contains(response.Body.String(), "beforeroundminute.com") {
		t.Error("Output data contains a domain which was added before or after the last round minute")
	}

	if !strings.Contains(response.Body.String(), "withinroundminute.com") {
		t.Error("Output data doesn't contain a domain which was added during the last round minute")
	}

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestStatisticsLastHourRequest(t *testing.T) {
	req, _ := http.NewRequest("GET", "/domains/statistics/last-hour", nil)
	response := executeRequest(req)

	if strings.Contains(response.Body.String(), "afterroundhour.com") || strings.Contains(response.Body.String(), "beforeroundhour.com") {
		t.Error("Output data contains a domain which was added before or after the last round hour")
	}

	if !strings.Contains(response.Body.String(), "withinroundhour.com") {
		t.Error("Output data doesn't contain a domain which was added during the last round hour")
	}

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestDomainsPost(t *testing.T) {

	values := getTestValues()

	jsonValue, _ := json.Marshal(values)

	req, _ := http.NewRequest("POST", "/domains", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)

	fmt.Print(response.Body)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	responseRecorder := httptest.NewRecorder()
	a.Router.ServeHTTP(responseRecorder, req)

	return responseRecorder
}

func getTestValues() *map[string]int64 {
	timestamp := time.Now().Unix()

	m := map[string]int64{"timestamp": timestamp, "example.cpm": 4}

	return &m
}

func clearTable() {
	a.DB.Exec("DELETE FROM domains")
	a.DB.Exec("ALTER SEQUENCE domains_id_seq RESTART WITH 1")
}

func fillDBWithDummyData() {
	// getting current time.
	t := time.Now()
	// finding how much seconds passed since the last round minute
	s := int64(t.Second())
	// getting unix epoch time and substract the seconds to get last round minute.
	ut := (t.Unix() - s - 1) // :59

	// last minute data

	d := domain{
		Timestamp: int(ut),
		Name:      "withinroundminute.com",
		Requests:  1,
	}

	d.createDomain(a.DB)

	d = domain{
		Timestamp: int(ut - 60),
		Name:      "afterroundminute.com",
		Requests:  1,
	}

	d.createDomain(a.DB)

	d = domain{
		Timestamp: int(ut + 1),
		Name:      "beforeroundminute.com",
		Requests:  1,
	}

	d.createDomain(a.DB)

	// last hour data

	// finding how much seconds passed since the last round hour
	s = int64(t.Second() + t.Minute()*60)
	// getting unix epoch time and substract the seconds to get last round hour.
	ut = (t.Unix() - s - 1)

	d = domain{
		Timestamp: int(ut),
		Name:      "withinroundhour.com",
		Requests:  1,
	}

	d.createDomain(a.DB)

	d = domain{
		Timestamp: int(ut - 60*60),
		Name:      "afterroundhour.com",
		Requests:  1,
	}

	d.createDomain(a.DB)

	d = domain{
		Timestamp: int(ut + 1),
		Name:      "beforeroundhour.com",
		Requests:  1,
	}

	d.createDomain(a.DB)

}
