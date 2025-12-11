package models

import "time"

// Captura representa cada peixe capturado
type Captura struct {
	BaseModel
	InscricaoID      string     `gorm:"type:uuid;not null;index" json:"inscricao_id" binding:"required"`
	Inscricao        *Inscricao `gorm:"foreignKey:InscricaoID" json:"inscricao,omitempty"`
	Especie          string     `gorm:"size:30;not null;index" json:"especie" binding:"required"`
	TamanhoOriginal  float64    `gorm:"type:decimal(10,2);not null" json:"tamanho_original" binding:"required,min=0"`
	Tamanho          float64    `gorm:"type:decimal(10,2);not null" json:"tamanho"`
	VideoURL         string     `gorm:"size:500;not null" json:"video_url" binding:"required"`
	ThumbnailURL     string     `gorm:"size:500" json:"thumbnail_url"`
	DuracaoVideo     int        `json:"duracao_video"`
	TamanhoArquivo   int64      `json:"tamanho_arquivo"`
	Validado         bool       `gorm:"default:false;index" json:"validado"`
	ValidadoPor      string     `gorm:"size:100" json:"validado_por,omitempty"`
	DataValidacao    *time.Time `json:"data_validacao,omitempty"`
	Penalidade       float64    `gorm:"type:decimal(10,2);default:0" json:"penalidade"`
	MotivoPenalidade string     `gorm:"type:text" json:"motivo_penalidade,omitempty"`
	Anulado          bool       `gorm:"default:false;index" json:"anulado"`
	MotivoAnulacao   string     `gorm:"type:text" json:"motivo_anulacao,omitempty"`
	HoraCaptura      time.Time  `gorm:"not null;index" json:"hora_captura"`
	Observacoes      string     `gorm:"type:text" json:"observacoes,omitempty"`
	ContaCota        bool       `gorm:"default:true" json:"conta_cota"`
}

// TableName especifica o nome da tabela
func (Captura) TableName() string {
	return "capturas"
}

// CalcularTamanhoFinal aplica penalidade
func (c *Captura) CalcularTamanhoFinal() float64 {
	tamanho := c.TamanhoOriginal - c.Penalidade
	if tamanho < 0 {
		return 0
	}
	return tamanho
}

// EstaValidada verifica se está validada e não anulada
func (c *Captura) EstaValidada() bool {
	return c.Validado && !c.Anulado
}

// Validar valida a captura
func (c *Captura) Validar(validador string, penalidade float64, motivo string) {
	c.Validado = true
	c.ValidadoPor = validador
	now := time.Now()
	c.DataValidacao = &now
	c.Penalidade = penalidade
	c.MotivoPenalidade = motivo
	c.Tamanho = c.CalcularTamanhoFinal()
}

// Anular anula a captura
func (c *Captura) Anular(motivo string) {
	c.Anulado = true
	c.MotivoAnulacao = motivo
}

// AtingeTamanhoMinimo verifica se atinge tamanho mínimo
func (c *Captura) AtingeTamanhoMinimo() bool {
	if c.Especie == EspecieTraira {
		return true // traíra não tem mínimo
	}
	return c.Tamanho >= TamanhoMinimoTucunare
}
