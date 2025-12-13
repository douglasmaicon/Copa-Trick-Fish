package handlers

import (
	"net/http"
	"strings"

	"github.com/douglasmaicon/Copa-Trick-Fish/internal/auth"
	"github.com/douglasmaicon/Copa-Trick-Fish/internal/config"
	"github.com/douglasmaicon/Copa-Trick-Fish/internal/database"
	"github.com/douglasmaicon/Copa-Trick-Fish/internal/models"
	"github.com/gin-gonic/gin"
)

// LoginUsuario faz login de usuário (admin, organizador, fiscal)
func LoginUsuario(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
		Senha string `json:"senha" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dados inválidos: " + err.Error(),
		})
		return
	}

	// Buscar usuário
	var usuario models.Usuario
	if err := database.DB.Where("email = ?", input.Email).First(&usuario).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Email ou senha incorretos",
		})
		return
	}

	// Verificar se pode acessar
	if !usuario.PodeAcessar() {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Usuário inativo ou bloqueado",
		})
		return
	}

	// Verificar senha
	if !usuario.CheckPassword(input.Senha) {
		// Registrar falha
		usuario.RegistrarFalhaLogin()
		database.DB.Save(&usuario)

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Email ou senha incorretos",
		})
		return
	}

	// Registrar acesso
	usuario.RegistrarAcesso()
	database.DB.Save(&usuario)

	// Gerar token
	cfg := config.AppConfig
	token, err := auth.GerarTokenUsuario(&usuario, cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao gerar token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"tipo":  usuario.Tipo,
		"nome":  usuario.Nome,
		"email": usuario.Email,
	})
}

// LoginCompetidor faz login de competidor
func LoginCompetidor(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
		Senha string `json:"senha" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dados inválidos: " + err.Error(),
		})
		return
	}

	// Buscar competidor
	var competidor models.Competidor
	if err := database.DB.Where("email = ?", input.Email).First(&competidor).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Email ou senha incorretos",
		})
		return
	}

	// Verificar se pode acessar
	if pode, motivo := competidor.PodeSeCadastrar(); !pode {
		c.JSON(http.StatusForbidden, gin.H{
			"error": motivo,
		})
		return
	}

	// Verificar senha
	if !competidor.CheckPassword(input.Senha) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Email ou senha incorretos",
		})
		return
	}

	// Registrar acesso
	competidor.RegistrarAcesso()
	database.DB.Save(&competidor)

	// Gerar token
	cfg := config.AppConfig
	token, err := auth.GerarTokenCompetidor(&competidor, cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao gerar token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"tipo":  "competidor",
		"nome":  competidor.Nome,
		"email": competidor.Email,
	})
}

// RefreshToken renova o token
func RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Token não fornecido",
		})
		return
	}

	// Extrair token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Formato de token inválido",
		})
		return
	}

	tokenString := parts[1]
	cfg := config.AppConfig

	// Renovar token
	newToken, err := auth.RefreshToken(tokenString, cfg)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Erro ao renovar token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": newToken,
	})
}

// MeuPerfil retorna dados do usuário autenticado
func MeuPerfil(c *gin.Context) {
	userID, _ := c.Get("user_id")
	tipo, _ := c.Get("tipo")

	if tipo == "competidor" {
		var competidor models.Competidor
		if err := database.DB.First(&competidor, "id = ?", userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Competidor não encontrado",
			})
			return
		}

		c.JSON(http.StatusOK, competidor)
	} else {
		var usuario models.Usuario
		if err := database.DB.First(&usuario, "id = ?", userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Usuário não encontrado",
			})
			return
		}

		c.JSON(http.StatusOK, usuario)
	}
}
