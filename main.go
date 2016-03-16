package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"

	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/Financial-Times/locations-rw-neo4j/locations"
	log "github.com/Sirupsen/logrus"
	"github.com/jawher/mow.cli"
	"github.com/jmcvetta/neoism"
)

func init() {
	log.SetFormatter(new(log.JSONFormatter))
}

func main() {
	app := cli.App("locations-rw-neo4j", "A RESTful API for managing Locations in neo4j")
	neoURL := app.String(cli.StringOpt{
		Name:   "neo-url",
		Value:  "http://localhost:7474/db/data",
		Desc:   "neo4j endpoint URL",
		EnvVar: "NEO_URL",
	})
	graphiteTCPAddress := app.String(cli.StringOpt{
		Name:   "graphite-tcp-address",
		Value:  "",
		Desc:   "Graphite TCP address, e.g. graphite.ft.com:2003. Leave as default if you do NOT want to output to graphite (e.g. if running locally)",
		EnvVar: "GRAPHITE_TCP_ADDRESS",
	})
	graphitePrefix := app.String(cli.StringOpt{
		Name:   "graphite-prefix",
		Value:  "",
		Desc:   "Prefix to use. Should start with content, include the environment, and the host name. e.g. coco.pre-prod.locations-rw-neo4j.1",
		EnvVar: "GRAPHITE_PREFIX",
	})
	port := app.Int(cli.IntOpt{
		Name:   "port",
		Value:  8080,
		Desc:   "Port to listen on",
		EnvVar: "PORT",
	})
	batchSize := app.Int(cli.IntOpt{
		Name:   "batch-size",
		Value:  1024,
		Desc:   "Maximum number of statements to execute per batch",
		EnvVar: "BATCH_SIZE",
	})
	logMetrics := app.Bool(cli.BoolOpt{
		Name:   "log-metrics",
		Value:  false,
		Desc:   "Whether to log metrics. Set to true if running locally and you want metrics output",
		EnvVar: "LOG_METRICS",
	})

	app.Action = func() {
		db, err := neoism.Connect(*neoURL)
		if err != nil {
			log.Errorf("Could not connect to neo4j, error=[%s]\n", err)
		}

		batchRunner := neoutils.NewBatchCypherRunner(neoutils.StringerDb{db}, *batchSize)
		locationsDriver := locations.NewCypherLocationsService(batchRunner, db)

		baseftrwapp.OutputMetricsIfRequired(*graphiteTCPAddress, *graphitePrefix, *logMetrics)

		endpoints := map[string]baseftrwapp.Service{
			"locations": locationsDriver,
		}

		var checks []v1a.Check
		for _, e := range endpoints {
			checks = append(checks, makeCheck(e, batchRunner))
		}

		baseftrwapp.RunServer(endpoints,
			v1a.Handler("ft-locations_rw_neo4j ServiceModule", "Writes 'locations' to Neo4j, usually as part of a bulk upload done on a schedule", checks...),
			*port, "locations-rw-neo4j", "local")
	}

	app.Run(os.Args)
}

func makeCheck(service baseftrwapp.Service, cr neoutils.CypherRunner) v1a.Check {
	return v1a.Check{
		BusinessImpact:   "Cannot read/write locations via this writer",
		Name:             "Check connectivity to Neo4j - neoUrl is a parameter in hieradata for this service",
		PanicGuide:       "TODO - write panic guide",
		Severity:         1,
		TechnicalSummary: fmt.Sprintf("Cannot connect to Neo4j instance %s with at least one location loaded in it", cr),
		Checker:          func() (string, error) { return "", service.Check() },
	}
}
