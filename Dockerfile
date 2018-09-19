FROM golang:1.11.0
WORKDIR /go/src/github.com/mytaxi/sonarqube_exporter/
COPY . .
RUN go build

FROM debian:stable-slim
COPY --from=0 /go/src/github.com/mytaxi/sonarqube_exporter/sonarqube_exporter /sonarqube_exporter
USER nobody
ENTRYPOINT ["/sonarqube_exporter"]
