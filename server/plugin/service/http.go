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

package service

import (
	"bytes"
	"context"
	"fmt"
	"time"

	json "github.com/bytedance/sonic"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/plugins"
	"github.com/olive-io/gflow/server/plugin/sdk"
)

var _ plugins.Plugin = (*httpPlugin)(nil)

type httpPlugin struct {
	lg       *otelzap.Logger
	flowType types.FlowNodeType
	typ      string
}

func newHTTPPlugin(lg *otelzap.Logger, flowType types.FlowNodeType) *httpPlugin {
	hp := &httpPlugin{
		lg:       lg,
		flowType: flowType,
		typ:      plugins.HTTPPlugin,
	}

	return hp
}

func (hp *httpPlugin) Name() string { return hp.typ }

func (hp *httpPlugin) Do(ctx context.Context, req *plugins.Request, opts ...plugins.DoOption) (*plugins.Response, error) {
	lg := hp.lg
	options := plugins.NewDoOptions(opts...)

	if options.Stage > types.CallTaskStage_Commit {
		return &plugins.Response{}, nil
	}

	hc := &sdk.HTTPConfig{
		Timeout: time.Duration(options.Timeout) * time.Millisecond,
	}

	headers := req.Headers
	var ok bool
	hc.URL, ok = req.Headers["url"]
	if !ok {
		return nil, fmt.Errorf("missing url property")
	}

	hc.Username = headers["username"]
	hc.Password = headers["password"]
	hc.Token = headers["token"]

	var method string
	method, ok = headers["method"]
	if !ok {
		return nil, fmt.Errorf("missing method property")
	}

	in := map[string]any{}
	for name, pt := range req.Properties {
		in[name] = pt.ToSchemaValue().ValueFor()
	}
	data, _ := json.Marshal(in)
	body := bytes.NewBuffer(data)

	client := sdk.NewHTTPClient(lg, hc)

	presp := &plugins.Response{}

	target := map[string]any{}
	hresp, err := client.Do(ctx, method, headers, body, &target)
	if err != nil {
		presp.Error = err.Error()
		return nil, err
	}

	results := make(map[string]*types.Value)
	results["code"] = types.NewValue(hresp.StatusCode)
	if len(hresp.Message) != 0 {
		results["message"] = types.NewValue(hresp.Message)
	}
	for k, v := range target {
		results[k] = types.NewValue(v)
	}
	presp.Results = results

	return presp, nil
}
