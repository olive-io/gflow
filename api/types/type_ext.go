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

package types

import (
	"path"
)

func (in *Endpoint) URL() string {
	url := path.Join(in.Type.String())
	if in.Kind != "" {
		url = path.Join(url, in.Kind)
	}
	if in.Name != "" {
		url = path.Join(url, in.Name)
	}
	return url
}

func (in *Value) DeepCopyInto(out *Value) {
	*out = *in
}

func (in *Value) DeepCopy() *Value {
	if in == nil {
		return nil
	}
	out := new(Value)
	in.DeepCopyInto(out)
	return out
}

func (in *ProcessContext) DeepCopyInto(out *ProcessContext) {
	*out = *in
	for k, v := range in.DataObjects {
		out.DataObjects[k] = v.DeepCopy()
	}
	for k, v := range in.Variables {
		out.Variables[k] = v.DeepCopy()
	}
}

func (in *ProcessContext) DeepCopy() *ProcessContext {
	if in == nil {
		return nil
	}
	out := new(ProcessContext)
	in.DeepCopyInto(out)
	return out
}

func (in *Process) DeepCopyInto(out *Process) {
	*out = *in
	if in.Context != nil {
		in, out := in.Context, out.Context
		in.DeepCopyInto(out)
	}
}

func (in *Process) DeepCopy() *Process {
	if in == nil {
		return nil
	}
	out := new(Process)
	in.DeepCopyInto(out)
	return out
}

func (in *FlowNode) DeepCopyInto(out *FlowNode) {
	*out = *in
	if in.Headers != nil {
		for k, v := range in.Headers {
			out.Headers[k] = v
		}
	}
	if in.Properties != nil {
		for k, v := range in.Properties {
			out.Properties[k] = v.DeepCopy()
		}
	}
	if in.DataObjects != nil {
		for k, v := range in.DataObjects {
			out.DataObjects[k] = v.DeepCopy()
		}
	}
}

func (in *FlowNode) DeepCopy() *FlowNode {
	if in == nil {
		return nil
	}
	out := new(FlowNode)
	in.DeepCopyInto(out)
	return out
}

func (in *Runner) DeepCopyInto(out *Runner) {
	*out = *in
	if in.Metadata != nil {
		for k, v := range in.Metadata {
			out.Metadata[k] = v
		}
	}
	if in.Features != nil {
		for k, v := range in.Features {
			out.Features[k] = v
		}
	}
}

func (in *Runner) DeepCopy() *Runner {
	if in == nil {
		return nil
	}
	out := new(Runner)
	in.DeepCopyInto(out)
	return out
}

func (in *RunnerStat) DeepCopyInto(out *RunnerStat) {
	*out = *in
}

func (in *RunnerStat) DeepCopy() *RunnerStat {
	if in == nil {
		return nil
	}
	out := new(RunnerStat)
	in.DeepCopyInto(out)
	return out
}
