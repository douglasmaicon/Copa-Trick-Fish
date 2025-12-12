package models

type Modalidade struct {
	BaseModel
	Nome      string `gorm:"size:50;not null;uniqueIndex" json:"nome" binding:"required"`
	Descricao string `gorm:"type:text" json:"descricao"`
	IconeURL  string `gorm:"size:500" json:"icone_url"`
	Ativa     bool   `gorm:"default:true" json:"ativa"`
	Ordem     int    `gorm:"default:0" json:"ordem"` // para ordenação na exibição
}

func (Modalidade) TableName() string {
	return "modalidades"
}
