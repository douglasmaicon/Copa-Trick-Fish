package handlers

import (
	"net/http"

	"github.com/douglasmaicon/Copa-Trick-Fish/internal/database"
	"github.com/douglasmaicon/Copa-Trick-Fish/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListarEtapas retorna todas as etapas com filtros opcionais
func ListarEtapas(c *gin.Context) {
	edicaoID := c.Query("edicao_id")
	status := c.Query("status")

	var etapas []models.Etapa
	query := database.DB.Preload("Modalidade").Preload("Edicao")

	if edicaoID != "" {
		query = query.Where("edicao_id = ?", edicaoID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	result := query.Order("data_largada ASC").Find(&etapas)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao buscar etapas",
		})
		return
	}

	c.JSON(http.StatusOK, etapas)
}

// BuscarEtapa retorna uma etapa específica por ID
func BuscarEtapa(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}

	var etapa models.Etapa
	result := database.DB.
		Preload("Modalidade").
		Preload("Edicao").
		Preload("Inscricoes.Competidor").
		First(&etapa, "id = ?", id)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Etapa não encontrada",
		})
		return
	}

	c.JSON(http.StatusOK, etapa)
}

// CriarEtapa cria uma nova etapa
func CriarEtapa(c *gin.Context) {
	var etapa models.Etapa

	if err := c.ShouldBindJSON(&etapa); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dados inválidos: " + err.Error(),
		})
		return
	}

	// Validar se a edição existe
	var edicao models.Edicao
	if err := database.DB.First(&edicao, "id = ?", etapa.EdicaoID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Edição não encontrada",
		})
		return
	}

	// Validar se a modalidade existe
	var modalidade models.Modalidade
	if err := database.DB.First(&modalidade, "id = ?", etapa.ModalidadeID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Modalidade não encontrada",
		})
		return
	}

	result := database.DB.Create(&etapa)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao criar etapa: " + result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, etapa)
}

// AtualizarEtapa atualiza uma etapa existente
func AtualizarEtapa(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}

	var etapa models.Etapa
	if err := database.DB.First(&etapa, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Etapa não encontrada",
		})
		return
	}

	if err := c.ShouldBindJSON(&etapa); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dados inválidos: " + err.Error(),
		})
		return
	}

	database.DB.Save(&etapa)

	c.JSON(http.StatusOK, etapa)
}

// DeletarEtapa remove uma etapa (soft delete)
func DeletarEtapa(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}

	result := database.DB.Delete(&models.Etapa{}, "id = ?", id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao deletar etapa",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Etapa não encontrada",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Etapa deletada com sucesso",
	})
}
