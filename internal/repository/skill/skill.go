package skill

import (
	"context"
	"errors"

	"aipivot/internal/shared/po"
	"aipivot/internal/shared/query"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// Repository Skill 数据仓储接口。
type Repository interface {
	Create(ctx context.Context, s *po.Skill) error
	GetByID(ctx context.Context, id, tenantID int64) (*po.Skill, error)
	GetListByTenant(ctx context.Context, tenantID int64) ([]*po.Skill, error)
	// GetEnabledByTenant 只返回 enabled=true 的工具，供 Agent 请求时动态加载。
	GetEnabledByTenant(ctx context.Context, tenantID int64) ([]*po.Skill, error)
	Update(ctx context.Context, s *po.Skill) error
	Delete(ctx context.Context, id, tenantID int64) error
}

// SkillRepo Skill 仓储实现。
type SkillRepo struct {
	q  *query.Query
	db *gorm.DB
}

func NewSkillRepo(q *query.Query, db *gorm.DB) *SkillRepo {
	return &SkillRepo{q: q, db: db}
}

func (r *SkillRepo) Create(ctx context.Context, s *po.Skill) error {
	err := r.q.Skill.WithContext(ctx).Create(s)
	if err != nil {
		logx.WithContext(ctx).Errorf("SkillRepo.Create err: %v", err)
		return err
	}
	return nil
}

func (r *SkillRepo) GetByID(ctx context.Context, id, tenantID int64) (*po.Skill, error) {
	sk := r.q.Skill
	result, err := sk.WithContext(ctx).
		Where(sk.ID.Eq(id), sk.TenantID.Eq(tenantID)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logx.WithContext(ctx).Errorf("SkillRepo.GetByID err: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *SkillRepo) GetListByTenant(ctx context.Context, tenantID int64) ([]*po.Skill, error) {
	sk := r.q.Skill
	list, err := sk.WithContext(ctx).
		Where(sk.TenantID.Eq(tenantID)).
		Order(sk.CreatedAt.Desc()).
		Find()
	if err != nil {
		logx.WithContext(ctx).Errorf("SkillRepo.GetListByTenant err: %v", err)
		return nil, err
	}
	return list, nil
}

func (r *SkillRepo) GetEnabledByTenant(ctx context.Context, tenantID int64) ([]*po.Skill, error) {
	sk := r.q.Skill
	list, err := sk.WithContext(ctx).
		Where(sk.TenantID.Eq(tenantID), sk.Enabled.Is(true)).
		Order(sk.Name.Asc()).
		Find()
	if err != nil {
		logx.WithContext(ctx).Errorf("SkillRepo.GetEnabledByTenant err: %v", err)
		return nil, err
	}
	return list, nil
}

func (r *SkillRepo) Update(ctx context.Context, s *po.Skill) error {
	_, err := r.q.Skill.WithContext(ctx).
		Where(r.q.Skill.ID.Eq(s.ID), r.q.Skill.TenantID.Eq(s.TenantID)).
		Updates(s)
	if err != nil {
		logx.WithContext(ctx).Errorf("SkillRepo.Update err: %v", err)
		return err
	}
	return nil
}

func (r *SkillRepo) Delete(ctx context.Context, id, tenantID int64) error {
	sk := r.q.Skill
	_, err := sk.WithContext(ctx).
		Where(sk.ID.Eq(id), sk.TenantID.Eq(tenantID)).
		Delete()
	if err != nil {
		logx.WithContext(ctx).Errorf("SkillRepo.Delete err: %v", err)
		return err
	}
	return nil
}
