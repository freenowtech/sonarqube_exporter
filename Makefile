GOLANG_VERSION = 1.11.0
VERSION ?= master

test:
	go build
	go test

build_darwin:
	docker run --rm -v "$(PWD):/go/src/github.com/mytaxi/sonarqube_exporter" -w "/go/src/github.com/mytaxi/sonarqube_exporter" -e "GOARCH=amd64" -e "GOOS=darwin" golang:$(GOLANG_VERSION) go build -o sonarqube_exporter-${VERSION}.darwin-amd64

build_linux:
	docker run --rm -v "$(PWD):/go/src/github.com/mytaxi/sonarqube_exporter" -w "/go/src/github.com/mytaxi/sonarqube_exporter" -e "GOARCH=amd64" -e "GOOS=linux" golang:$(GOLANG_VERSION) go build -o sonarqube_exporter-${VERSION}.linux-amd64

build_docker:
	docker build -t mytaxi/sonarqube_exporter:${VERSION} .
