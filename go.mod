module github.com/Financial-Times/public-annotations-api/v3

go 1.22

toolchain go1.22.2

require (
	github.com/Financial-Times/annotations-rw-neo4j/v4 v4.9.1
	github.com/Financial-Times/api-endpoint v1.0.0
	github.com/Financial-Times/base-ft-rw-app-go/v2 v2.0.0
	github.com/Financial-Times/cm-graph-ontology/v2 v2.0.12
	github.com/Financial-Times/cm-neo4j-driver v1.1.1
	github.com/Financial-Times/concepts-rw-neo4j v1.35.9
	github.com/Financial-Times/content-rw-neo4j/v3 v3.5.3
	github.com/Financial-Times/go-fthealth v0.6.2
	github.com/Financial-Times/go-logger v0.0.0-20180323124113-febee6537e90
	github.com/Financial-Times/go-logger/v2 v2.0.1
	github.com/Financial-Times/http-handlers-go/v2 v2.3.0
	github.com/Financial-Times/service-status-go v0.3.0
	github.com/gorilla/mux v1.8.1
	github.com/jawher/mow.cli v1.2.0
	github.com/joho/godotenv v1.3.0
	github.com/neo4j/neo4j-go-driver/v4 v4.4.7
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/Financial-Times/cm-annotations-ontology v1.0.4 // indirect
	github.com/Financial-Times/cm-graph-ontology v1.2.0 // indirect
	github.com/Financial-Times/http-handlers-go v0.0.0-20180517120644-2c20324ab887 // indirect
	github.com/Financial-Times/transactionid-utils-go v1.0.0 // indirect
	github.com/Financial-Times/up-rw-app-api-go v0.0.0-20210202155002-307a978447bd // indirect
	github.com/cyberdelia/go-metrics-graphite v0.0.0-20161219230853-39f87cc3b432 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dchest/uniuri v1.2.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/r3labs/diff/v3 v3.0.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	golang.org/x/exp v0.0.0-20221126150942-6ab00d035af9 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace gopkg.in/stretchr/testify.v1 => github.com/stretchr/testify v1.4.0
