package handlers

import (
	"net/http"

	"github.com/douglasmaicon/Copa-Trick-Fish/internal/database"
	"github.com/douglasmaicon/Copa-Trick-Fish/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListarEdicoes retorna todas as edições
func ListarEdicoes(c *gin.Context) {
	var edicoes []models.Edicao

	result := database.DB.Preload("Modalidades").
		Order("ano DESC").
		Find(&edicoes)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao buscar edições",
		})
		return
	}

	c.JSON(http.StatusOK, edicoes)
}

// BuscarEdicao retorna uma edição específica por ID
func BuscarEdicao(c *gin.Context) {
	id := c.Param("id")

	// Validar UUID
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}

	var edicao models.Edicao
	result := database.DB.Preload("Modalidades").
		Preload("Etapas").
		First(&edicao, "id = ?", id)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Edição não encontrada",
		})
		return
	}

	c.JSON(http.StatusOK, edicao)
}

// BuscarEdicaoAtiva retorna a edição atualmente ativa
func BuscarEdicaoAtiva(c *gin.Context) {
	var edicao models.Edicao

	result := database.DB.Where("ativa = ?", true).
		Preload("Modalidades").
		Preload("Etapas").
		First(&edicao)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Nenhuma edição ativa encontrada",
		})
		return
	}

	c.JSON(http.StatusOK, edicao)
}

// CriarEdicao cria uma nova edição
func CriarEdicao(c *gin.Context) {
	var edicao models.Edicao

	if err := c.ShouldBindJSON(&edicao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dados inválidos: " + err.Error(),
		})
		return
	}

	// Se for ativa, desativar outras edições
	if edicao.Ativa {
		database.DB.Model(&models.Edicao{}).
			Where("ativa = ?", true).
			Update("ativa", false)
	}

	result := database.DB.Create(&edicao)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao criar edição: " + result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, edicao)
}

// AtualizarEdicao atualiza uma edição existente
func AtualizarEdicao(c *gin.Context) {
	id := c.Param("id")

	// Validar UUID
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}

	var edicao models.Edicao
	if err := database.DB.First(&edicao, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Edição não encontrada",
		})
		return
	}

	if err := c.ShouldBindJSON(&edicao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dados inválidos: " + err.Error(),
		})
		return
	}

	// Se for ativa, desativar outras edições
	if edicao.Ativa {
		database.DB.Model(&models.Edicao{}).
			Where("ativa = ? AND id != ?", true, id).
			Update("ativa", false)
	}

	database.DB.Save(&edicao)

	c.JSON(http.StatusOK, edicao)
}

// DeletarEdicao remove uma edição (soft delete)
func DeletarEdicao(c *gin.Context) {
	id := c.Param("id")

	// Validar UUID
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}

	result := database.DB.Delete(&models.Edicao{}, "id = ?", id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao deletar edição",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Edição não encontrada",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Edição deletada com sucesso",
	})
}
