package api

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rwirdemann/datafrog/pkg/df"
	"github.com/rwirdemann/datafrog/pkg/mocks"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var testname string

func init() {
	testname = "create-job"
}

func TestStartRecordingNoChannels(t *testing.T) {
	logFactory := mocks.LogFactory{}
	repository := &mocks.TestRepository{}
	rr := startRecording(t, logFactory, repository)
	assert.Equal(t, http.StatusFailedDependency, rr.Code)
}

func TestStartStopRecording(t *testing.T) {
	config.Channels = append(config.Channels, df.Channel{})
	logFactory := mocks.LogFactory{}
	repository := &mocks.TestRepository{}
	rr := startRecording(t, logFactory, repository)
	assert.Equal(t, http.StatusAccepted, rr.Code)
	runner, ok := runners[testname]
	assert.True(t, ok)
	runner.Stop()
	tc, err := repository.Get(testname)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "create-job", tc.Name)
}

func startRecording(t *testing.T, logFactory df.LogFactory, repository df.TestRepository) *httptest.ResponseRecorder {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/tests/%s/recordings", testname), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/tests/{name}/recordings", StartRecording(logFactory, repository)).Methods("POST")
	r.ServeHTTP(rr, req)
	return rr
}
