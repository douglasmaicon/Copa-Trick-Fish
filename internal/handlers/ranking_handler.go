package handlers

import (
	"net/http"

	"github.com/douglasmaicon/Copa-Trick-Fish/internal/database"
	"github.com/douglasmaicon/Copa-Trick-Fish/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// BuscarRankingEtapa retorna o ranking de uma etapa
func BuscarRankingEtapa(c *gin.Context) {
	etapaID := c.Param("id")
	categoria := c.Query("categoria") // geral, maior_azul, maior_amarelo, maior_traira
	
	if _, err := uuid.Parse(etapaID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}
	
	// Verificar se a etapa existe
	var etapa models.Etapa
	if err := database.DB.First(&etapa, "id = ?", etapaID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Etapa não encontrada",
		})
		return
	}
	
	var rankings []models.Ranking
	query := database.DB.Where("etapa_id = ?", etapaID).
		Preload("Inscricao.Competidor").
		Preload("Etapa")
	
	if categoria != "" {
		query = query.Where("categoria = ?", categoria)
	}
	
	result := query.Order("posicao ASC").Find(&rankings)
	
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao buscar ranking",
		})
		return
	}
	
	c.JSON(http.StatusOK, rankings)
}

// GerarRanking gera/atualiza o ranking de uma etapa
func GerarRanking(c *gin.Context) {
	etapaID := c.Param("id")
	
	if _, err := uuid.Parse(etapaID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}
	
	// Verificar se a etapa existe
	var etapa models.Etapa
	if err := database.DB.First(&etapa, "id = ?", etapaID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Etapa não encontrada",
		})
		return
	}
	
	// Limpar ranking anterior
	database.DB.Where("etapa_id = ?", etapaID).Delete(&models.Ranking{})
	
	// Buscar todas as inscrições válidas com capturas
	var inscricoes []models.Inscricao
	database.DB.Where("etapa_id = ? AND status_pagamento = ? AND eliminado = ?", 
		etapaID, models.StatusPagamentoPago, false).
		Preload("Competidor").
		Preload("Capturas", "validado = ? AND anulado = ?", true, false).
		Find(&inscricoes)
	
	if len(inscricoes) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Nenhuma inscrição válida encontrada",
		})
		return
	}
	
	// Calcular pontuações e criar ranking geral
	type RankingTemp struct {
		Inscricao        *models.Inscricao
		PontuacaoTotal   float64
		MaiorPeixe       float64
		QuantidadePeixes int
	}
	
	var rankingTemp []RankingTemp
	
	for i := range inscricoes {
		inscricao := &inscricoes[i]
		pontuacao := inscricao.CalcularPontuacao()
		database.DB.Save(inscricao)
		
		// Encontrar maior peixe
		maiorPeixe := 0.0
		for _, captura := range inscricao.Capturas {
			if captura.Tamanho > maiorPeixe {
				maiorPeixe = captura.Tamanho
			}
		}
		
		rankingTemp = append(rankingTemp, RankingTemp{
			Inscricao:        inscricao,
			PontuacaoTotal:   pontuacao,
			MaiorPeixe:       maiorPeixe,
			QuantidadePeixes: inscricao.QuantidadePeixes,
		})
	}
	
	// Ordenar por pontuação (maior primeiro)
	for i := 0; i < len(rankingTemp); i++ {
		for j := i + 1; j < len(rankingTemp); j++ {
			if rankingTemp[j].PontuacaoTotal > rankingTemp[i].PontuacaoTotal {
				rankingTemp[i], rankingTemp[j] = rankingTemp[j], rankingTemp[i]
			}
		}
	}
	
	// Criar ranking geral
	for i, rt := range rankingTemp {
		ranking := models.Ranking{
			EtapaID:          etapaID,
			InscricaoID:      rt.Inscricao.ID.String(),
			Posicao:          i + 1,
			PontuacaoTotal:   rt.PontuacaoTotal,
			MaiorPeixe:       rt.MaiorPeixe,
			QuantidadePeixes: rt.QuantidadePeixes,
			Categoria:        models.CategoriaGeral,
		}
		
		// Definir premiação para os 3 primeiros
		if i == 0 {
			ranking.Premiacao = "1º Lugar"
		} else if i == 1 {
			ranking.Premiacao = "2º Lugar"
		} else if i == 2 {
			ranking.Premiacao = "3º Lugar"
		}
		
		database.DB.Create(&ranking)
	}
	
	// Gerar rankings de maiores peixes por espécie
	gerarRankingMaiorPeixe(etapaID, models.EspecieTucunareAzul, models.CategoriaMaiorAzul)
	gerarRankingMaiorPeixe(etapaID, models.EspecieTucunareAmarelo, models.CategoriaMaiorAmarelo)
	gerarRankingMaiorPeixe(etapaID, models.EspecieTraira, models.CategoriaMaiorTraira)
	
	c.JSON(http.StatusOK, gin.H{
		"message":           "Ranking gerado com sucesso",
		"total_competidores": len(rankingTemp),
	})
}

// gerarRankingMaiorPeixe gera ranking do maior peixe de uma espécie
func gerarRankingMaiorPeixe(etapaID string, especie string, categoria string) {
	// Buscar a maior captura da espécie
	var captura models.Captura
	err := database.DB.
		Joins("JOIN inscricoes ON capturas.inscricao_id = inscricoes.id").
		Where("inscricoes.etapa_id = ? AND capturas.especie = ? AND capturas.validado = ? AND capturas.anulado = ?",
			etapaID, especie, true, false).
		Order("capturas.tamanho DESC").
		Preload("Inscricao.Competidor").
		First(&captura).Error
	
	if err != nil {
		return // Nenhuma captura desta espécie
	}
	
	// Criar ranking
	ranking := models.Ranking{
		EtapaID:          etapaID,
		InscricaoID:      captura.InscricaoID,
		Posicao:          1,
		PontuacaoTotal:   0,
		MaiorPeixe:       captura.Tamanho,
		QuantidadePeixes: 1,
		Categoria:        categoria,
		Premiacao:        "Maior " + models.GetNomeEspecie(especie),
	}
	
	database.DB.Create(&ranking)
}

// ListarRankings retorna rankings com filtros
func ListarRankings(c *gin.Context) {
	etapaID := c.Query("etapa_id")
	edicaoID := c.Query("edicao_id")
	competidorID := c.Query("competidor_id")
	
	var rankings []models.Ranking
	query := database.DB.Preload("Inscricao.Competidor").Preload("Etapa")
	
	if etapaID != "" {
		query = query.Where("etapa_id = ?", etapaID)
	}
	
	if edicaoID != "" {
		query = query.Joins("JOIN etapas ON rankings.etapa_id = etapas.id").
			Where("etapas.edicao_id = ?", edicaoID)
	}
	
	if competidorID != "" {
		query = query.Joins("JOIN inscricoes ON rankings.inscricao_id = inscricoes.id").
			Where("inscricoes.competidor_id = ?", competidorID)
	}
	
	result := query.Order("posicao ASC").Find(&rankings)
	
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao buscar rankings",
		})
		return
	}
	
	c.JSON(http.StatusOK, rankings)
}

// DeletarRanking deleta um ranking específico
func DeletarRanking(c *gin.Context) {
	id := c.Param("id")
	
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}
	
	result := database.DB.Delete(&models.Ranking{}, "id = ?", id)
	
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao deletar ranking",
		})
		return
	}
	
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Ranking não encontrado",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Ranking deletado com sucesso",
	})
}