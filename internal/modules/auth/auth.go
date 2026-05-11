package auth

import (
	"context"
	"errors"
	"regexp"

	"aipivot/internal/shared/po"

	"golang.org/x/crypto/bcrypt"
)

// ========== 校验 ==========

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("邮箱不能为空")
	}
	if !emailRegex.MatchString(email) {
		return errors.New("邮箱格式不正确")
	}
	return nil
}

func ValidatePassword(password string) error {
	if password == "" {
		return errors.New("密码不能为空")
	}
	if len(password) < 6 {
		return errors.New("密码长度不能少于6位")
	}
	return nil
}

func EncryptPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPasswordMatch(password, hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

// ========== Repository 接口 ==========

type UserRepository interface {
	Create(ctx context.Context, user *po.User) error
	GetByEmail(ctx context.Context, tenantID int64, email string) (*po.User, error)
	GetByID(ctx context.Context, id int64) (*po.User, error)
	UpdateLastLogin(ctx context.Context, id int64) error
}

type TenantRepository interface {
	GetBySlug(ctx context.Context, slug string) (*po.Tenant, error)
	GetByID(ctx context.Context, id int64) (*po.Tenant, error)
}

type ApiKeyRepository interface {
	Create(ctx context.Context, key *po.ApiKey) error
	GetByKeyHash(ctx context.Context, keyHash string) (*po.ApiKey, error)
	GetByID(ctx context.Context, id int64) (*po.ApiKey, error)
	GetListByTenant(ctx context.Context, tenantID int64) ([]*po.ApiKey, error)
	UpdateLastUsed(ctx context.Context, id int64) error
	Revoke(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
}
