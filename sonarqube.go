package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type metricsSearchResponse struct {
	Metrics  []*metric `json:"metrics"`
	Total    int       `json:"total"`
	Page     int       `json:"p"`
	PageSize int       `json:"ps"`
}

type metric struct {
	Key  string `json:"key"`
	Type string `json:"type"`
}

type componentMeasuresResponse struct {
	Component *componentMeasureItem
}

type componentMeasureItem struct {
	ID       string
	Key      string
	Measures []*measure
}

type componentSearchResponse struct {
	Paging     *componentSearchPaging
	Components []*componentSearchItem
}

type componentSearchPaging struct {
	PageIndex int
	PageSize  int
	Total     int
}

type componentSearchItem struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

type measure struct {
	Metric string
	Value  string
}

type apiClient struct {
	client       *http.Client
	password     string
	sonarqubeURL string
	username     string
}

func (a *apiClient) findAllMetrics() ([]*metric, error) {
	page := 1
	pageSize := 100
	metrics := []*metric{}
	var total int
	for {
		msr := &metricsSearchResponse{}
		url := fmt.Sprintf("%s/api/metrics/search?ps=%d&p=%d", a.sonarqubeURL, pageSize, page)
		err := a.requestJSON(url, msr)
		if err != nil {
			return nil, err
		}

		pageSize = msr.PageSize
		total = msr.Total
		metrics = append(metrics, msr.Metrics...)

		if (page * pageSize) >= total {
			return metrics, nil
		}

		page = page + 1
	}
}

func (a *apiClient) findAllProjects() ([]*componentSearchItem, error) {
	pageIndex := 1
	pageSize := 100
	components := []*componentSearchItem{}
	for {
		csr := &componentSearchResponse{}
		url := fmt.Sprintf("%s/api/components/search?ps=%d&p=%d&qualifiers=TRK", a.sonarqubeURL, pageSize, pageIndex)
		err := a.requestJSON(url, csr)
		if err != nil {
			return nil, err
		}

		pageSize = csr.Paging.PageSize
		components = append(components, csr.Components...)
		if (pageIndex * pageSize) >= csr.Paging.Total {
			return components, nil
		}

		pageIndex = csr.Paging.PageIndex + 1
	}
}

func (a *apiClient) findMeasuresForComponent(componentID string, metricKeys []string) (*componentMeasuresResponse, error) {
	url := fmt.Sprintf("%s/api/measures/component?componentId=%s&metricKeys=%s", a.sonarqubeURL, componentID, strings.Join(metricKeys, ","))
	cmr := &componentMeasuresResponse{}
	err := a.requestJSON(url, cmr)
	if err != nil {
		return nil, err
	}

	return cmr, nil
}

func (a *apiClient) requestJSON(url string, v interface{}) error {
	log.Debugf("Seding request to '%s'", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if a.password != "" && a.username != "" {
		log.Debug("Setting basic auth")
		req.SetBasicAuth(a.username, a.password)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, v)
	if err != nil {
		return err
	}

	return nil
}

func newAPIClient(h *http.Client, username string, password string, url string) *apiClient {
	if h == nil {
		h = &http.Client{Timeout: 1 * time.Second}
	}

	return &apiClient{
		client:       h,
		password:     password,
		sonarqubeURL: url,
		username:     username,
	}
}
