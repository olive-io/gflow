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

package authutil

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"

	"github.com/olive-io/gflow/api/types"
)

var (
	jwtSecret = []byte("Gflow Server")
	jwtIssuer = "gflow"

	defaultAgentKey   = "gflow-agent"
	defaultAgentValue = "runner"
)

type Claims struct {
	jwt.RegisteredClaims
	UserId int64 `json:"userId"`
	RoleId int64 `json:"roleId"`
}

func GenerateToken(userId, roleId int64) (*types.Token, error) {
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

	token := &types.Token{
		Text:     tok,
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
