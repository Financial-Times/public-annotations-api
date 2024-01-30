package annotations

import (
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
			req:  newRequest(fmt.Sprintf("/content/%s/annotations", knownUUID)),
			annotationsDriver: mockDriver{
				readFunc: func(string, string) (anns Annotations, found bool, err error) {
					return []Annotation{}, true, nil
				},
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "{\"message\":\"No annotations found for content with uuid 12345 for the specified filters.\"}",
		},
		{
			name: "NotFound",
			req:  newRequest(fmt.Sprintf("/content/%s/annotations", "99999")),
			annotationsDriver: mockDriver{
				readFunc: func(string, string) (anns Annotations, found bool, err error) {
					return []Annotation{}, false, nil
				},
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       message("No annotations found for content with uuid 99999."),
		},
		{
			name: "ReadError",
			req:  newRequest(fmt.Sprintf("/content/%s/annotations", knownUUID)),
			annotationsDriver: mockDriver{
				readFunc: func(string, string) (anns Annotations, found bool, err error) {
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
			Log:                logger.NewUPPLogger("test-public-annotations-api", "panic"),
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
		expectedAnnotations Annotations
	}{
		"request with valid lifecycle parameter should succeed": {
			annotationsDriver: mockDriver{
				readFunc: func(string, string) (anns Annotations, found bool, err error) {
					return []Annotation{}, true, nil
				},
			},
			lifecycleParams:    "lifecycle=pac",
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "{\"message\":\"No annotations found for content with uuid 12345 for the specified filters.\"}",
		},
		"request with invalid lifecycle parameter should fail": {
			annotationsDriver: mockDriver{
				readFunc: func(string, string) (anns Annotations, found bool, err error) {
					return []Annotation{}, true, nil
				},
			},
			lifecycleParams:    "lifecycle=invalid",
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"invalid query parameter"}`,
		},
		"request with lifecycle parameters should apply additional filtering": {
			annotationsDriver: mockDriver{
				readFunc: func(string, string) (anns Annotations, found bool, err error) {
					return []Annotation{pacAnnotationA, pacAnnotationB, v1AnnotationA, v1AnnotationB, v2AnnotationA, v2AnnotationB}, true, nil
				},
			},
			lifecycleParams:    "lifecycle=pac&lifecycle=v1",
			expectedStatusCode: http.StatusOK,
			expectedAnnotations: Annotations{
				Annotation{
					Predicate: "http://www.ft.com/ontology/annotation/about",
					ID:        "6bbd0457-15ab-4ddc-ab82-0cd5b8d9ce18",
				},
				Annotation{
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

			actualAnns := Annotations{}
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
			req:  newRequest(fmt.Sprintf("/content/%s/annotations/", knownUUID)),
			annotationsDriver: mockDriver{
				readFunc: func(string, string) (anns Annotations, found bool, err error) {
					return []Annotation{}, true, nil
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

func TestGetHandlerWithSetBookmarkHeader(t *testing.T) {
	tests := []struct {
		name               string
		req                *http.Request
		annotationsDriver  mockDriver
		bookmark           string
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "Empty bookmark",
			req:  newRequest(fmt.Sprintf("/content/%s/annotations", knownUUID)),
			annotationsDriver: mockDriver{
				readFunc: func(uuid string, bookmark string) (anns Annotations, found bool, err error) {
					if bookmark != "" {
						return []Annotation{}, false, errors.New("unexpected bookmark")
					}

					return Annotations{
						Annotation{
							Predicate: "http://www.ft.com/ontology/annotation/about",
							ID:        "6bbd0457-15ab-4ddc-ab82-0cd5b8d9ce18",
						},
					}, true, nil
				},
			},
			bookmark:           "",
			expectedStatusCode: http.StatusOK,
			expectedBody:       "[{\"predicate\":\"http://www.ft.com/ontology/annotation/about\",\"id\":\"6bbd0457-15ab-4ddc-ab82-0cd5b8d9ce18\",\"apiUrl\":\"\",\"types\":null}]",
		},
		{
			name: "Not empty bookmark",
			req:  newRequest(fmt.Sprintf("/content/%s/annotations", knownUUID)),
			annotationsDriver: mockDriver{
				readFunc: func(uuid string, bookmark string) (anns Annotations, found bool, err error) {
					if bookmark != "FB:kcwQnrEEnFpfSJ2PtiykK/JNh8oBozhIkA==" {
						return []Annotation{}, false, errors.New("unexpected bookmark")
					}

					return Annotations{
						Annotation{
							Predicate: "http://www.ft.com/ontology/annotation/about",
							ID:        "6bbd0457-15ab-4ddc-ab82-0cd5b8d9ce18",
						},
					}, true, nil
				},
			},
			bookmark:           "FB:kcwQnrEEnFpfSJ2PtiykK/JNh8oBozhIkA==",
			expectedStatusCode: http.StatusOK,
			expectedBody:       "[{\"predicate\":\"http://www.ft.com/ontology/annotation/about\",\"id\":\"6bbd0457-15ab-4ddc-ab82-0cd5b8d9ce18\",\"apiUrl\":\"\",\"types\":null}]",
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

		// Set the test bookmark in the request header.
		// The verification that the correct bookmark header was sent to the read method is checked in the mock read
		// method of the test object.
		test.req.Header.Add(Neo4jBookmarkHeader, test.bookmark)
		r.ServeHTTP(rec, test.req)

		assert.True(t, test.expectedStatusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.expectedStatusCode))
		assert.JSONEq(t, test.expectedBody, rec.Body.String(), fmt.Sprintf("%s: Wrong body", test.name))
	}
}

func newRequest(url string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	return req
}

func message(errMsg string) string {
	return fmt.Sprintf("{\"message\": \"%s\"}\n", errMsg)
}

type mockDriver struct {
	readFunc              func(string, string) (Annotations, bool, error)
	checkConnectivityFunc func() error
}

func (md mockDriver) read(contentUUID, bookmark string) (Annotations, bool, error) {
	if md.readFunc == nil {
		return nil, false, errors.New("not implemented")
	}

	return md.readFunc(contentUUID, bookmark)
}

func (md mockDriver) checkConnectivity() error {
	if md.checkConnectivityFunc == nil {
		return errors.New("not implemented")
	}

	return md.checkConnectivityFunc()
}
