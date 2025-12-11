package handlers

import (
	"net/http"
	"time"

	"github.com/douglasmaicon/Copa-Trick-Fish/internal/database"
	"github.com/douglasmaicon/Copa-Trick-Fish/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListarInscricoes retorna todas as inscrições com filtros
func ListarInscricoes(c *gin.Context) {
	etapaID := c.Query("etapa_id")
	competidorID := c.Query("competidor_id")
	status := c.Query("status_pagamento")

	var inscricoes []models.Inscricao
	query := database.DB.Preload("Etapa").Preload("Competidor").Preload("Regua")

	if etapaID != "" {
		query = query.Where("etapa_id = ?", etapaID)
	}

	if competidorID != "" {
		query = query.Where("competidor_id = ?", competidorID)
	}

	if status != "" {
		query = query.Where("status_pagamento = ?", status)
	}

	result := query.Order("data_inscricao DESC").Find(&inscricoes)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao buscar inscrições",
		})
		return
	}

	c.JSON(http.StatusOK, inscricoes)
}

// BuscarInscricao retorna uma inscrição específica
func BuscarInscricao(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}

	var inscricao models.Inscricao
	result := database.DB.
		Preload("Etapa").
		Preload("Competidor").
		Preload("Regua").
		Preload("Capturas").
		First(&inscricao, "id = ?", id)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Inscrição não encontrada",
		})
		return
	}

	c.JSON(http.StatusOK, inscricao)
}

// CriarInscricao cria uma nova inscrição
func CriarInscricao(c *gin.Context) {
	var inscricao models.Inscricao

	if err := c.ShouldBindJSON(&inscricao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dados inválidos: " + err.Error(),
		})
		return
	}

	// Verificar se a etapa existe e está aberta
	var etapa models.Etapa
	if err := database.DB.First(&etapa, "id = ?", inscricao.EtapaID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Etapa não encontrada",
		})
		return
	}

	// Verificar se pode inscrever
	if pode, motivo := etapa.PodeInscrever(); !pode {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": motivo,
		})
		return
	}

	// Verificar se o competidor existe
	var competidor models.Competidor
	if err := database.DB.First(&competidor, "id = ?", inscricao.CompetidorID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Competidor não encontrado",
		})
		return
	}

	// Verificar se o competidor pode se cadastrar
	if pode, motivo := competidor.PodeSeCadastrar(); !pode {
		c.JSON(http.StatusForbidden, gin.H{
			"error": motivo,
		})
		return
	}

	// Verificar se já existe inscrição
	var count int64
	database.DB.Model(&models.Inscricao{}).
		Where("etapa_id = ? AND competidor_id = ?", inscricao.EtapaID, inscricao.CompetidorID).
		Count(&count)

	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Competidor já inscrito nesta etapa",
		})
		return
	}

	// Definir valores padrão
	inscricao.DataInscricao = time.Now()
	inscricao.ValorPago = etapa.ValorInscricao
	inscricao.StatusPagamento = models.StatusPagamentoPendente

	// Criar inscrição
	result := database.DB.Create(&inscricao)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao criar inscrição: " + result.Error.Error(),
		})
		return
	}

	// Incrementar vagas ocupadas
	etapa.IncrementarVagas()
	database.DB.Save(&etapa)

	c.JSON(http.StatusCreated, inscricao)
}

// ConfirmarPagamento confirma o pagamento de uma inscrição
func ConfirmarPagamento(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}

	var input struct {
		ComprovantePgto string `json:"comprovante_pgto"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dados inválidos: " + err.Error(),
		})
		return
	}

	var inscricao models.Inscricao
	if err := database.DB.First(&inscricao, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Inscrição não encontrada",
		})
		return
	}

	inscricao.ConfirmarPagamento()
	inscricao.ComprovantePgto = input.ComprovantePgto
	database.DB.Save(&inscricao)

	c.JSON(http.StatusOK, gin.H{
		"message":   "Pagamento confirmado com sucesso",
		"inscricao": inscricao,
	})
}

// EliminarCompetidor elimina um competidor de uma etapa
func EliminarCompetidor(c *gin.Context) {
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
			"error": "Motivo da eliminação é obrigatório",
		})
		return
	}

	var inscricao models.Inscricao
	if err := database.DB.First(&inscricao, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Inscrição não encontrada",
		})
		return
	}

	inscricao.Eliminar(input.Motivo)
	database.DB.Save(&inscricao)

	c.JSON(http.StatusOK, gin.H{
		"message": "Competidor eliminado com sucesso",
	})
}

// DevolverRegua marca a régua como devolvida
func DevolverRegua(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}

	var inscricao models.Inscricao
	if err := database.DB.Preload("Regua").First(&inscricao, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Inscrição não encontrada",
		})
		return
	}

	if inscricao.NumeroReguaID == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nenhuma régua associada a esta inscrição",
		})
		return
	}

	inscricao.DevolverRegua()
	database.DB.Save(&inscricao)

	// Marcar régua como devolvida
	if inscricao.Regua != nil {
		inscricao.Regua.MarcarDevolvida()
		database.DB.Save(inscricao.Regua)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Régua devolvida com sucesso",
	})
}
