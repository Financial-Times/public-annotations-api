package annotations

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const (
	knownUUID = "12345"
)

func TestGetHandler(t *testing.T) {
	tests := []struct {
		name               string
		req                *http.Request
		annotationsDriver  mockDriver
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "Success",
			req:  newRequest("GET", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", nil),
			annotationsDriver: mockDriver{
				readFunc: func(string) (anns annotations, found bool, err error) {
					return []annotation{}, true, nil
				},
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "null",
		},
		{
			name: "NotFound",
			req:  newRequest("GET", fmt.Sprintf("/content/%s/annotations", "99999"), "application/json", nil),
			annotationsDriver: mockDriver{
				readFunc: func(string) (anns annotations, found bool, err error) {
					return []annotation{}, false, nil
				},
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       message("No annotations found for content with uuid 99999."),
		},
		{
			name: "ReadError",
			req:  newRequest("GET", fmt.Sprintf("/content/%s/annotations", knownUUID), "application/json", nil),
			annotationsDriver: mockDriver{
				readFunc: func(string) (anns annotations, found bool, err error) {
					return nil, false, errors.New("TEST failing to READ")
				},
			},
			expectedStatusCode: http.StatusServiceUnavailable,
			expectedBody:       message("Error getting annotations for content with uuid 12345"),
		},
	}

	for _, test := range tests {
		hctx := &HandlerCtx{
			AnnotationsDriver:  test.annotationsDriver,
			CacheControlHeader: "test-header",
			Log:                logger.NewUPPInfoLogger("test-public-annotations-api"),
		}
		rec := httptest.NewRecorder()
		r := mux.NewRouter()
		r.HandleFunc("/content/{uuid}/annotations", GetAnnotations(hctx)).Methods("GET")
		r.ServeHTTP(rec, test.req)
		assert.True(t, test.expectedStatusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.expectedStatusCode))
		assert.JSONEq(t, test.expectedBody, rec.Body.String(), fmt.Sprintf("%s: Wrong body", test.name))
	}
}

func TestGetHandlerWithLifecycleQueryParams(t *testing.T) {
	tests := map[string]struct {
		annotationsDriver   mockDriver
		lifecycleParams     string
		expectedStatusCode  int
		expectedBody        string
		expectedAnnotations annotations
	}{
		"request with valid lifecycle parameter should succeed": {
			annotationsDriver: mockDriver{
				readFunc: func(string) (anns annotations, found bool, err error) {
					return []annotation{}, true, nil
				},
			},
			lifecycleParams:    "lifecycle=pac",
			expectedStatusCode: http.StatusOK,
			expectedBody:       "null",
		},
		"request with invalid lifecycle parameter should fail": {
			annotationsDriver: mockDriver{
				readFunc: func(string) (anns annotations, found bool, err error) {
					return []annotation{}, true, nil
				},
			},
			lifecycleParams:    "lifecycle=invalid",
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"invalid query parameter"}`,
		},
		"request with lifecycle parameters should apply additional filtering": {
			annotationsDriver: mockDriver{
				readFunc: func(string) (anns annotations, found bool, err error) {
					return []annotation{pacAnnotationA, pacAnnotationB, v1AnnotationA, v1AnnotationB, v2AnnotationA, v2AnnotationB}, true, nil
				},
			},
			lifecycleParams:    "lifecycle=pac&lifecycle=v1",
			expectedStatusCode: http.StatusOK,
			expectedAnnotations: annotations{
				annotation{
					Predicate: "http://www.ft.com/ontology/annotation/about",
					ID:        "6bbd0457-15ab-4ddc-ab82-0cd5b8d9ce18",
				},
				annotation{
					Predicate: "http://www.ft.com/ontology/annotation/mentions",
					ID:        "0ab61bfc-a2b1-4b08-a864-4233fd72f250",
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			hctx := &HandlerCtx{
				AnnotationsDriver:  tc.annotationsDriver,
				CacheControlHeader: "test-header",
				Log:                logger.NewUPPLogger("test-public-annotations-api", "PANIC"),
			}
			req, err := http.NewRequest("GET", fmt.Sprintf("/content/%s/annotations?%s", knownUUID, tc.lifecycleParams), nil)
			if err != nil {
				t.Fatal(err)
			}

			rec := httptest.NewRecorder()
			r := mux.NewRouter()
			r.HandleFunc("/content/{uuid}/annotations", GetAnnotations(hctx)).Methods("GET")
			r.ServeHTTP(rec, req)
			assert.True(t, tc.expectedStatusCode == rec.Code, fmt.Sprintf("Wrong response code, was %d, should be %d", rec.Code, tc.expectedStatusCode))
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rec.Body.String(), "Wrong error response body")
				return
			}

			actualAnns := annotations{}
			err = json.Unmarshal(rec.Body.Bytes(), &actualAnns)
			if err != nil {
				t.Fatal(err)
			}
			assert.ElementsMatch(t, tc.expectedAnnotations, actualAnns, "Wrong response body")
		})
	}
}

func TestMethodeNotFound(t *testing.T) {
	tests := []struct {
		name               string
		req                *http.Request
		annotationsDriver  mockDriver
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "NotFound",
			req:  newRequest("GET", fmt.Sprintf("/content/%s/annotations/", knownUUID), "application/json", nil),
			annotationsDriver: mockDriver{
				readFunc: func(string) (anns annotations, found bool, err error) {
					return []annotation{}, true, nil
				},
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "404 page not found\n",
		},
	}

	for _, test := range tests {
		hctx := &HandlerCtx{
			AnnotationsDriver:  test.annotationsDriver,
			CacheControlHeader: "test-header",
			Log:                logger.NewUPPInfoLogger("test-public-annotations-api"),
		}
		rec := httptest.NewRecorder()
		r := mux.NewRouter()
		r.HandleFunc("/content/{uuid}/annotations", GetAnnotations(hctx)).Methods("GET")
		r.ServeHTTP(rec, test.req)
		assert.True(t, test.expectedStatusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.expectedStatusCode))
		assert.Equal(t, test.expectedBody, rec.Body.String(), fmt.Sprintf("%s: Wrong body", test.name))
	}
}

func newRequest(method, url, contentType string, body []byte) *http.Request {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", contentType)
	return req
}

func message(errMsg string) string {
	return fmt.Sprintf("{\"message\": \"%s\"}\n", errMsg)
}

type mockDriver struct {
	readFunc              func(string) (annotations, bool, error)
	checkConnectivityFunc func() error
}

func (md mockDriver) read(contentUUID string) (annotations, bool, error) {
	if md.readFunc == nil {
		return nil, false, errors.New("not implemented")
	}

	return md.readFunc(contentUUID)
}

func (md mockDriver) checkConnectivity() error {
	if md.checkConnectivityFunc == nil {
		return errors.New("not implemented")
	}

	return md.checkConnectivityFunc()
}
