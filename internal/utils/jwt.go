package utils

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"webapi-v2-neo/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

func ValidateJWT(tokenString string, jwtConfig config.JWTConfig) error {
	// 解析公钥
	publicKey, err := parsePublicKey(jwtConfig.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	// 解析和验证 token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if token.Method.Alg() != jwtConfig.Algorithm {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	// 验证 claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if iss, ok := claims["iss"].(string); ok && iss != jwtConfig.Issuer {
			return fmt.Errorf("invalid issuer")
		}
		if sub, ok := claims["sub"].(string); ok && sub != jwtConfig.Subject {
			return fmt.Errorf("invalid subject")
		}
	}

	return nil
}

func parsePublicKey(publicKeyPEM string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	ecdsaKey, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an ECDSA public key")
	}

	return ecdsaKey, nil
}