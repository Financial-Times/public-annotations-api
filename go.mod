module github.com/Financial-Times/public-annotations-api/v3

go 1.15

require (
	github.com/Financial-Times/annotations-rw-neo4j/v3 v3.2.0
	github.com/Financial-Times/api-endpoint v0.0.0-20170713111258-802a63542ff0 // indirect
	github.com/Financial-Times/base-ft-rw-app-go v0.0.0-20180522140206-1ea8a13e1f37
	github.com/Financial-Times/concepts-rw-neo4j v1.28.0-dev-identifiers-rc1
	github.com/Financial-Times/content-rw-neo4j/v3 v3.3.0
	github.com/Financial-Times/go-fthealth v0.0.0-20181009114238-ca83ad65381f
	github.com/Financial-Times/go-logger/v2 v2.0.1
	github.com/Financial-Times/http-handlers-go/v2 v2.3.0
	github.com/Financial-Times/neo-model-utils-go v1.0.0
	github.com/Financial-Times/neo-utils-go/v2 v2.0.0
	github.com/Financial-Times/service-status-go v0.0.0-20160323111542-3f5199736a3d
	github.com/gorilla/mux v1.7.3
	github.com/jawher/mow.cli v1.1.0
	github.com/jmcvetta/neoism v1.3.2-0.20160701082253-9d29cb10be18
	github.com/joho/godotenv v1.3.0
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563
	github.com/stretchr/testify v1.6.1
)

replace gopkg.in/stretchr/testify.v1 => github.com/stretchr/testify v1.4.0
