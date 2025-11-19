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
	"strings"
)

var (
	SystemPolicies = []string{
		"/v1/users/self (GET)|(PATCH)",

		"/v1/runners (GET)",

		"/v1/endpoints (GET)",
	}

	OperatorPolicies = []string{
		"/v1/users/self (GET)|(PATCH)",

		"/v1/endpoints (GET)",
	}
)

func (p *Policy) Rules() []string {
	return []string{p.Subject, p.Object, p.Action}
}

func ParsePolicy(rules []string) *Policy {
	p := &Policy{}
	if len(rules) > 0 {
		p.Subject = rules[0]
	}
	if len(rules) > 1 {
		p.Object = rules[1]
	}
	if len(rules) > 2 {
		p.Action = rules[2]
	}
	return p
}

func ListInitPolicies() []Policy {
	policies := make([]Policy, 0)

	sub := DefaultSystemRole
	for _, p := range SystemPolicies {
		policy := ParsePolicy(append([]string{sub}, strings.Split(p, " ")...))
		policies = append(policies, *policy)
	}

	sub = DefaultOperatorRole
	for _, p := range OperatorPolicies {
		policy := ParsePolicy(append([]string{sub}, strings.Split(p, " ")...))
		policies = append(policies, *policy)
	}

	return policies
}
