package main

import (
	"net/http"
	"os"

	"github.com/andrejbranch/avalanche-query/query"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	queryInterval  = kingpin.Flag("query-interval", "run query every {interval} seconds.").Default("5").Int()
	cortexFrontend = kingpin.Flag("frontend", "cortex frontend dns").Default("localhost").String()
	runType        = kingpin.Flag("run-type", "run type (instant | two-day | four-hour").Default("instant").String()
)

func main() {
	kingpin.Version("0.3")
	kingpin.CommandLine.Help = "single vs multi tenant avalanche query latency comparison"
	kingpin.Parse()
	logger := log.With(log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)), "caller", log.DefaultCaller)
	query.RegisterMetrics()
	query.Run(*queryInterval, *cortexFrontend)
	serveMetrics(log.With(logger, "component", "metrics"))
}

func serveMetrics(logger log.Logger) {
	server := http.Server{
		Addr: ":9001",
	}
	http.Handle("/metrics", promhttp.Handler())

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		level.Error(logger).Log("msg", "Error exposing metrics ", "err", err)
	}
}
