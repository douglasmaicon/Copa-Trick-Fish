package handlers

import (
	"net/http"

	"github.com/douglasmaicon/Copa-Trick-Fish/internal/database"
	"github.com/douglasmaicon/Copa-Trick-Fish/internal/models"
	"github.com/gin-gonic/gin"
)

// ListarModalidades retorna todas as modalidades ativas
func ListarModalidades(c *gin.Context) {
	var modalidades []models.Modalidade

	result := database.DB.Where("ativa = ?", true).
		Order("ordem ASC").
		Find(&modalidades)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao buscar modalidades",
		})
		return
	}

	c.JSON(http.StatusOK, modalidades)
}
