# Script Simplificado - Copa Trick-Fish
# Versao sem caracteres especiais

Write-Host "Iniciando criacao dos models..." -ForegroundColor Green

# Criar diretorio
New-Item -ItemType Directory -Path "internal\models" -Force | Out-Null

# 1. BASE.GO
Write-Host "Criando base.go" -ForegroundColor Cyan
$base = @'
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (base *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	return nil
}

func (base *BaseModel) IsDeleted() bool {
	return base.DeletedAt.Valid
}

func (base *BaseModel) GetID() string {
	return base.ID.String()
}
'@
$base | Out-File "internal\models\base.go" -Encoding UTF8

# 2. CONSTANTS.GO
Write-Host "Criando constants.go" -ForegroundColor Cyan
$constants = @'
package models

const (
	TipoUsuarioAdmin       = "admin"
	TipoUsuarioOrganizador = "organizador"
	TipoUsuarioFiscal      = "fiscal"
)

const (
	StatusEtapaAberta      = "aberta"
	StatusEtapaEmAndamento = "em_andamento"
	StatusEtapaFinalizada  = "finalizada"
	StatusEtapaCancelada   = "cancelada"
)

const (
	StatusPagamentoPendente    = "pendente"
	StatusPagamentoPago        = "pago"
	StatusPagamentoCancelado   = "cancelado"
	StatusPagamentoReembolsado = "reembolsado"
)

const (
	EspecieTucunareAzul    = "tucunare_azul"
	EspecieTucunareAmarelo = "tucunare_amarelo"
	EspecieTraira          = "traira"
)

const (
	CategoriaGeral        = "geral"
	CategoriaMaiorAzul    = "maior_azul"
	CategoriaMaiorAmarelo = "maior_amarelo"
	CategoriaMaiorTraira  = "maior_traira"
)

const (
	TamanhoMinimoTucunare = 20.0
	CotaMaximaPeixes      = 4
	PenalidadeMinima      = 0.5
	PenalidadeMaxima      = 3.0
	HorarioLimiteRetorno  = "16:00"
)
'@
$constants | Out-File "internal\models\constants.go" -Encoding UTF8

# 3. USUARIO.GO
Write-Host "Criando usuario.go" -ForegroundColor Cyan
$usuario = @'
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
'@
$usuario | Out-File "internal\models\usuario.go" -Encoding UTF8

# 4-11. PLACEHOLDERS
Write-Host "Criando placeholders..." -ForegroundColor Cyan

"competidor", "edicao", "modalidade", "etapa", "inscricao", "regua", "captura", "ranking" | ForEach-Object {
    $name = $_
    $nameUpper = $name.Substring(0,1).ToUpper() + $name.Substring(1)
    $content = "package models

import `"time`"

type $nameUpper struct {
	BaseModel
}

func ($nameUpper) TableName() string {
	return `"${name}s`"
}
"
    $content | Out-File "internal\models\$name.go" -Encoding UTF8
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "SUCESSO! Arquivos criados:" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Get-ChildItem "internal\models\*.go" | ForEach-Object { Write-Host "  - $($_.Name)" -ForegroundColor Cyan }
Write-Host ""
Write-Host "Proximos passos:" -ForegroundColor Yellow
Write-Host "1. Abra os arquivos no VS Code"
Write-Host "2. Complete o conteudo dos placeholders"
Write-Host "3. Execute: go mod tidy"
Write-Host ""