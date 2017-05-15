*DECOMISSIONED*
See [Concepts RW Neo4j](https://github.com/Financial-Times/concepts-rw-neo4j) instead

# Locations Reader/Writer for Neo4j (locations-rw-neo4j)
[![Circle CI](https://circleci.com/gh/Financial-Times/locations-rw-neo4j.svg?style=shield)](https://circleci.com/gh/Financial-Times/locations-rw-neo4j)[![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/locations-rw-neo4j)](https://goreportcard.com/report/github.com/Financial-Times/locations-rw-neo4j) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/locations-rw-neo4j/badge.svg)](https://coveralls.io/github/Financial-Times/locations-rw-neo4j)
__An API for reading/writing locations into Neo4j. Expects the locations json supplied to be in the format that comes out of the locations transformer.__

## Installation

For the first time:

`go get github.com/Financial-Times/locations-rw-neo4j`

or update:

`go get -u github.com/Financial-Times/locations-rw-neo4j`

## Running

```
export|set PORT=8080
export|set NEO_URL={neo4jUrl}
export|set BATCH_SIZE=50
export|set GRAPHITE_TCP_ADDRESS=graphite.ft.com:2003
export|set GRAPHITE_PREFIX=coco.{env}.services.locations-rw-neo4j.{instanceNumber}
export|set LOG_METRICS=true
$GOPATH/bin/locations-rw-neo4j
```

With Docker:

`docker build -t coco/locations-rw-neo4j .`

`docker run -ti --env NEO_URL=<base url> coco/locations-rw-neo4j`


All arguments are optional, they default to a local Neo4j install on the default port (7474), application running on port 8080, batchSize of 1024, graphiteTCPAddress of "" (meaning metrics won't be written to Graphite), graphitePrefix of "" and logMetrics false.

NB: the default batchSize is much higher than the throughput the instance data ingester currently can cope with.

## Endpoints
/locations/{uuid}
### PUT
"The only mandatory fields are the uuid, the prefLabel and the alternativeIdentifier uuids (because the uuid is also listed in the alternativeIdentifier uuids list)."

Every request results in an attempt to update that location: unlike with GraphDB there is no check on whether the location already exists and whether there are any changes between what's there and what's being written. We just do a MERGE which is Neo4j for create if not there, update if it is there.

A successful PUT results in 200.

We run queries in batches. If a batch fails, all failing requests will get a 500 server error response.

Invalid json body input, or uuids that don't match between the path and the body will result in a 400 bad request response.

Example:
`curl -XPUT -H "X-Request-Id: 123" -H "Content-Type: application/json" localhost:8080/locations/bba39990-c78d-3629-ae83-808c333c6dbc --data '{"uuid":"bba39990-c78d-3629-ae83-808c333c6dbc","prefLabel":"Metals Markets", "alternativeIdentifiers":{"TME":["MTE3-U3ViamVjdHM="],"uuids": ["bba39990-c78d-3629-ae83-808c333c6dbc","6a2a0170-6afa-4bcc-b427-430268d2ac50"]},"type":"Location"}'`

The type field is not currently validated - instead, the Locations Writer writes type Location and its parent types (Location, Concept and Thing) as labels for the Location

### GET
The internal read should return what got written

If not found, you'll get a 404 response.

Empty fields are omitted from the response.
`curl -H "X-Request-Id: 123" localhost:8080/locations/bba39990-c78d-3629-ae83-808c333c6dbc`

### DELETE
Will return 204 if successful, 404 if not found
`curl -XDELETE -H "X-Request-Id: 123" localhost:8080/locations/bba39990-c78d-3629-ae83-808c333c6dbc`

### Admin endpoints
Healthchecks: [http://localhost:8080/__health](http://localhost:8080/__health)

Ping: [http://localhost:8080/ping](http://localhost:8080/ping) or [http://localhost:8080/__ping](http://localhost:8080/__ping)
