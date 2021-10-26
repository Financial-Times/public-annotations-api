# Public API for Annotations (public-annotations-api)

[![Circle CI](https://circleci.com/gh/Financial-Times/public-annotations-api.svg?style=shield)](https://circleci.com/gh/Financial-Times/public-annotations-api)[![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/public-annotations-api)](https://goreportcard.com/report/github.com/Financial-Times/public-annotations-api) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/public-annotations-api/badge.svg)](https://coveralls.io/github/Financial-Times/public-annotations-api)
__Provides a public API for Annotations stored in a Neo4J graph database__

## Installation & running locally

```sh
go get -u github.com/Financial-Times/public-annotations-api
cd $GOPATH/src/github.com/Financial-Times/public-annotations-api
go build -mod=readonly
./public-annotations-api
```

Command line options:
```sh
--neo-url            neo4j endpoint URL (env $NEO_URL) (default "bolt://localhost:7687")
--port               Port to listen on (env $PORT) (default "8080")
--env                environment this app is running in (default "local")
--cache-duration     Duration Get requests should be cached for. e.g. 2h45m would set the max-age value to '7440' seconds (env $CACHE_DURATION) (default "30s")
--log-level          Log level for the service (env $LOG_LEVEL) (default "info")
--dbDriverLogLevel   Db's driver logging level (DEBUG, INFO, WARN, ERROR) (env $DB_DRIVER_LOG_LEVEL) (default "WARN")
--api-yml            Location of the API Swagger YML file. (env $API_YML) (default "./api.yml")
```

* `curl http://localhost:8080/content/143ba45c-2fb3-35bc-b227-a6ed80b5c517/annotations | json_pp`
* Or using [httpie](https://github.com/jkbrzt/httpie) `http GET http://localhost:8080/content/143ba45c-2fb3-35bc-b227-a6ed80b5c517/annotations`

## Testing

* Run unit tests only: `go test -race ./...`
* Run unit and integration tests:

    In order to execute the integration tests you must provide GITHUB_USERNAME and GITHUB_TOKEN values, because the service is depending on internal repositories.
    ```sh
    GITHUB_USERNAME=<username> GITHUB_TOKEN=<token> \
    docker-compose -f docker-compose-tests.yml up -d --build && \
    docker logs -f test-runner && \
    docker-compose -f docker-compose-tests.yml down
    ```

## Build & deployment

Continuously built by CircleCI. The docker image of the service is built by Dockerhub based on the git release tag.
To prepare a new git release, go to the repo page on GitHub and create a new release.

* Cluster deployment:  [public-annotations-api](https://upp-k8s-jenkins.in.ft.com/job/k8s-deployment/job/apps-deployment/job/public-annotations-api-auto-deploy/)
* CI provided by CircleCI: [public-annotations-api](https://circleci.com/gh/Financial-Times/public-annotations-api)
* Code coverage provided by Coverall: [public-annotations-api](https://coveralls.io/github/Financial-Times/public-annotations-api)

## API definition

Based on the following [google doc](https://docs.google.com/a/ft.com/document/d/1kQH3tk1GhXnupHKdDhkDE5UyJIHm2ssWXW3zjs3g2h8/edit?usp=sharing)

### GET content/{uuid}/annotations endpoint

Returns all annotations for a given uuid of a piece of content in json format.

*Please note* that

* the `public-annotations-api` will return more brands than the ones the article has been annotated with.
This is because it will return also the parent of the brands from any brands annotations.
If those brands have parents, then they too will be brought into the result.

* the `public-annotations-api` uses annotations lifecycle to determine which annotations are returned. If curated (tag-me) annotations (lifecycle pac) for a piece of content exist, they will be returned combined with V2 annotations by default, other non-pac lifecycle annotations are omitted.
If there are no pac lifecycle annotations, non-pac annotations will be returned. The filtering described in the next paragraph relates to non-pac annotations. Additional filtering by annotations lifecycle could be applied using the optional "lifecycle" query parameter.

* the `public-annotations-api` will filter out less important annotations if a more important annotation is also present for the same concept.  
_For example_, if a piece of content is annotated with a concept with "About", "Major Mentions" and "Mentions" relationships
only the annotation with "About" relationship will be returned.
Similarly if a piece of content is annotated with a Concept "Is Classified By" and "Is Primarily Classified By"
only the annotation with "Is Primarily Classified By" relationship will be returned.

## Admin endpoints

* Healthchecks: [http://localhost:8080/__health](http://localhost:8080/__health)  
* Build Info: [http://localhost:8080/__build-info](http://localhost:8080/__build-info)  
* GTG: [http://localhost:8080/__gtg](http://localhost:8080/__gtg)

### Logging

Logging requires an env app parameter: for all environments other than local, logs are written to file. When running locally logging is written to console (if you want to log locally to file you need to pass in an env parameter that is != local).

NOTE: <http://localhost:8080/__gtg> end point is not logged as it is called every second from varnish and this information is not needed in logs/splunk
