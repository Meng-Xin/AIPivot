package model

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type UserAuth struct {
	NickName string
	Email    string
	Password string
}

func (u *UserAuth) CheckEmail() error {
	if u.Email == "" {
		return errors.New("邮箱不能为空")
	}
	if !emailRegex.MatchString(u.Email) {
		return errors.New("邮箱格式不正确")
	}
	return nil
}

func (u *UserAuth) CheckPassword() error {
	if u.Password == "" {
		return errors.New("密码不能为空")
	}
	if len(u.Password) < 6 {
		return errors.New("密码长度不能少于6位")
	}
	return nil
}

func (u *UserAuth) EncryptPassword() (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (u *UserAuth) CheckPasswordMatch(hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(u.Password)) == nil
}
