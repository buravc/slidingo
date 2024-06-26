package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var a *App

func TestMain(m *testing.M) {
	filePath := flag.String("save", "/tmp/requestcounter.json", "file path the state will be saved to")
	addr := flag.String("addr", ":3000", "address the server listens")
	autosaveDurationStr := flag.String("autosave", "30s", "autosave interval")
	counterWindowStr := flag.String("window", "60s", "window of the request counter")

	maxConcurReq := flag.Int("maxCon", 5, "maximum number of concurrent requests")
	timeoutStr := flag.String("timeout", "300ms", "timeout window")
	flag.Parse()

	autosaveDuration, err := time.ParseDuration(*autosaveDurationStr)
	if err != nil {
		panic(fmt.Errorf("invalid duration is passed as an argument: %w", err))
	}

	counterWindow, err := time.ParseDuration(*counterWindowStr)
	if err != nil {
		panic(fmt.Errorf("invalid window is passed as an argument: %w", err))
	}

	timeout, err := time.ParseDuration(*timeoutStr)
	if err != nil {
		panic(fmt.Errorf("invalid timeout is passed as an argument: %w", err))
	}

	a = NewAppWithZeroState(*filePath, *addr, *maxConcurReq, timeout, autosaveDuration, counterWindow)

	go a.Start()

	code := m.Run()

	os.Exit(code)
}

func Test_ParallelRequest_HalfTimeout(t *testing.T) {
	expectedTimeouts := 5
	expectedSuccesses := 5

	expectedBodies := map[int]struct{}{
		1: {},
		2: {},
		3: {},
		4: {},
		5: {},
	}

	codeChan, bodyChan := execRequests(t, 10)

	assertStatusCodes(t, codeChan, expectedSuccesses, expectedTimeouts)

	assertResponseBody(t, bodyChan, expectedBodies)

	close(codeChan)
}

func Test_ParallelRequest_AllSuccess(t *testing.T) {
	expectedTimeouts := 0
	expectedSuccesses := 10

	expectedBodies := map[int]struct{}{
		1:  {},
		2:  {},
		3:  {},
		4:  {},
		5:  {},
		6:  {},
		7:  {},
		8:  {},
		9:  {},
		10: {},
	}

	a.SetTimeout(time.Second * 3)
	a.Clear()

	codeChan, bodyChan := execRequests(t, 10)

	assertStatusCodes(t, codeChan, expectedSuccesses, expectedTimeouts)

	assertResponseBody(t, bodyChan, expectedBodies)

	close(codeChan)
}

func execRequests(t *testing.T, count int) (chan int, chan io.ReadCloser) {
	codeChan := make(chan int)
	bodyChan := make(chan io.ReadCloser)
	for i := 0; i < count; i++ {
		go func() {
			req, err := http.NewRequest(http.MethodGet, "/", nil)
			if err != nil {
				t.Errorf("unable to create http request")
			}
			response := execReq(req)

			codeChan <- response.Code
			if response.Code == http.StatusOK {
				bodyChan <- response.Result().Body
			}
		}()
	}

	return codeChan, bodyChan
}

func assertStatusCodes(t *testing.T, codeChan chan int, expectedSuccesses, expectedTimeouts int) {
	actualTimeouts := 0
	actualSuccesses := 0

	for i := 0; i < 10; i++ {
		code := <-codeChan
		if code == http.StatusServiceUnavailable {
			actualTimeouts++
			continue
		}

		if code == http.StatusOK {
			actualSuccesses++
		}
	}

	if actualTimeouts != expectedTimeouts {
		t.Errorf("expected timeout error count does not match: expected: %d, actual: %d", expectedTimeouts, actualTimeouts)
	}

	if actualSuccesses != expectedSuccesses {
		t.Errorf("expected success count does not match: expected: %d, actual: %d", expectedSuccesses, actualSuccesses)
	}

}

func assertResponseBody(t *testing.T, bodyChan chan io.ReadCloser, expectedBodies map[int]struct{}) {
	length := len(expectedBodies)

	for i := 0; i < length; i++ {
		body := <-bodyChan
		actualBody := bodyToInt(t, body)
		if _, ok := expectedBodies[actualBody]; ok {
			delete(expectedBodies, actualBody)
		} else {
			t.Errorf("received unexpected response body: %d", actualBody)
		}
	}

	if len(expectedBodies) != 0 {
		t.Errorf("unable to receive some expected response bodies: %d", len(expectedBodies))
	}
}

func bodyToInt(t *testing.T, responseBody io.ReadCloser) int {
	var num int
	if err := json.NewDecoder(responseBody).Decode(&num); err != nil {
		t.Error(err)
	}

	return num
}

func execReq(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.server.Handler.ServeHTTP(rr, req)

	return rr
}
