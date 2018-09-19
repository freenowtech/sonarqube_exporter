# sonarqube_exporter

An exporter of metrics for each project in Sonarqube.

## Usage

```
./sonarqube_exporter -h
Usage of ./sonarqube_exporter:
  -http.address string
    	Address to listen on (default ":9344")
  -http.telemetry-path string
    	Path under which the exporter exposes its metrics (default "/metrics")
  -log.level string
    	Log level (default "ERROR")
  -sonarqube.password string
    	Password to use for authentication
  -sonarqube.project-filter string
    	Regexp to limit the number of projects to scrape. Applied to the key of each project. (default ".*")
  -sonarqube.url string
    	URL of Sonarqube (default "http://localhost:8080")
  -sonarqube.username string
    	Username to use for authentication
```

## Development

### Building

```
$ go build
```

### Sonarqube API documentation

API docs can be found in the local installation of Sonarqube at `http://<sonarqube>/web_api`.
