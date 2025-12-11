package models

import "time"

// Inscricao representa a inscrição de um competidor em uma etapa
type Inscricao struct {
	BaseModel
	EtapaID          string      `gorm:"type:uuid;not null;index" json:"etapa_id" binding:"required"`
	Etapa            *Etapa      `gorm:"foreignKey:EtapaID" json:"etapa,omitempty"`
	CompetidorID     string      `gorm:"type:uuid;not null;index" json:"competidor_id" binding:"required"`
	Competidor       *Competidor `gorm:"foreignKey:CompetidorID" json:"competidor,omitempty"`
	NumeroReguaID    *string     `gorm:"type:uuid;index" json:"numero_regua_id,omitempty"`
	Regua            *Regua      `gorm:"foreignKey:NumeroReguaID" json:"regua,omitempty"`
	DataInscricao    time.Time   `gorm:"autoCreateTime;not null" json:"data_inscricao"`
	ValorPago        float64     `gorm:"type:decimal(10,2)" json:"valor_pago"`
	StatusPagamento  string      `gorm:"size:20;default:'pendente';index" json:"status_pagamento"`
	DataPagamento    *time.Time  `json:"data_pagamento,omitempty"`
	ComprovantePgto  string      `gorm:"size:500" json:"comprovante_pgto,omitempty"`
	ReguaDevolvida   bool        `gorm:"default:false" json:"regua_devolvida"`
	DataDevolucao    *time.Time  `json:"data_devolucao,omitempty"`
	Eliminado        bool        `gorm:"default:false;index" json:"eliminado"`
	MotivoEliminacao string      `gorm:"type:text" json:"motivo_eliminacao,omitempty"`
	DataEliminacao   *time.Time  `json:"data_eliminacao,omitempty"`
	PontuacaoTotal   float64     `gorm:"type:decimal(10,2);default:0" json:"pontuacao_total"`
	QuantidadePeixes int         `gorm:"default:0" json:"quantidade_peixes"`

	// Relacionamentos
	Capturas []Captura `gorm:"foreignKey:InscricaoID" json:"capturas,omitempty"`
}

// TableName especifica o nome da tabela
func (Inscricao) TableName() string {
	return "inscricoes"
}

// PodeParticipar verifica se a inscrição está válida
func (i *Inscricao) PodeParticipar() (bool, string) {
	if i.StatusPagamento != StatusPagamentoPago {
		return false, "Pagamento pendente"
	}
	if i.Eliminado {
		return false, "Competidor eliminado"
	}
	return true, ""
}

// ConfirmarPagamento confirma o pagamento
func (i *Inscricao) ConfirmarPagamento() {
	i.StatusPagamento = StatusPagamentoPago
	now := time.Now()
	i.DataPagamento = &now
}

// Eliminar elimina o competidor
func (i *Inscricao) Eliminar(motivo string) {
	i.Eliminado = true
	i.MotivoEliminacao = motivo
	now := time.Now()
	i.DataEliminacao = &now
}

// DevolverRegua marca a régua como devolvida
func (i *Inscricao) DevolverRegua() {
	i.ReguaDevolvida = true
	now := time.Now()
	i.DataDevolucao = &now
}

// CalcularPontuacao calcula a pontuação total
func (i *Inscricao) CalcularPontuacao() float64 {
	total := 0.0
	count := 0

	for _, captura := range i.Capturas {
		if captura.EstaValidada() && captura.ContaCota && count < CotaMaximaPeixes {
			total += captura.Tamanho
			count++
		}
	}

	i.PontuacaoTotal = total
	i.QuantidadePeixes = count
	return total
}
