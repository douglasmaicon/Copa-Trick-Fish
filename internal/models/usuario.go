package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Usuario struct {
	BaseModel
	Nome            string     `gorm:"size:100;not null" json:"nome" binding:"required"`
	Email           string     `gorm:"size:100;uniqueIndex;not null" json:"email" binding:"required,email"`
	Senha           string     `gorm:"size:255;not null" json:"-"`
	Tipo            string     `gorm:"size:20;not null;default:'organizador'" json:"tipo"`
	Ativo           bool       `gorm:"default:true" json:"ativo"`
	FotoURL         string     `gorm:"size:500" json:"foto_url,omitempty"`
	UltimoAcesso    *time.Time `json:"ultimo_acesso,omitempty"`
	TentativasLogin int        `gorm:"default:0" json:"-"`
	BloqueadoAte    *time.Time `json:"-"`
}

func (Usuario) TableName() string {
	return "usuarios"
}

func (u *Usuario) SetPassword(senha string) error {
	if len(senha) < 6 {
		return errors.New("senha deve ter no minimo 6 caracteres")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Senha = string(hashedPassword)
	return nil
}

func (u *Usuario) CheckPassword(senha string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Senha), []byte(senha))
	return err == nil
}

func (u *Usuario) IsAdmin() bool {
	return u.Tipo == TipoUsuarioAdmin
}

func (u *Usuario) PodeAcessar() bool {
	if !u.Ativo {
		return false
	}
	if u.BloqueadoAte != nil && u.BloqueadoAte.After(time.Now()) {
		return false
	}
	return true
}

// RegistrarAcesso atualiza o último acesso
func (u *Usuario) RegistrarAcesso() {
	now := time.Now()
	u.UltimoAcesso = &now
	u.TentativasLogin = 0 // resetar tentativas
}

// RegistrarFalhaLogin incrementa tentativas e bloqueia se necessário
func (u *Usuario) RegistrarFalhaLogin() {
	u.TentativasLogin++

	// Bloquear após 5 tentativas por 15 minutos
	if u.TentativasLogin >= 5 {
		bloqueio := time.Now().Add(15 * time.Minute)
		u.BloqueadoAte = &bloqueio
	}
}
