package models

// Ranking representa a classificação
type Ranking struct {
	BaseModel
	EtapaID          string     `gorm:"type:uuid;not null;index" json:"etapa_id"`
	Etapa            *Etapa     `gorm:"foreignKey:EtapaID" json:"etapa,omitempty"`
	InscricaoID      string     `gorm:"type:uuid;not null;index" json:"inscricao_id"`
	Inscricao        *Inscricao `gorm:"foreignKey:InscricaoID" json:"inscricao,omitempty"`
	Posicao          int        `gorm:"not null;index" json:"posicao"`
	PontuacaoTotal   float64    `gorm:"type:decimal(10,2)" json:"pontuacao_total"`
	MaiorPeixe       float64    `gorm:"type:decimal(10,2)" json:"maior_peixe"`
	QuantidadePeixes int        `json:"quantidade_peixes"`
	Categoria        string     `gorm:"size:30;index" json:"categoria"`
	Premiacao        string     `gorm:"size:200" json:"premiacao,omitempty"`
	ValorPremiacao   float64    `gorm:"type:decimal(10,2)" json:"valor_premiacao,omitempty"`
}

func (Ranking) TableName() string {
	return "rankings"
}
