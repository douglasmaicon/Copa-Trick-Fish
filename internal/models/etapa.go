package models

import (
	"time"

	"github.com/google/uuid"
)

type Etapa struct {
	BaseModel
	EdicaoID          uuid.UUID   `gorm:"type:uuid;not null;index" json:"edicao_id" binding:"required"`
	Edicao            *Edicao     `gorm:"foreignKey:EdicaoID" json:"edicao,omitempty"`
	ModalidadeID      uuid.UUID   `gorm:"type:uuid;not null;index" json:"modalidade_id" binding:"required"`
	Modalidade        *Modalidade `gorm:"foreignKey:ModalidadeID" json:"modalidade,omitempty"`
	Numero            int         `gorm:"not null" json:"numero"` // número da etapa (1ª, 2ª, 3ª...)
	Nome              string      `gorm:"size:100;not null" json:"nome" binding:"required"`
	Local             string      `gorm:"size:200;not null" json:"local" binding:"required"`
	Latitude          float64     `json:"latitude,omitempty"`
	Longitude         float64     `json:"longitude,omitempty"`
	DataLargada       time.Time   `gorm:"not null" json:"data_largada" binding:"required"`
	HoraLargada       string      `gorm:"size:10" json:"hora_largada"` // "07:00"
	HoraRetorno       string      `gorm:"size:10;default:'16:00'" json:"hora_retorno"`
	ValorInscricao    float64     `gorm:"type:decimal(10,2)" json:"valor_inscricao"`
	VagasDisponiveis  int         `json:"vagas_disponiveis"`
	VagasOcupadas     int         `gorm:"default:0" json:"vagas_ocupadas"`
	Status            string      `gorm:"size:20;default:'aberta'" json:"status"` // aberta, em_andamento, finalizada, cancelada
	Regulamento       string      `gorm:"type:text" json:"regulamento"`
	ObservacoesGerais string      `gorm:"type:text" json:"observacoes_gerais"`
	ImagemURL         string      `gorm:"size:500" json:"imagem_url"`
	Inscricoes        []Inscricao `gorm:"foreignKey:EtapaID" json:"inscricoes,omitempty"`
	Reguas            []Regua     `gorm:"foreignKey:EtapaID" json:"reguas,omitempty"`
}

// PodeInscrever verifica se pode fazer inscrições
func (e *Etapa) PodeInscrever() (bool, string) {
	if !e.EstaAberta() {
		return false, "Etapa não está aberta para inscrições"
	}

	if !e.TemVagasDisponiveis() {
		return false, "Etapa sem vagas disponíveis"
	}

	// Verificar se a data de largada já passou
	if time.Now().After(e.DataLargada) {
		return false, "Data de largada já passou"
	}

	return true, ""
}

// EstaAberta verifica se a etapa está aberta para inscrições
func (e *Etapa) EstaAberta() bool {
	return e.Status == StatusEtapaAberta
}

// TemVagasDisponiveis verifica se ainda há vagas
func (e *Etapa) TemVagasDisponiveis() bool {
	// Se VagasDisponiveis = 0, significa sem limite
	if e.VagasDisponiveis == 0 {
		return true
	}
	return e.VagasOcupadas < e.VagasDisponiveis
}

// IncrementarVagas adiciona uma vaga ocupada
func (e *Etapa) IncrementarVagas() {
	e.VagasOcupadas++
}

// DecrementarVagas remove uma vaga ocupada
func (e *Etapa) DecrementarVagas() {
	if e.VagasOcupadas > 0 {
		e.VagasOcupadas--
	}
}

func (Etapa) TableName() string {
	return "etapas"
}
