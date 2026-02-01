package utils

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"reflect"
)

// JWTParseToken 解析 JWT Token
func JWTParseToken[T any](token, secret string) (*T, error) {
	// 使用反射创建 T 类型的实例
	claimsType := reflect.TypeOf((*T)(nil)).Elem()
	if claimsType == nil {
		return nil, errors.New("invalid claims type")
	}

	// 创建新实例
	claimsValue := reflect.New(claimsType).Interface()

	// 转换为 jwt.Claims
	jwtClaims, ok := claimsValue.(jwt.Claims)
	if !ok {
		return nil, errors.New("type T does not implement jwt.Claims")
	}

	// 解析 Token
	parsedToken, err := jwt.ParseWithClaims(token, jwtClaims, func(tk *jwt.Token) (interface{}, error) {
		if _, ok := tk.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if parsedToken.Valid {
		// 使用反射进行比较
		parsedClaimsValue := reflect.ValueOf(parsedToken.Claims)
		expectedType := reflect.TypeOf((*T)(nil))

		if parsedClaimsValue.Type().AssignableTo(expectedType) {
			return parsedClaimsValue.Interface().(*T), nil
		}
	}

	return nil, errors.New("invalid token")
}

// JWTGenToken 生成 JWT Token
func JWTGenToken(secret string, claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
