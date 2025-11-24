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

type RoleDao struct {
	*InnerDao[types.Role]
}

func NewRoleDao(db *dbutil.DB) (*RoleDao, error) {
	inner, err := newDao(db, types.Role{})
	if err != nil {
		return nil, fmt.Errorf("auto migrate role models: %w", err)
	}
	dao := &RoleDao{
		InnerDao: inner,
	}

	return dao, nil
}

func (dao *RoleDao) GetByName(ctx context.Context, name string) (*types.Role, error) {
	tx := dao.db.NewSession(ctx).Model(&types.Role{})

	role := &types.Role{}
	err := tx.Where("name = ?", name).First(role).Error
	if err != nil {
		return nil, err
	}

	return role, nil
}

func (dao *RoleDao) RelationalUserCount(ctx context.Context, id int64, cond any) (int64, error) {
	var err error

	session := dao.NewSession(ctx).Model(dao.target)
	session = dao.applyCond(session, cond).Where("role_id = ?", id)

	var total int64
	if err = session.Count(&total).Error; err != nil {
		return total, err
	}

	return total, nil
}

func (dao *RoleDao) ListRelationUsers(ctx context.Context, id int64, page, size int, cond any) ([]*types.User, int64, error) {
	var err error

	session := dao.NewSession(ctx).Model(dao.target)
	session = dao.applyCond(session, cond).Where("role_id = ?", id)

	var total int64
	if err = session.Count(&total).Error; err != nil {
		return nil, total, err
	}
	if page > 0 && size > 0 {
		session.Offset((page - 1) * size).Limit(size)
	}

	list := make([]*types.User, 0)
	if err = session.Order("id desc").Find(&list).Error; err != nil {
		return nil, total, err
	}

	return list, total, nil
}
