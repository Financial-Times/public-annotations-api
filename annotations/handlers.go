package annotations

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/gorilla/mux"
)

// HandlerCtx contains objects needed from the annotations http handlers and is being passed to them as param
type HandlerCtx struct {
	AnnotationsDriver  driver
	CacheControlHeader string
	Log                *logger.UPPLogger
}

func NewHandlerCtx(d driver, ch string, log *logger.UPPLogger) *HandlerCtx {
	return &HandlerCtx{
		AnnotationsDriver:  d,
		CacheControlHeader: ch,
		Log:                log,
	}
}

// MethodNotAllowedHandler handles 405
func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func GetAnnotations(hctx *HandlerCtx) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uuid := vars["uuid"]

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if uuid == "" {
			http.Error(w, "uuid required", http.StatusBadRequest)
			return
		}

		params := r.URL.Query()

		var ok bool
		var lifecycleParams []string
		if lifecycleParams, ok = params["lifecycle"]; ok {
			err := validateLifecycleParams(lifecycleParams)
			if err != nil {
				hctx.Log.WithError(err).Error("invalid query parameter")
				w.WriteHeader(http.StatusBadRequest)
				msg := `{"message":"invalid query parameter"}`
				if _, err = w.Write([]byte(msg)); err != nil {
					hctx.Log.WithError(err).Errorf("Error while writing response: %s", msg)
				}
				return
			}
		}

		annotations, found, err := hctx.AnnotationsDriver.read(uuid)
		if err != nil {
			hctx.Log.WithError(err).WithUUID(uuid).Error("failed getting annotations for content")
			writeResponseError(hctx, w, http.StatusServiceUnavailable, uuid, `{"message":"Error getting annotations for content with uuid %s"}`)
			return
		}
		if !found {
			writeResponseError(hctx, w, http.StatusNotFound, uuid, `{"message":"No annotations found for content with uuid %s."}`)
			return
		}

		lifecycleFilter := newLifecycleFilter(withLifecycles(lifecycleParams))
		predicateFilter := NewAnnotationsPredicateFilter()
		chain := newAnnotationsFilterChain(lifecycleFilter, predicateFilter)

		annotations = chain.doNext(annotations)

		w.Header().Set("Cache-Control", hctx.CacheControlHeader)
		w.WriteHeader(http.StatusOK)

		if err = json.NewEncoder(w).Encode(annotations); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			msg := fmt.Sprintf(`{"message":"Error parsing annotations for content with uuid %s, err=%s"}`, uuid, err.Error())
			hctx.Log.Error(msg)
			if _, err = w.Write([]byte(msg)); err != nil {
				hctx.Log.WithError(err).Errorf("Error while writing response: %s", msg)
			}
		}
	}
}

func writeResponseError(hctx *HandlerCtx, w http.ResponseWriter, status int, uuid, message string) {
	w.WriteHeader(status)
	msg := fmt.Sprintf(message, uuid)
	if _, err := w.Write([]byte(msg)); err != nil {
		hctx.Log.WithError(err).Errorf("Error while writing response: %s", msg)
	}
}

func validateLifecycleParams(lifecycleParams []string) error {
	for _, lp := range lifecycleParams {
		if _, ok := lifecycleMap[lp]; !ok {
			return fmt.Errorf("invalid lifecycle value: %s", lp)
		}
	}

	return nil
}
