package dao

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/olive-io/gflow/pkg/dbutil"
)

type Op int

const (
	OpEq Op = iota + 1
	OpNe
	OpLt
	OpLe
	OpGt
	OpGe
	OpIn
	OpLike
)

type CondItem struct {
	Op    Op
	Name  string
	Value any
}

func NewCond(op Op, name string, value any) *CondItem {
	return &CondItem{
		Op:    op,
		Name:  name,
		Value: value,
	}
}

type InnerDao[T any] struct {
	db     *dbutil.DB
	target T
}

func newDao[T any](db *dbutil.DB, target T) (*InnerDao[T], error) {
	if err := db.AutoMigrate(target); err != nil {
		return nil, err
	}

	dao := &InnerDao[T]{
		db:     db,
		target: target,
	}

	return dao, nil
}

func (dao *InnerDao[T]) NewSession(ctx context.Context) *gorm.DB {
	return dao.db.Session(&gorm.Session{
		NewDB:       true,
		Initialized: true,
		Context:     ctx,
	})
}

func (dao *InnerDao[T]) Find(ctx context.Context, cond any) ([]*T, error) {

	var err error

	session := dao.NewSession(ctx).Model(dao.target)
	if cond != nil {
		session = dao.applyCond(session, cond)
	}

	results := make([]*T, 0)
	if err = session.Find(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

func (dao *InnerDao[T]) PageList(ctx context.Context, page, size int, cond any) ([]*T, int64, error) {
	var err error

	session := dao.NewSession(ctx).Model(dao.target)
	session = dao.applyCond(session, cond)

	var total int64
	if err = session.Count(&total).Error; err != nil {
		return nil, total, err
	}
	if page > 0 && size > 0 {
		session.Offset((page - 1) * size).Limit(size)
	}

	list := make([]*T, 0)
	if err = session.Order("id desc").Find(&list).Error; err != nil {
		return nil, total, err
	}

	return list, total, nil
}

func (dao *InnerDao[T]) Count(ctx context.Context, cond any) (int64, error) {

	var total int64
	var err error
	session := dao.NewSession(ctx).Model(dao.target)
	session = dao.applyCond(session, cond)
	if err = session.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (dao *InnerDao[T]) Get(ctx context.Context, id int64) (*T, error) {
	session := dao.NewSession(ctx).Model(dao.target)

	target := new(T)
	err := session.Where("id = ?", id).First(&target).Error
	if err != nil {
		return nil, err
	}
	return target, nil
}

func (dao *InnerDao[T]) First(ctx context.Context, cond any) (*T, error) {
	session := dao.NewSession(ctx).Model(dao.target)

	if cond != nil {
		session = dao.applyCond(session, cond)
	}

	target := new(T)
	err := session.First(&target).Error
	if err != nil {
		return nil, err
	}
	return target, nil
}

func (dao *InnerDao[T]) Create(ctx context.Context, value *T) (int64, error) {
	session := dao.NewSession(ctx).Model(dao.target)
	err := session.Create(value).Error
	if err != nil {
		return 0, err
	}
	return session.RowsAffected, nil
}

func (dao *InnerDao[T]) Update(ctx context.Context, id int64, value any) error {
	session := dao.NewSession(ctx).Model(dao.target)

	err := session.Where("id = ?", id).UpdateColumns(value).Error
	if err != nil {
		return err
	}
	if session.RowsAffected == 0 {
		return fmt.Errorf("%w: get by id %d", gorm.ErrRecordNotFound, id)
	}

	return nil
}

func (dao *InnerDao[T]) Delete(ctx context.Context, id int64) error {
	session := dao.NewSession(ctx).Model(dao.target)

	err := session.Delete(dao.target, "id = ?", id).Error
	if err != nil {
		return err
	}

	if session.RowsAffected == 0 {
		return fmt.Errorf("%w: get by id %d", gorm.ErrRecordNotFound, id)
	}
	return nil
}

func (dao *InnerDao[T]) applyCond(session *gorm.DB, cond any) *gorm.DB {
	if cond != nil {
		switch tv := cond.(type) {
		case map[string]any:
			for k, v := range tv {
				session = session.Where(fmt.Sprintf("%s = ?", k), v)
			}
		case *CondItem:
			item := tv
			switch item.Op {
			case OpEq:
				session = session.Where(fmt.Sprintf("%s = ?", item.Name), item.Value)
			case OpNe:
				session = session.Where(fmt.Sprintf("%s <> ?", item.Name), item.Value)
			case OpLt:
				session = session.Where(fmt.Sprintf("%s < ?", item.Name), item.Value)
			case OpLe:
				session = session.Where(fmt.Sprintf("%s <= ?", item.Name), item.Value)
			case OpGt:
				session = session.Where(fmt.Sprintf("%s > ?", item.Name), item.Value)
			case OpGe:
				session = session.Where(fmt.Sprintf("%s >= ?", item.Name), item.Value)
			case OpIn:
				session = session.Where(fmt.Sprintf("%s IN (?)", item.Name), item.Value)
			case OpLike:
				session = session.Where(fmt.Sprintf("%s LIKE ?", item.Name), item.Value)
			}
		case []*CondItem:
			for _, item := range tv {
				session = dao.applyCond(session, item)
			}
		default:
			session = session.Where(cond)
		}
	}
	return session
}
