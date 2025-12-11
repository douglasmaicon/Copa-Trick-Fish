package models

// Regua representa as réguas numeradas
type Regua struct {
	BaseModel
	EtapaID    string `gorm:"type:uuid;not null;index" json:"etapa_id" binding:"required"`
	Etapa      *Etapa `gorm:"foreignKey:EtapaID" json:"etapa,omitempty"`
	Numero     int    `gorm:"not null" json:"numero" binding:"required,min=1"`
	Disponivel bool   `gorm:"default:true;index" json:"disponivel"`
	Devolvida  bool   `gorm:"default:false" json:"devolvida"`
}

// TableName especifica o nome da tabela
func (Regua) TableName() string {
	return "reguas"
}

// Alocar marca a régua como alocada
func (r *Regua) Alocar() {
	r.Disponivel = false
}

// Liberar marca a régua como disponível
func (r *Regua) Liberar() {
	r.Disponivel = true
}

// MarcarDevolvida marca como devolvida
func (r *Regua) MarcarDevolvida() {
	r.Devolvida = true
}
