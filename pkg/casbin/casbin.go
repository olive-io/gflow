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

package casbin

import (
	"errors"
	"fmt"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

type AuthRule struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Ptype string `gorm:"size:100;uniqueIndex:unique_index"`
	V0    string `gorm:"size:100;uniqueIndex:unique_index"`
	V1    string `gorm:"size:100;uniqueIndex:unique_index"`
	V2    string `gorm:"size:100;uniqueIndex:unique_index"`
	V3    string `gorm:"size:100;uniqueIndex:unique_index"`
	V4    string `gorm:"size:100;uniqueIndex:unique_index"`
	V5    string `gorm:"size:100;uniqueIndex:unique_index"`
}

func (r *AuthRule) TableName() string {
	return "authority_rule"
}

func CreateAdapter(db *gorm.DB, model string) (*casbin.Enforcer, error) {
	if model == "" {
		return nil, errors.New("missing model configuration file")
	}

	ruleTable := &AuthRule{}
	adapter, err := gormadapter.NewAdapterByDBWithCustomTable(db, ruleTable, ruleTable.TableName())
	if err != nil {
		return nil, fmt.Errorf("create adapter: %w", err)
	}

	enforcer, err := casbin.NewEnforcer(model, adapter)
	if err != nil {
		return nil, fmt.Errorf("create casbin enforcer: %w", err)
	}

	if err = enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("load policy: %w", err)
	}

	return enforcer, nil
}
