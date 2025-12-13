package auth

import (
	"errors"
	"time"

	"github.com/douglasmaicon/Copa-Trick-Fish/internal/config"
	"github.com/douglasmaicon/Copa-Trick-Fish/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

// Claims representa os dados do token JWT
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Tipo   string `json:"tipo"` // admin, organizador, fiscal, competidor
	Nome   string `json:"nome"`
	jwt.RegisteredClaims
}

// GerarToken gera um token JWT para um usuário
func GerarTokenUsuario(usuario *models.Usuario, cfg *config.Config) (string, error) {
	expirationTime := time.Now().Add(time.Duration(cfg.JWT.Expiration) * time.Hour)

	claims := &Claims{
		UserID: usuario.ID.String(),
		Email:  usuario.Email,
		Tipo:   usuario.Tipo,
		Nome:   usuario.Nome,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "copa-trick-fish",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GerarTokenCompetidor gera um token JWT para um competidor
func GerarTokenCompetidor(competidor *models.Competidor, cfg *config.Config) (string, error) {
	expirationTime := time.Now().Add(time.Duration(cfg.JWT.Expiration) * time.Hour)

	claims := &Claims{
		UserID: competidor.ID.String(),
		Email:  competidor.Email,
		Tipo:   "competidor",
		Nome:   competidor.Nome,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "copa-trick-fish",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidarToken valida e decodifica um token JWT
func ValidarToken(tokenString string, cfg *config.Config) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de assinatura inválido")
		}
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("token inválido")
	}

	return claims, nil
}

// RefreshToken renova um token JWT
func RefreshToken(tokenString string, cfg *config.Config) (string, error) {
	claims, err := ValidarToken(tokenString, cfg)
	if err != nil {
		return "", err
	}

	// Criar novo token com nova expiração
	expirationTime := time.Now().Add(time.Duration(cfg.JWT.Expiration) * time.Hour)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	newTokenString, err := token.SignedString([]byte(cfg.JWT.Secret))

	if err != nil {
		return "", err
	}

	return newTokenString, nil
}
