# Public API for Annotations (public-annotations-api)
__Provides a public API for Annotations stored in a Neo4J graph database__

## Build & deployment etc:
_NB You will need to tag a commit in order to build, since the UI asks for a tag to build / deploy_
* [Jenkins view](http://ftjen10085-lvpr-uk-p:8181/view/JOBS-public-annotations-api/)
* [Build and publish to forge](http://ftjen10085-lvpr-uk-p:8181/job/public-annotations-api-build)
* [Deploy to test or production](http://ftjen10085-lvpr-uk-p:8181/job/public-annotations-api-deploy)


## Installation & running locally
* `go get -u github.com/Financial-Times/public-annotations-api`
* `cd $GOPATH/src/github.com/Financial-Times/public-annotations-api`
* `go test ./...`
* `go install`
* `$GOPATH/bin/public-annotations-api --neo-url={neo4jUrl} --port={port} --log-level={DEBUG|INFO|WARN|ERROR}--cache-duration{e.g. 22h10m3s}`
_Optional arguments are:
--neo-url defaults to http://localhost:7474/db/data, which is the out of box url for a local neo4j instance.
--port defaults to 8080.
--cache-duration defaults to 1 hour._
* `curl http://localhost:8080/content/143ba45c-2fb3-35bc-b227-a6ed80b5c517/annotations | json_pp`
Or using [httpie](https://github.com/jkbrzt/httpie)
* `http GET http://localhost:8080/content/143ba45c-2fb3-35bc-b227-a6ed80b5c517/annotations`

## API definition
Based on the following [google doc](https://docs.google.com/a/ft.com/document/d/1kQH3tk1GhXnupHKdDhkDE5UyJIHm2ssWXW3zjs3g2h8/edit?usp=sharing)

## Healthchecks
Healthchecks: [http://localhost:8080/__health](http://localhost:8080/__health)

### Logging
The application uses logrus, the logfile is initialised in main.go.

Logging requires an env app parameter: for all environments other than local, logs are written to file. When running locally logging
is written to console (if you want to log locally to file you need to pass in an env parameter that is != local).

NOTE: http://localhost:8080/__gtg end point is not logged as it is called every second from varnish and this information is not needed in logs/splunk
