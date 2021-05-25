package main

import (
	"net/http"
	"os"

	"fmt"
	"strconv"
	"time"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/http-handlers-go/v2/httphandlers"
	"github.com/Financial-Times/neo-utils-go/v2/neoutils"

	apiEndpoint "github.com/Financial-Times/api-endpoint"
	"github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/public-annotations-api/v3/annotations"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/gorilla/mux"
	cli "github.com/jawher/mow.cli"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rcrowley/go-metrics"
)

const (
	appDescription = "A public RESTful API for accessing Annotations in neo4j"
)

func main() {
	app := cli.App("public-annotations-api", appDescription)
	neoURL := app.String(cli.StringOpt{
		Name:   "neo-url",
		Value:  "http://localhost:7474/db/data",
		Desc:   "neo4j endpoint URL",
		EnvVar: "NEO_URL"})
	port := app.String(cli.StringOpt{
		Name:   "port",
		Value:  "8080",
		Desc:   "Port to listen on",
		EnvVar: "PORT",
	})
	env := app.String(cli.StringOpt{
		Name:  "env",
		Value: "local",
		Desc:  "environment this app is running in",
	})
	cacheDuration := app.String(cli.StringOpt{
		Name:   "cache-duration",
		Value:  "30s",
		Desc:   "Duration Get requests should be cached for. e.g. 2h45m would set the max-age value to '7440' seconds",
		EnvVar: "CACHE_DURATION",
	})
	logLevel := app.String(cli.StringOpt{
		Name:   "log-level",
		Value:  "info",
		Desc:   "Log level for the service",
		EnvVar: "LOG_LEVEL",
	})
	apiYml := app.String(cli.StringOpt{
		Name:   "api-yml",
		Value:  "./api.yml",
		Desc:   "Location of the API Swagger YML file.",
		EnvVar: "API_YML",
	})

	log := logger.NewUPPLogger("public-annotations-api", *logLevel)

	app.Action = func() {
		log.Infof("public-annotations-api will listen on port: %s, connecting to: %s", *port, *neoURL)
		err := runServer(*neoURL, *port, *cacheDuration, *env, log, apiYml)
		if err != nil {
			log.WithError(err).Error("failed to start public-annotations-api service")
			return
		}
	}

	log.Infof("Application started with args %s", os.Args)
	err := app.Run(os.Args)
	if err != nil {
		log.WithError(err).Error("public-annotations-api could not start!")
		return
	}
}

func runServer(neoURL string, port string, cacheDuration string, env string, log *logger.UPPLogger, apiYml *string) error {
	duration, durationErr := time.ParseDuration(cacheDuration)
	if durationErr != nil {
		return fmt.Errorf("failed to parse cache duration string: %w", durationErr)
	}
	cacheControlHeader := fmt.Sprintf("max-age=%s, public", strconv.FormatFloat(duration.Seconds(), 'f', 0, 64))

	conf := neoutils.ConnectionConfig{
		BatchSize:     1024,
		Transactional: false,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 100,
			},
			Timeout: 1 * time.Minute,
		},
		BackgroundConnect: true,
	}
	db, err := neoutils.Connect(neoURL, &conf, log)
	if err != nil {
		return fmt.Errorf("failed connecting to neo4j: %w", err)
	}

	annotationsDriver := annotations.NewCypherDriver(db, env)
	handlersCtx := annotations.NewHandlerCtx(annotationsDriver, cacheControlHeader, log)
	return routeRequests(port, handlersCtx, apiYml)
}

func routeRequests(port string, hctx *annotations.HandlerCtx, apiYml *string) error {
	// Standard endpoints
	healthCheck := fthealth.TimedHealthCheck{
		HealthCheck: fthealth.HealthCheck{
			SystemCode:  "annotationsapi",
			Name:        "public-annotations-api",
			Description: appDescription,
			Checks: []fthealth.Check{
				annotations.HealthCheck(hctx),
			},
		},
		Timeout: 10 * time.Second,
	}
	http.HandleFunc("/__health", fthealth.Handler(healthCheck))
	http.HandleFunc(status.GTGPath, status.NewGoodToGoHandler(annotations.GoodToGo(hctx)))
	http.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)

	// API specific endpoints
	servicesRouter := mux.NewRouter()

	servicesRouter.HandleFunc("/content/{uuid}/annotations", annotations.GetAnnotations(hctx)).Methods("GET")
	servicesRouter.HandleFunc("/content/{uuid}/annotations", annotations.MethodNotAllowedHandler)
	if apiYml != nil {
		if endpoint, err := apiEndpoint.NewAPIEndpointForFile(*apiYml); err == nil {
			servicesRouter.HandleFunc(apiEndpoint.DefaultPath, endpoint.ServeHTTP).Methods("GET")
		}
	}

	var monitoringRouter http.Handler = servicesRouter
	monitoringRouter = httphandlers.TransactionAwareRequestLoggingHandler(hctx.Log, monitoringRouter)
	monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)

	http.Handle("/", monitoringRouter)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}
