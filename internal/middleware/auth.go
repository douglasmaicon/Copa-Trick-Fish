package middleware

import (
	"net/http"
	"strings"

	"github.com/douglasmaicon/Copa-Trick-Fish/internal/auth"
	"github.com/douglasmaicon/Copa-Trick-Fish/internal/config"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware verifica se o usuário está autenticado
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obter token do header Authorization
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token de autenticação não fornecido",
			})
			c.Abort()
			return
		}

		// Formato esperado: "Bearer {token}"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Formato de token inválido. Use: Bearer {token}",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validar token
		claims, err := auth.ValidarToken(tokenString, cfg)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token inválido ou expirado",
			})
			c.Abort()
			return
		}

		// Armazenar claims no contexto para usar nos handlers
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("tipo", claims.Tipo)
		c.Set("nome", claims.Nome)

		c.Next()
	}
}

// AdminOnly verifica se o usuário é admin
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		tipo, exists := c.Get("tipo")

		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Usuário não autenticado",
			})
			c.Abort()
			return
		}

		if tipo != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Acesso restrito a administradores",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OrganizadorOnly verifica se é organizador ou admin
func OrganizadorOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		tipo, exists := c.Get("tipo")

		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Usuário não autenticado",
			})
			c.Abort()
			return
		}

		tipoStr := tipo.(string)
		if tipoStr != "admin" && tipoStr != "organizador" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Acesso restrito a organizadores",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// FiscalOnly verifica se é fiscal, organizador ou admin
func FiscalOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		tipo, exists := c.Get("tipo")

		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Usuário não autenticado",
			})
			c.Abort()
			return
		}

		tipoStr := tipo.(string)
		if tipoStr != "admin" && tipoStr != "organizador" && tipoStr != "fiscal" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Acesso restrito a fiscais",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
