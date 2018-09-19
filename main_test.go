package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"
)

func TestExporter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/metrics/search":
			page := r.URL.Query().Get("p")
			b, err := ioutil.ReadFile(fmt.Sprintf("fixtures/metrics_search_%s.json", page))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(b)
		case "/api/components/search":
			page := r.URL.Query().Get("p")
			b, err := ioutil.ReadFile(fmt.Sprintf("fixtures/components_search_%s.json", page))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(b)
		case "/api/measures/component":
			cID := r.URL.Query().Get("componentId")
			b, err := ioutil.ReadFile(fmt.Sprintf("fixtures/measures_component_%s.json", cID))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(b)
		}
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "./sonarqube_exporter", "-sonarqube.url", ts.URL, "-log.level", "debug")
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	defer cancel()
	go func() {
		err := cmd.Run()
		if err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(1 * time.Second)
	resp, err := http.Get("http://localhost:9344/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	tests := []string{
		`sonarqube_exporter_scrapes_total 1`,
		`sonarqube_up 1`,
		`sonarqube_measures{component_key="identifier:testOne",metric="branch_coverage"} 35.4`,
		`sonarqube_measures{component_key="identifier:testOne",metric="conditions_to_cover"} 794`,
		`sonarqube_measures{component_key="identifier:testTwo",metric="class_complexity"} 4.9`,
	}
	for _, test := range tests {
		if !bytes.Contains(body, []byte(test)) {
			t.Errorf("want metrics to include %q, have:\n%s", test, body)
			t.Errorf("STDOUT:\n%s", string(stdout.Bytes()))
			t.Errorf("STDERR:\n%s", string(stderr.Bytes()))
		}
	}
}
