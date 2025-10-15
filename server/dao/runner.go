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
	"context"
	"fmt"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/pkg/dbutil"
)

type RunnerDao struct {
	*InnerDao[types.Runner]
}

func NewRunnerDao(db *dbutil.DB) (*RunnerDao, error) {
	inner, err := newDao(db, types.Runner{})
	if err != nil {
		return nil, fmt.Errorf("auto migrate runner models: %w", err)
	}
	dao := &RunnerDao{
		InnerDao: inner,
	}

	return dao, nil
}

func (dao *RunnerDao) Find(ctx context.Context) ([]*types.Runner, error) {
	tx := dao.db.NewSession(ctx).Model(&types.Runner{})

	runners := make([]*types.Runner, 0)
	err := tx.Order("id desc").Find(&runners).Error
	if err != nil {
		return nil, err
	}

	return runners, nil
}

func (dao *RunnerDao) Get(ctx context.Context, id int64, uid string) (*types.Runner, error) {
	tx := dao.db.NewSession(ctx).Model(&types.Runner{})

	if id != 0 {
		tx = tx.Where("id = ?", id)
	}
	if uid != "" {
		tx = tx.Where("uid = ?", uid)
	}

	var runner types.Runner
	err := tx.First(&runner).Error
	if err != nil {
		return nil, err
	}
	return &runner, nil
}

func (dao *RunnerDao) Delete(ctx context.Context, id int64, uid string) error {
	tx := dao.db.NewSession(ctx).Model(&types.Runner{})

	if id != 0 {
		tx = tx.Where("id = ?", id)
	}
	if uid != "" {
		tx = tx.Where("uid = ?", uid)
	}

	err := tx.Delete(&types.Runner{}).Error
	if err != nil {
		return err
	}
	return nil
}
