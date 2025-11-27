package models

type Usuario struct {
	BaseModel
	Nome    string `gorm:"size:100;not null" json:"nome" binding:"required"`
	Email   string `gorm:"size:100;uniqueIndex;not null" json:"email" binding:"required,email"`
	Senha   string `gorm:"size:255;not null" json:"-"`
	Tipo    string `gorm:"size:20;not null;default:'organizador'" json:"tipo"`
	Ativo   bool   `gorm:"default:true" json:"ativo"`
	FotoURL string `gorm:"size:500" json:"foto_url,omitempty"`
}

func (Usuario) TableName() string {
	return "usuarios"
}
