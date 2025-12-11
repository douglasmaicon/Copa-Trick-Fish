package handlers

import (
	"net/http"
	"time"

	"github.com/douglasmaicon/Copa-Trick-Fish/internal/database"
	"github.com/douglasmaicon/Copa-Trick-Fish/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListarCapturas retorna todas as capturas com filtros
func ListarCapturas(c *gin.Context) {
	etapaID := c.Query("etapa_id")
	inscricaoID := c.Query("inscricao_id")
	validado := c.Query("validado")
	especie := c.Query("especie")

	var capturas []models.Captura
	query := database.DB.Preload("Inscricao.Competidor").Preload("Inscricao.Etapa")

	if inscricaoID != "" {
		query = query.Where("inscricao_id = ?", inscricaoID)
	}

	if etapaID != "" {
		query = query.Joins("JOIN inscricoes ON capturas.inscricao_id = inscricoes.id").
			Where("inscricoes.etapa_id = ?", etapaID)
	}

	if validado != "" {
		query = query.Where("validado = ?", validado == "true")
	}

	if especie != "" {
		query = query.Where("especie = ?", especie)
	}

	result := query.Order("hora_captura DESC").Find(&capturas)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao buscar capturas",
		})
		return
	}

	c.JSON(http.StatusOK, capturas)
}

// BuscarCaptura retorna uma captura específica
func BuscarCaptura(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}

	var captura models.Captura
	result := database.DB.
		Preload("Inscricao.Competidor").
		Preload("Inscricao.Etapa").
		First(&captura, "id = ?", id)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Captura não encontrada",
		})
		return
	}

	c.JSON(http.StatusOK, captura)
}

// CriarCaptura registra uma nova captura
func CriarCaptura(c *gin.Context) {
	var captura models.Captura

	if err := c.ShouldBindJSON(&captura); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dados inválidos: " + err.Error(),
		})
		return
	}

	// Verificar se a inscrição existe
	var inscricao models.Inscricao
	if err := database.DB.Preload("Etapa").First(&inscricao, "id = ?", captura.InscricaoID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Inscrição não encontrada",
		})
		return
	}

	// Verificar se pode participar
	if pode, motivo := inscricao.PodeParticipar(); !pode {
		c.JSON(http.StatusForbidden, gin.H{
			"error": motivo,
		})
		return
	}

	// Validar espécie
	if !models.ValidarEspecie(captura.Especie) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Espécie inválida",
		})
		return
	}

	// Verificar cota de peixes (apenas tucunarés)
	if captura.Especie != models.EspecieTraira {
		var count int64
		database.DB.Model(&models.Captura{}).
			Where("inscricao_id = ? AND validado = ? AND anulado = ? AND especie != ?",
				captura.InscricaoID, true, false, models.EspecieTraira).
			Count(&count)

		if count >= models.CotaMaximaPeixes {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Cota máxima de peixes atingida (4 tucunarés)",
			})
			return
		}
	}

	// Definir valores padrão
	captura.Tamanho = captura.TamanhoOriginal
	captura.Validado = false
	captura.Anulado = false

	if captura.HoraCaptura.IsZero() {
		captura.HoraCaptura = time.Now()
	}

	// Verificar se está dentro do horário permitido
	if inscricao.Etapa != nil {
		// Aqui você pode adicionar validação de horário de retorno
		// Por exemplo: captura.HoraCaptura deve ser antes das 16h
	}

	result := database.DB.Create(&captura)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao registrar captura: " + result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, captura)
}

// ValidarCaptura valida uma captura e aplica penalidades
func ValidarCaptura(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}

	var input struct {
		Penalidade       float64 `json:"penalidade" binding:"min=0,max=3"`
		MotivoPenalidade string  `json:"motivo_penalidade"`
		ValidadoPor      string  `json:"validado_por" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dados inválidos: " + err.Error(),
		})
		return
	}

	// Validar penalidade
	if !models.ValidarPenalidade(input.Penalidade) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Penalidade deve estar entre 0 e 3 cm",
		})
		return
	}

	var captura models.Captura
	if err := database.DB.First(&captura, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Captura não encontrada",
		})
		return
	}

	if captura.Validado {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Captura já foi validada",
		})
		return
	}

	// Validar captura
	captura.Validar(input.ValidadoPor, input.Penalidade, input.MotivoPenalidade)

	// Verificar tamanho mínimo
	if !captura.AtingeTamanhoMinimo() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Peixe abaixo do tamanho mínimo após penalidade",
			"tamanho": captura.Tamanho,
			"minimo":  models.TamanhoMinimoTucunare,
		})
		return
	}

	database.DB.Save(&captura)

	// Atualizar pontuação da inscrição
	var inscricao models.Inscricao
	if err := database.DB.Preload("Capturas").First(&inscricao, "id = ?", captura.InscricaoID).Error; err == nil {
		inscricao.CalcularPontuacao()
		database.DB.Save(&inscricao)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Captura validada com sucesso",
		"captura": captura,
	})
}

// AnularCaptura anula uma captura
func AnularCaptura(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}

	var input struct {
		MotivoAnulacao string `json:"motivo_anulacao" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Motivo da anulação é obrigatório",
		})
		return
	}

	var captura models.Captura
	if err := database.DB.First(&captura, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Captura não encontrada",
		})
		return
	}

	captura.Anular(input.MotivoAnulacao)
	database.DB.Save(&captura)

	// Atualizar pontuação da inscrição
	var inscricao models.Inscricao
	if err := database.DB.Preload("Capturas").First(&inscricao, "id = ?", captura.InscricaoID).Error; err == nil {
		inscricao.CalcularPontuacao()
		database.DB.Save(&inscricao)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Captura anulada com sucesso",
	})
}

// DeletarCaptura remove uma captura
func DeletarCaptura(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}

	result := database.DB.Delete(&models.Captura{}, "id = ?", id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao deletar captura",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Captura não encontrada",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Captura deletada com sucesso",
	})
}
