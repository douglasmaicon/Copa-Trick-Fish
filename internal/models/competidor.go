package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Competidor representa os pescadores que participam do torneio
type Competidor struct {
	BaseModel
	Nome            string     `gorm:"size:100;not null" json:"nome" binding:"required"`
	Email           string     `gorm:"size:100;uniqueIndex;not null" json:"email" binding:"required,email"`
	Senha           string     `gorm:"size:255;not null" json:"-"`
	Telefone        string     `gorm:"size:20" json:"telefone" binding:"required"`
	CPF             string     `gorm:"size:14;uniqueIndex" json:"cpf" binding:"required"`
	DataNascimento  *time.Time `json:"data_nascimento" binding:"required"`
	Cidade          string     `gorm:"size:100" json:"cidade" binding:"required"`
	Estado          string     `gorm:"size:2" json:"estado" binding:"required"`
	LicencaPesca    string     `gorm:"size:50" json:"licenca_pesca"`
	ValidadeLicenca *time.Time `json:"validade_licenca"`
	FotoURL         string     `gorm:"size:500" json:"foto_url"`
	Ativo           bool       `gorm:"default:true" json:"ativo"`
	Banido          bool       `gorm:"default:false" json:"banido"`
	MotivoBANimento string     `gorm:"type:text" json:"motivo_banimento,omitempty"`
	DataBanimento   *time.Time `json:"data_banimento,omitempty"`
	UltimoAcesso    *time.Time `json:"ultimo_acesso,omitempty"`

	// Relacionamentos
	Inscricoes []Inscricao `gorm:"foreignKey:CompetidorID" json:"inscricoes,omitempty"`
}

// TableName especifica o nome da tabela
func (Competidor) TableName() string {
	return "competidores"
}

// SetPassword gera o hash bcrypt da senha
func (c *Competidor) SetPassword(senha string) error {
	if len(senha) < 6 {
		return errors.New("senha deve ter no mínimo 6 caracteres")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	c.Senha = string(hashedPassword)
	return nil
}

// CheckPassword verifica se a senha está correta
func (c *Competidor) CheckPassword(senha string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(c.Senha), []byte(senha))
	return err == nil
}

// PodeSeCadastrar verifica se o competidor pode se cadastrar em etapas
func (c *Competidor) PodeSeCadastrar() (bool, string) {
	if !c.Ativo {
		return false, "Competidor inativo"
	}

	if c.Banido {
		return false, "Competidor banido do torneio"
	}

	if !c.LicencaValida() {
		return false, "Licença de pesca vencida ou não informada"
	}

	return true, ""
}

// LicencaValida verifica se a licença de pesca está válida
func (c *Competidor) LicencaValida() bool {
	if c.ValidadeLicenca == nil {
		return false
	}
	return c.ValidadeLicenca.After(time.Now())
}

// Idade calcula a idade do competidor
func (c *Competidor) Idade() int {
	if c.DataNascimento == nil {
		return 0
	}

	now := time.Now()
	age := now.Year() - c.DataNascimento.Year()

	if now.Month() < c.DataNascimento.Month() ||
		(now.Month() == c.DataNascimento.Month() && now.Day() < c.DataNascimento.Day()) {
		age--
	}

	return age
}

// EhInfantil verifica se é categoria infantil (até 16 anos)
func (c *Competidor) EhInfantil() bool {
	return c.Idade() <= 16
}

// Banir bane o competidor do torneio
func (c *Competidor) Banir(motivo string) {
	c.Banido = true
	c.MotivoBANimento = motivo
	now := time.Now()
	c.DataBanimento = &now
}

// Desbanir remove o banimento
func (c *Competidor) Desbanir() {
	c.Banido = false
	c.MotivoBANimento = ""
	c.DataBanimento = nil
}

// RegistrarAcesso atualiza o último acesso
func (c *Competidor) RegistrarAcesso() {
	now := time.Now()
	c.UltimoAcesso = &now
}
