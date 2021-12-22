package annotations

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/stretchr/testify/assert"
)

func TestGTGUnhealthyCluster(t *testing.T) {
	//create a request to pass to our handler
	req := httptest.NewRequest("GET", "/__gtg", nil)

	annotationsDriver := mockDriver{
		checkConnectivityFunc: func() error {
			return errors.New("test error")
		},
	}
	hctx := &HandlerCtx{
		AnnotationsDriver:  annotationsDriver,
		CacheControlHeader: "test-header",
		Log:                logger.NewUPPLogger("test-public-annotations-api", "PANIC"),
	}

	//create a responseRecorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(httphandlers.NewGoodToGoHandler(GoodToGo(hctx)))

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)
	actual := rr.Result()
	defer actual.Body.Close()

	// Series of verifications:
	assert.Equal(t, http.StatusServiceUnavailable, actual.StatusCode, "status code")
	assert.Equal(t, "no-cache", actual.Header.Get("Cache-Control"), "cache-control header")
	assert.Equal(t, "test error", rr.Body.String(), "GTG response body")
}

func TestGTGHealthyCluster(t *testing.T) {
	//create a request to pass to our handler
	req := httptest.NewRequest("GET", "/__gtg", nil)
	annotationsDriver := mockDriver{
		checkConnectivityFunc: func() error {
			return nil
		},
	}
	hctx := &HandlerCtx{
		AnnotationsDriver:  annotationsDriver,
		CacheControlHeader: "test-header",
		Log:                logger.NewUPPLogger("test-public-annotations-api", "PANIC"),
	}
	//create a responseRecorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(httphandlers.NewGoodToGoHandler(GoodToGo(hctx)))

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)
	actual := rr.Result()
	defer actual.Body.Close()

	// Series of verifications:
	assert.Equal(t, http.StatusOK, actual.StatusCode, "status code")
	assert.Equal(t, "no-cache", actual.Header.Get("Cache-Control"), "cache-control header")
	assert.Equal(t, "OK", rr.Body.String(), "GTG response body")
}

func TestNeo4jConnectivityChecker_Healthy(t *testing.T) {
	annotationsDriver := mockDriver{
		checkConnectivityFunc: func() error {
			return nil
		},
	}
	message, err := Neo4jChecker(annotationsDriver)()
	assert.Equal(t, "Connectivity to neo4j is ok", message)
	assert.Equal(t, nil, err)
}

func TestNeo4jConnectivityChecker_Unhealthy(t *testing.T) {
	annotationsDriver := mockDriver{
		checkConnectivityFunc: func() error {
			return errors.New("test error")
		},
	}
	message, err := Neo4jChecker(annotationsDriver)()
	assert.Equal(t, "Error connecting to neo4j", message)
	assert.Equal(t, "test error", err.Error())
}
