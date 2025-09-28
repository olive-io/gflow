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

type DefinitionsDao struct {
	db *gorm.DB
}

func NewDefinitionsDao(db *gorm.DB) (*DefinitionsDao, error) {
	err := db.AutoMigrate(
		&types.Definitions{},
	)
	if err != nil {
		return nil, fmt.Errorf("auto migrate definitions models: %w", err)
	}

	dao := &DefinitionsDao{
		db: db,
	}

	return dao, nil
}

func (dao *DefinitionsDao) ListDefinitions(ctx context.Context, page, size int32) ([]*types.Definitions, int64, error) {
	tx := dao.db.Session(&gorm.Session{}).WithContext(ctx).Model(&types.Definitions{})

	total := int64(0)
	err := tx.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := int((page - 1) * size)
	limit := int(size)
	definitionsList := make([]*types.Definitions, 0)
	tx = dao.db.Session(&gorm.Session{}).WithContext(ctx).Model(&types.Definitions{})
	if offset > 0 {
		tx = tx.Offset(offset)
	}
	if limit > 0 {
		tx = tx.Limit(limit)
	}
	err = tx.Order("id desc").Find(&definitionsList).Error
	if err != nil {
		return nil, 0, err
	}

	return definitionsList, total, nil
}

func (dao *DefinitionsDao) GetDefinitions(ctx context.Context, id int64, uid string) (*types.Definitions, error) {
	tx := dao.db.Session(&gorm.Session{}).WithContext(ctx).Model(&types.Definitions{})

	var definitions types.Definitions

	if id != 0 {
		tx = tx.Where("id = ?", id)
	}
	if uid != "" {
		tx = tx.Where("uid = ?", uid)
	}

	err := tx.First(&definitions).Error
	if err != nil {
		return nil, err
	}
	return &definitions, nil
}

func (dao *DefinitionsDao) GetDefinitionsWithVersion(ctx context.Context, id int64, uid string, version uint64) (*types.Definitions, error) {
	tx := dao.db.Session(&gorm.Session{}).WithContext(ctx).Model(&types.Definitions{})

	var definitions types.Definitions

	if id != 0 {
		tx = tx.Where("id = ?", id)
	}
	if uid != "" {
		tx = tx.Where("uid = ?", uid)
	}
	if version != 0 {
		tx = tx.Where("version = ?", version)
	}

	err := tx.First(&definitions).Error
	if err != nil {
		return nil, err
	}
	return &definitions, nil
}

func (dao *DefinitionsDao) CreateDefinitions(ctx context.Context, definitions *types.Definitions) (int64, error) {
	tx := dao.db.Session(&gorm.Session{}).WithContext(ctx).Model(&types.Definitions{})

	if definitions.Version == 0 {
		definitions.Version = 1
	}
	err := tx.Create(&definitions).Error
	if err != nil {
		return 0, err
	}

	return tx.RowsAffected, nil
}

func (dao *DefinitionsDao) UpdateDefinitions(ctx context.Context, definitions *types.Definitions) error {
	tx := dao.db.Session(&gorm.Session{}).WithContext(ctx).Model(&types.Definitions{})
	err := tx.Where("id = ?", definitions.Id).Updates(definitions).Error
	if err != nil {
		return err
	}

	return nil
}
