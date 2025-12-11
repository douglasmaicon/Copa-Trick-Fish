package handlers

import (
	"net/http"

	"github.com/douglasmaicon/Copa-Trick-Fish/internal/database"
	"github.com/douglasmaicon/Copa-Trick-Fish/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListarCompetidores retorna todos os competidores
func ListarCompetidores(c *gin.Context) {
	var competidores []models.Competidor

	query := database.DB.Select("id", "nome", "email", "telefone", "cidade", "estado", "ativo", "banido")

	// Filtro por status
	if ativo := c.Query("ativo"); ativo != "" {
		query = query.Where("ativo = ?", ativo == "true")
	}

	if banido := c.Query("banido"); banido != "" {
		query = query.Where("banido = ?", banido == "true")
	}

	result := query.Order("nome ASC").Find(&competidores)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao buscar competidores",
		})
		return
	}

	c.JSON(http.StatusOK, competidores)
}

// BuscarCompetidor retorna um competidor específico
func BuscarCompetidor(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}

	var competidor models.Competidor
	result := database.DB.
		Select("id", "nome", "email", "telefone", "cpf", "data_nascimento", "cidade", "estado", "licenca_pesca", "validade_licenca", "foto_url", "ativo", "banido", "motivo_banimento", "created_at").
		First(&competidor, "id = ?", id)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Competidor não encontrado",
		})
		return
	}

	c.JSON(http.StatusOK, competidor)
}

// CriarCompetidor registra um novo competidor
func CriarCompetidor(c *gin.Context) {
	var competidor models.Competidor

	if err := c.ShouldBindJSON(&competidor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dados inválidos: " + err.Error(),
		})
		return
	}

	// Verificar se email já existe
	var count int64
	database.DB.Model(&models.Competidor{}).Where("email = ?", competidor.Email).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Email já cadastrado",
		})
		return
	}

	// Verificar se CPF já existe
	if competidor.CPF != "" {
		database.DB.Model(&models.Competidor{}).Where("cpf = ?", competidor.CPF).Count(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error": "CPF já cadastrado",
			})
			return
		}
	}

	// Hash da senha (precisa vir no campo senha do JSON)
	senhaTemporaria := c.PostForm("senha")
	if senhaTemporaria == "" {
		senhaTemporaria = "123456" // Senha padrão temporária
	}

	if err := competidor.SetPassword(senhaTemporaria); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Erro ao processar senha: " + err.Error(),
		})
		return
	}

	result := database.DB.Create(&competidor)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao criar competidor: " + result.Error.Error(),
		})
		return
	}

	// Limpar senha antes de retornar
	competidor.Senha = ""

	c.JSON(http.StatusCreated, competidor)
}

// AtualizarCompetidor atualiza dados de um competidor
func AtualizarCompetidor(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}

	var competidor models.Competidor
	if err := database.DB.First(&competidor, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Competidor não encontrado",
		})
		return
	}

	if err := c.ShouldBindJSON(&competidor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dados inválidos: " + err.Error(),
		})
		return
	}

	database.DB.Save(&competidor)

	// Limpar senha antes de retornar
	competidor.Senha = ""

	c.JSON(http.StatusOK, competidor)
}

// BanirCompetidor bane um competidor do torneio
func BanirCompetidor(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}

	var input struct {
		Motivo string `json:"motivo" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Motivo do banimento é obrigatório",
		})
		return
	}

	var competidor models.Competidor
	if err := database.DB.First(&competidor, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Competidor não encontrado",
		})
		return
	}

	competidor.Banir(input.Motivo)
	database.DB.Save(&competidor)

	c.JSON(http.StatusOK, gin.H{
		"message": "Competidor banido com sucesso",
	})
}

// DesbanirCompetidor remove o banimento de um competidor
func DesbanirCompetidor(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}

	var competidor models.Competidor
	if err := database.DB.First(&competidor, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Competidor não encontrado",
		})
		return
	}

	competidor.Desbanir()
	database.DB.Save(&competidor)

	c.JSON(http.StatusOK, gin.H{
		"message": "Banimento removido com sucesso",
	})
}
