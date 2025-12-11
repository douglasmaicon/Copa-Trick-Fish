package handlers

import (
	"net/http"

	"github.com/douglasmaicon/Copa-Trick-Fish/internal/database"
	"github.com/douglasmaicon/Copa-Trick-Fish/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListarReguas retorna todas as réguas de uma etapa
func ListarReguas(c *gin.Context) {
	etapaID := c.Query("etapa_id")
	disponivel := c.Query("disponivel")

	if etapaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "etapa_id é obrigatório",
		})
		return
	}

	var reguas []models.Regua
	query := database.DB.Where("etapa_id = ?", etapaID)

	if disponivel != "" {
		query = query.Where("disponivel = ?", disponivel == "true")
	}

	result := query.Order("numero ASC").Find(&reguas)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao buscar réguas",
		})
		return
	}

	c.JSON(http.StatusOK, reguas)
}

// GerarReguas gera réguas numeradas para uma etapa
func GerarReguas(c *gin.Context) {
	var input struct {
		EtapaID    string `json:"etapa_id" binding:"required"`
		Quantidade int    `json:"quantidade" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dados inválidos: " + err.Error(),
		})
		return
	}

	// Verificar se a etapa existe
	var etapa models.Etapa
	if err := database.DB.First(&etapa, "id = ?", input.EtapaID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Etapa não encontrada",
		})
		return
	}

	// Verificar quantas réguas já existem
	var count int64
	database.DB.Model(&models.Regua{}).Where("etapa_id = ?", input.EtapaID).Count(&count)

	numeroInicial := int(count) + 1
	reguas := []models.Regua{}

	// Criar as réguas
	for i := 0; i < input.Quantidade; i++ {
		regua := models.Regua{
			EtapaID:    input.EtapaID,
			Numero:     numeroInicial + i,
			Disponivel: true,
			Devolvida:  false,
		}
		reguas = append(reguas, regua)
	}

	// Salvar em lote
	result := database.DB.Create(&reguas)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao gerar réguas: " + result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Réguas geradas com sucesso",
		"quantidade": input.Quantidade,
		"reguas":     reguas,
	})
}

// SortearReguas sorteia réguas para as inscrições de uma etapa
func SortearReguas(c *gin.Context) {
	var input struct {
		EtapaID string `json:"etapa_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dados inválidos: " + err.Error(),
		})
		return
	}

	// Buscar inscrições sem régua
	var inscricoes []models.Inscricao
	database.DB.Where("etapa_id = ? AND numero_regua_id IS NULL", input.EtapaID).
		Find(&inscricoes)

	if len(inscricoes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nenhuma inscrição sem régua encontrada",
		})
		return
	}

	// Buscar réguas disponíveis
	var reguas []models.Regua
	database.DB.Where("etapa_id = ? AND disponivel = ?", input.EtapaID, true).
		Order("numero ASC").
		Find(&reguas)

	if len(reguas) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nenhuma régua disponível. Gere réguas primeiro.",
		})
		return
	}

	if len(reguas) < len(inscricoes) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Réguas insuficientes. Gere mais réguas.",
		})
		return
	}

	// Sortear réguas para inscrições
	sorteadas := 0
	for i, inscricao := range inscricoes {
		if i < len(reguas) {
			reguaID := reguas[i].ID.String()
			inscricao.NumeroReguaID = &reguaID
			database.DB.Save(&inscricao)

			// Marcar régua como alocada
			reguas[i].Alocar()
			database.DB.Save(&reguas[i])

			sorteadas++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Réguas sorteadas com sucesso",
		"total_sorteadas":  sorteadas,
		"total_inscricoes": len(inscricoes),
	})
}

// DeletarRegua remove uma régua
func DeletarRegua(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}

	// Verificar se a régua está em uso
	var count int64
	database.DB.Model(&models.Inscricao{}).
		Where("numero_regua_id = ?", id).
		Count(&count)

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Régua está em uso e não pode ser deletada",
		})
		return
	}

	result := database.DB.Delete(&models.Regua{}, "id = ?", id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao deletar régua",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Régua não encontrada",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Régua deletada com sucesso",
	})
}
