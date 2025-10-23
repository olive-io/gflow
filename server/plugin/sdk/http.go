/*
Copyright 2025 The gflow Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sdk

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

type HttpResponse struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

type HTTPConfig struct {
	URL      string
	Username string
	Password string
	Token    string
	Timeout  time.Duration
}

type HTTPClient struct {
	cfg *HTTPConfig
	lg  *otelzap.Logger
}

func NewHTTPClient(lg *otelzap.Logger, cfg *HTTPConfig) *HTTPClient {
	hc := &HTTPClient{
		cfg: cfg,
		lg:  lg,
	}
	return hc
}

func (hc *HTTPClient) Do(ctx context.Context, method string, headers map[string]string, body io.Reader, target any) (*HttpResponse, error) {
	lg := hc.lg
	cfg := hc.cfg

	timeout := cfg.Timeout
	transport := &http.Transport{}
	conn := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	var url *urlpkg.URL
	url, err := urlpkg.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %v", err)
	}

	if url.Scheme == "https" {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	hr, err := http.NewRequestWithContext(ctx, method, url.String(), body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	var ct string
	header := http.Header{}
	for name, value := range headers {
		if name == "Content-Type" {
			ct = value
			continue
		}
		header.Set(name, value)
	}
	if ct == "" {
		ct = "application/json"
	}
	header.Set("Content-Type", ct)

	hr.Header = header

	lg.Debug("http request",
		zap.String("method", method),
		zap.String("url", url.String()))

	resp, err := conn.Do(hr)
	if err != nil {
		return nil, fmt.Errorf("request to %s: %w", url.String(), err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	hresp := &HttpResponse{
		StatusCode: resp.StatusCode,
	}

	if resp.StatusCode < 400 {
		err = json.Unmarshal(data, target)
		if err != nil {
			return nil, fmt.Errorf("unmarshal response: %w", err)
		}
	} else {
		hresp.Message = string(data)
	}

	return hresp, nil
}
