package models

type Edicao struct {
	BaseModel
	Ano         int          `gorm:"not null;uniqueIndex" json:"ano" binding:"required"`
	Nome        string       `gorm:"size:100;not null" json:"nome" binding:"required"`
	Descricao   string       `gorm:"type:text" json:"descricao"`
	ImagemURL   string       `gorm:"size:500" json:"imagem_url"`
	Ativa       bool         `gorm:"default:true" json:"ativa"`
	Etapas      []Etapa      `gorm:"foreignKey:EdicaoID" json:"etapas,omitempty"`
	Modalidades []Modalidade `gorm:"many2many:edicao_modalidades;" json:"modalidades,omitempty"`
}

func (Edicao) TableName() string {
	return "edicoes"
}
