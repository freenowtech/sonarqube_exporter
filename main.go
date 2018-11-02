package main

import (
	"flag"
	"net/http"
	"regexp"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const (
	namespace = "sonarqube"
)

var (
	httpAddress         = flag.String("http.address", ":9344", "Address to listen on")
	httpTelemetryPath   = flag.String("http.telemetry-path", "/metrics", "Path under which the exporter exposes its metrics")
	logLevel            = flag.String("log.level", "ERROR", "Log level")
	sonarqubePassword   = flag.String("sonarqube.password", "", "Password to use for authentication")
	sonarqubeURL        = flag.String("sonarqube.url", "http://localhost:8080", "URL of Sonarqube")
	sonarqubeUsername   = flag.String("sonarqube.username", "", "Username to use for authentication")
	projectFilterRegex  = flag.String("sonarqube.project-filter", ".*", "Regexp to limit the number of projects to scrape. Applied to the key of each project.")
	accpetedMetricTypes = map[string]struct{}{"INT": {}, "PERCENT": {}, "FLOAT": {}, "DATA": {}, "RATING": {}, "LEVEL": {}}
)

type exporter struct {
	client         *apiClient
	projectFilter  *regexp.Regexp
	projectMetrics *prometheus.GaugeVec
	up             prometheus.Gauge
	totalScrapes   prometheus.Counter
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	e.scrape()
	e.projectMetrics.Collect(ch)
	ch <- e.totalScrapes
	ch <- e.up
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	e.projectMetrics.Describe(ch)
	ch <- e.totalScrapes.Desc()
	ch <- e.up.Desc()
}

func (e *exporter) scrape() {
	e.totalScrapes.Inc()
	allMetrics, err := e.client.findAllMetrics()
	if err != nil {
		log.Errorf("Finding all metrics: %s ", err)
		e.up.Set(0)
		return
	}

	log.Debugf("Found %d metrics", len(allMetrics))
	metricKeys := []string{}
	for _, m := range allMetrics {
		if _, exists := accpetedMetricTypes[m.Type]; exists {
			metricKeys = append(metricKeys, m.Key)
		}
	}

	allProjects, err := e.client.findAllProjects()
	if err != nil {
		log.Errorf("Finding all projects: %s ", err)
		e.up.Set(0)
		return
	}

	log.Debugf("Found %d projects", len(allProjects))
	for _, p := range allProjects {
		if !e.projectFilter.MatchString(p.Key) {
			continue
		}

		log.Debugf("Requesting metrics for %s...", p.ID)
		r, err := e.client.findMeasuresForComponent(p.ID, metricKeys)
		if err != nil {
			log.Errorf("Finding measures for component '%s': %s", p.ID, err)
			e.up.Set(0)
			return
		}

		for _, measure := range r.Component.Measures {
			var measureFloat float64
			if metric, exists := dataMetricsValues[measure.Metric]; exists {
				var valueExists bool
				measureFloat, valueExists = metric[measure.Value]
				if !valueExists {
					continue
				}
			} else if measure.Value != "" {
				if measureFloat, err = strconv.ParseFloat(measure.Value, 64); err != nil {
					log.Debugf("Value of measure '%s' could not be parsed: %s", measure.Metric, err)
					continue
				}
			}

			e.projectMetrics.WithLabelValues(p.Key, measure.Metric).Set(measureFloat)
		}
	}

	e.up.Set(1)
}

func mustInitLogging(lvl string) {
	level, err := log.ParseLevel(lvl)
	if err != nil {
		log.Fatal(err)
	}

	log.SetLevel(level)
}

func main() {
	flag.Parse()
	mustInitLogging(*logLevel)
	e := &exporter{
		client: newAPIClient(nil, *sonarqubeUsername, *sonarqubePassword, *sonarqubeURL),
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Was the last scrape of Sonarqube successful.",
		}),
		projectFilter: regexp.MustCompile(*projectFilterRegex),
		projectMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "measures",
				Help:        "A measure of a project in Sonarqube.",
				ConstLabels: nil,
			},
			[]string{"component_key", "metric"},
		),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrapes_total",
			Help:      "Total scrapes of the Sonarqube exporter.",
		}),
	}
	prometheus.MustRegister(e)
	http.Handle(*httpTelemetryPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
<head><title>Sonarqube Exporter</title></head>
<body>
<h1>Sonarqube Exporter</h1>
<p><a href='` + *httpTelemetryPath + `'>Metrics</a></p>
</body>
</html>`))
	})
	log.Fatal(http.ListenAndServe(*httpAddress, nil))
}
