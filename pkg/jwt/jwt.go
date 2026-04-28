package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	OrgID     string `json:"org_id"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type JWTManager struct {
	secret          []byte
	accessTokenExp  time.Duration
	refreshTokenExp time.Duration
}

func NewJWTManager(secret string, accessExpHours, refreshExpDays int) *JWTManager {
	return &JWTManager{
		secret:          []byte(secret),
		accessTokenExp:  time.Duration(accessExpHours) * time.Hour,
		refreshTokenExp: time.Duration(refreshExpDays) * 24 * time.Hour,
	}
}

// GenerateTokenPair 生成访问令牌和刷新令牌
func (m *JWTManager) GenerateTokenPair(userID, username, orgID, role string) (*TokenPair, error) {
	now := time.Now()

	// 访问令牌
	accessClaims := Claims{
		UserID:   userID,
		Username: username,
		OrgID:    orgID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenExp)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(m.secret)
	if err != nil {
		return nil, err
	}

	// 刷新令牌
	refreshClaims := Claims{
		UserID:   userID,
		Username: username,
		OrgID:    orgID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTokenExp)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(m.secret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(m.accessTokenExp.Seconds()),
	}, nil
}

// ValidateToken 验证令牌
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshAccessToken 使用刷新令牌生成新的访问令牌
func (m *JWTManager) RefreshAccessToken(refreshTokenString string) (*TokenPair, error) {
	claims, err := m.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	return m.GenerateTokenPair(claims.UserID, claims.Username, claims.OrgID, claims.Role)
}