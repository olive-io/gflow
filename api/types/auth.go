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

package types

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

const (
	DefaultAdministratorRole = "administrator"
	DefaultSystemRole        = "system"
	DefaultOperatorRole      = "operator"
	DefaultAdministratorUser = "admin"
	DefaultPassword          = "p@ssw0rd"

	DefaultMaxFailedLoginRetry = 5
	DefaultUserLockedDuration  = time.Minute
)

var (
	jwtSecret = []byte("Gflow Server")
	jwtIssuer = "gflow"

	defaultAgentKey   = "gflow-agent"
	defaultAgentValue = "runner"

	userInfoKey = "User-Info"
)

type systemRole map[int64]*Role

var SystemRoles = systemRole{
	1: &Role{
		Id:       1,
		Type:     RoleType_Admin,
		Name:     DefaultAdministratorRole,
		Metadata: map[string]string{},
	},
	2: &Role{
		Id:       2,
		Type:     RoleType_System,
		Name:     DefaultSystemRole,
		Metadata: map[string]string{},
	},
	3: &Role{
		Id:       3,
		Type:     RoleType_Operator,
		Name:     DefaultOperatorRole,
		Metadata: map[string]string{},
	},
}

var RootUser *User

func init() {
	RootUser = &User{
		Id:       1,
		Uid:      DefaultAdministratorUser,
		Username: DefaultAdministratorUser,
		Metadata: map[string]string{},
		RoleId:   SystemRoles[1].Id,
	}
	RootUser.SetPassword(DefaultPassword)
}

type Claims struct {
	jwt.RegisteredClaims
	UserId int64 `json:"userId"`
	RoleId int64 `json:"roleId"`
}

func GenerateToken(userId, roleId int64) (*Token, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(2 * time.Hour)

	claims := &Claims{
		UserId: userId,
		RoleId: roleId,
	}
	claims.ExpiresAt = jwt.NewNumericDate(expireTime)
	claims.Issuer = jwtIssuer
	claims.ID = fmt.Sprintf("%d", userId)

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tok, err := tokenClaims.SignedString(jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	token := &Token{
		Text:     tok,
		CreateAt: time.Now().Unix(),
		ExpireAt: expireTime.Unix(),
		Enable:   1,
		UserId:   userId,
		RoleId:   roleId,
	}

	return token, nil
}

func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (any, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := tokenClaims.Claims.(*Claims)
	if !ok || !tokenClaims.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	if time.Now().After(claims.ExpiresAt.UTC()) {
		return nil, fmt.Errorf("expired token")
	}

	return claims, nil
}

func AppendGflowAgent(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, defaultAgentKey, defaultAgentValue)
}

func IsGflowAgent(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}
	values := md.Get(defaultAgentKey)
	if len(values) == 0 {
		return false
	}
	return values[0] == defaultAgentValue
}

func (u *User) IsLocked() bool {
	if u.Login == nil {
		return false
	}

	after := time.Unix(u.Login.LockTimestamp, 0).Add(DefaultUserLockedDuration)
	return u.Login.IsLocked != 0 && time.Now().Before(after)
}

func (u *User) SetPassword(password string) {
	sha := sha512.New()
	sha.Write([]byte(password))
	u.HashedPassword = hex.EncodeToString(sha.Sum(nil))
}

func (u *User) VerifyPassword(password string) bool {
	sha := sha512.New()
	sha.Write([]byte(password))
	hash := hex.EncodeToString(sha.Sum(nil))
	return hash == u.HashedPassword
}

type UserInfo struct {
	Claims

	User *User
	Role *Role
}

func (u *UserInfo) IsAdmin() bool {
	role := u.Role
	if role == nil {
		return false
	}
	return role.Type == RoleType_Admin
}

func (u *UserInfo) IsManager() bool {
	role := u.Role
	if role == nil {
		return false
	}

	return role.Type == RoleType_Admin || role.Type == RoleType_System
}

func (u *UserInfo) IsUser() bool {
	role := u.Role
	if role == nil {
		return false
	}

	return role.Type == RoleType_Admin || role.Type == RoleType_System || role.Type == RoleType_Operator
}

func GetUserInfo(ctx context.Context) (*UserInfo, bool) {
	userInfo, ok := ctx.Value(userInfoKey).(*UserInfo)
	return userInfo, ok
}

func SetUserInfo(ctx context.Context, userInfo *UserInfo) context.Context {
	ctx = context.WithValue(ctx, userInfoKey, userInfo)
	return ctx
}
