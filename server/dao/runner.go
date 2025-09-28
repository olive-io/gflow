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

	"gorm.io/gorm"

	"github.com/olive-io/gflow/api/types"
)

type RunnerDao struct {
	db *gorm.DB
}

func NewRunnerDao(db *gorm.DB) (*RunnerDao, error) {
	err := db.AutoMigrate(
		&types.Runner{},
	)
	if err != nil {
		return nil, fmt.Errorf("auto migrate runner models: %w", err)
	}

	dao := &RunnerDao{
		db: db,
	}

	return dao, nil
}

func (dao *RunnerDao) FindRunners(ctx context.Context) ([]*types.Runner, error) {
	tx := dao.db.Session(&gorm.Session{}).WithContext(ctx).Model(&types.Runner{})

	runners := make([]*types.Runner, 0)
	err := tx.Order("id desc").Find(&runners).Error
	if err != nil {
		return nil, err
	}

	return runners, nil
}

func (dao *RunnerDao) ListRunners(ctx context.Context, page, size int32) ([]*types.Runner, int64, error) {
	tx := dao.db.Session(&gorm.Session{}).WithContext(ctx).Model(&types.Runner{})

	total := int64(0)
	err := tx.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := int((page - 1) * size)
	limit := int(size)
	runners := make([]*types.Runner, 0)
	tx = dao.db.Session(&gorm.Session{}).WithContext(ctx).Model(&types.Runner{})
	if offset > 0 {
		tx = tx.Offset(offset)
	}
	if limit > 0 {
		tx = tx.Limit(limit)
	}
	err = tx.Order("id desc").Find(&runners).Error
	if err != nil {
		return nil, 0, err
	}

	return runners, total, nil
}

func (dao *RunnerDao) GetRunner(ctx context.Context, id uint64, uid string) (*types.Runner, error) {
	tx := dao.db.Session(&gorm.Session{}).WithContext(ctx).Model(&types.Runner{})

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

func (dao *RunnerDao) CreateRunner(ctx context.Context, runner *types.Runner) error {
	tx := dao.db.Session(&gorm.Session{}).WithContext(ctx).Model(&types.Runner{})

	err := tx.Create(runner).Error
	if err != nil {
		return err
	}
	runner.Id = uint64(tx.RowsAffected)

	return nil
}

func (dao *RunnerDao) UpdateRunner(ctx context.Context, runner *types.Runner) error {
	tx := dao.db.Session(&gorm.Session{}).WithContext(ctx).Model(&types.Runner{})

	err := tx.Where("uid = ?", runner.Uid).Updates(runner).Error
	if err != nil {
		return err
	}
	return nil
}

func (dao *RunnerDao) RemoveRunner(ctx context.Context, id uint64, uid string) error {
	tx := dao.db.Session(&gorm.Session{}).WithContext(ctx).Model(&types.Runner{})

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
