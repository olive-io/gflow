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

package dao

import (
	"fmt"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/pkg/dbutil"
)

type EndpointDao struct {
	*InnerDao[types.Endpoint]
}

func NewEndpointDao(db *dbutil.DB) (*EndpointDao, error) {
	inner, err := newDao(db, types.Endpoint{})
	if err != nil {
		return nil, fmt.Errorf("auto migrate endpoint models: %w", err)
	}
	dao := &EndpointDao{
		InnerDao: inner,
	}

	return dao, nil
}
