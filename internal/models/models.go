package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel contém campos comuns a todos os modelos
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate hook para gerar UUID antes de criar
func (base *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	return nil
}

// ============================================
// USUÁRIOS E AUTENTICAÇÃO
// ============================================

// Usuario representa um usuário do sistema (admin, organizador, fiscal)
type Usuario struct {
	BaseModel
	Nome    string `gorm:"size:100;not null" json:"nome" binding:"required"`
	Email   string `gorm:"size:100;uniqueIndex;not null" json:"email" binding:"required,email"`
	Senha   string `gorm:"size:255;not null" json:"-"`                         // nunca expor em JSON
	Tipo    string `gorm:"size:20;not null;default:'organizador'" json:"tipo"` // admin, organizador, fiscal
	Ativo   bool   `gorm:"default:true" json:"ativo"`
	FotoURL string `gorm:"size:500" json:"foto_url,omitempty"`
}

// TableName especifica o nome da tabela
func (Usuario) TableName() string {
	return "usuarios"
}

// ============================================
// ESTRUTURA DO TORNEIO
// ============================================

// Edicao representa um ano/temporada do torneio
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

// Modalidade representa tipos de competição (embarcada, caiaque, casais, feminino, infantil)
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

// Etapa representa cada evento da edição
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

func (Etapa) TableName() string {
	return "etapas"
}

// TemVagasDisponiveis verifica se ainda há vagas
func (e *Etapa) TemVagasDisponiveis() bool {
	if e.VagasDisponiveis == 0 {
		return true // sem limite
	}
	return e.VagasOcupadas < e.VagasDisponiveis
}

// ============================================
// COMPETIDORES
// ============================================

// Competidor representa os pescadores
type Competidor struct {
	BaseModel
	Nome            string      `gorm:"size:100;not null" json:"nome" binding:"required"`
	Email           string      `gorm:"size:100;uniqueIndex;not null" json:"email" binding:"required,email"`
	Senha           string      `gorm:"size:255;not null" json:"-"`
	Telefone        string      `gorm:"size:20" json:"telefone"`
	CPF             string      `gorm:"size:14;uniqueIndex" json:"cpf"`
	DataNascimento  *time.Time  `json:"data_nascimento"`
	Cidade          string      `gorm:"size:100" json:"cidade"`
	Estado          string      `gorm:"size:2" json:"estado"` // UF
	LicencaPesca    string      `gorm:"size:50" json:"licenca_pesca"`
	ValidadeLicenca *time.Time  `json:"validade_licenca"`
	FotoURL         string      `gorm:"size:500" json:"foto_url"`
	Ativo           bool        `gorm:"default:true" json:"ativo"`
	Banido          bool        `gorm:"default:false" json:"banido"`
	MotivoBANimento string      `gorm:"type:text" json:"motivo_banimento,omitempty"`
	DataBanimento   *time.Time  `json:"data_banimento,omitempty"`
	Inscricoes      []Inscricao `gorm:"foreignKey:CompetidorID" json:"inscricoes,omitempty"`
}

func (Competidor) TableName() string {
	return "competidores"
}

// PodeSeCadastrar verifica se o competidor pode se cadastrar em etapas
func (c *Competidor) PodeSeCadastrar() bool {
	return c.Ativo && !c.Banido
}

// LicencaValida verifica se a licença de pesca está válida
func (c *Competidor) LicencaValida() bool {
	if c.ValidadeLicenca == nil {
		return false
	}
	return c.ValidadeLicenca.After(time.Now())
}

// ============================================
// INSCRIÇÕES E RÉGUAS
// ============================================

// Inscricao representa a inscrição de um competidor em uma etapa
type Inscricao struct {
	BaseModel
	EtapaID          uuid.UUID   `gorm:"type:uuid;not null;index" json:"etapa_id" binding:"required"`
	Etapa            *Etapa      `gorm:"foreignKey:EtapaID" json:"etapa,omitempty"`
	CompetidorID     uuid.UUID   `gorm:"type:uuid;not null;index" json:"competidor_id" binding:"required"`
	Competidor       *Competidor `gorm:"foreignKey:CompetidorID" json:"competidor,omitempty"`
	NumeroReguaID    *uuid.UUID  `gorm:"type:uuid;index" json:"numero_regua_id,omitempty"`
	Regua            *Regua      `gorm:"foreignKey:NumeroReguaID" json:"regua,omitempty"`
	DataInscricao    time.Time   `gorm:"autoCreateTime;not null" json:"data_inscricao"`
	ValorPago        float64     `gorm:"type:decimal(10,2)" json:"valor_pago"`
	StatusPagamento  string      `gorm:"size:20;default:'pendente'" json:"status_pagamento"` // pendente, pago, cancelado, reembolsado
	DataPagamento    *time.Time  `json:"data_pagamento,omitempty"`
	ComprovantePgto  string      `gorm:"size:500" json:"comprovante_pgto,omitempty"`
	ReguaDevolvida   bool        `gorm:"default:false" json:"regua_devolvida"`
	DataDevolucao    *time.Time  `json:"data_devolucao,omitempty"`
	Eliminado        bool        `gorm:"default:false" json:"eliminado"`
	MotivoEliminacao string      `gorm:"type:text" json:"motivo_eliminacao,omitempty"`
	DataEliminacao   *time.Time  `json:"data_eliminacao,omitempty"`
	PontuacaoTotal   float64     `gorm:"type:decimal(10,2);default:0" json:"pontuacao_total"`
	QuantidadePeixes int         `gorm:"default:0" json:"quantidade_peixes"`
	Capturas         []Captura   `gorm:"foreignKey:InscricaoID" json:"capturas,omitempty"`
}

func (Inscricao) TableName() string {
	return "inscricoes"
}

// PodeParticipar verifica se a inscrição está válida para participação
func (i *Inscricao) PodeParticipar() bool {
	return i.StatusPagamento == "pago" && !i.Eliminado
}

// Regua representa as réguas numeradas para validação de vídeos
type Regua struct {
	BaseModel
	EtapaID    uuid.UUID `gorm:"type:uuid;not null;index" json:"etapa_id" binding:"required"`
	Etapa      *Etapa    `gorm:"foreignKey:EtapaID" json:"etapa,omitempty"`
	Numero     int       `gorm:"not null" json:"numero" binding:"required"`
	Disponivel bool      `gorm:"default:true" json:"disponivel"`
	Devolvida  bool      `gorm:"default:false" json:"devolvida"`
}

func (Regua) TableName() string {
	return "reguas"
}

// ============================================
// CAPTURAS E PEIXES
// ============================================

// Captura representa cada peixe capturado e registrado
type Captura struct {
	BaseModel
	InscricaoID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"inscricao_id" binding:"required"`
	Inscricao        *Inscricao `gorm:"foreignKey:InscricaoID" json:"inscricao,omitempty"`
	Especie          string     `gorm:"size:30;not null" json:"especie" binding:"required"` // tucunare_azul, tucunare_amarelo, traira
	TamanhoOriginal  float64    `gorm:"type:decimal(10,2);not null" json:"tamanho_original" binding:"required"`
	Tamanho          float64    `gorm:"type:decimal(10,2);not null" json:"tamanho"` // após penalidades
	VideoURL         string     `gorm:"size:500;not null" json:"video_url" binding:"required"`
	ThumbnailURL     string     `gorm:"size:500" json:"thumbnail_url"`
	DuracaoVideo     int        `json:"duracao_video"`   // em segundos
	TamanhoArquivo   int64      `json:"tamanho_arquivo"` // em bytes
	Validado         bool       `gorm:"default:false" json:"validado"`
	ValidadoPor      string     `gorm:"size:100" json:"validado_por,omitempty"`
	DataValidacao    *time.Time `json:"data_validacao,omitempty"`
	Penalidade       float64    `gorm:"type:decimal(10,2);default:0" json:"penalidade"`
	MotivoPenalidade string     `gorm:"type:text" json:"motivo_penalidade,omitempty"`
	Anulado          bool       `gorm:"default:false" json:"anulado"`
	MotivoAnulacao   string     `gorm:"type:text" json:"motivo_anulacao,omitempty"`
	HoraCaptura      time.Time  `gorm:"not null" json:"hora_captura"`
	Observacoes      string     `gorm:"type:text" json:"observacoes,omitempty"`
	ContaCota        bool       `gorm:"default:true" json:"conta_cota"` // se conta para a cota de 4 peixes
}

func (Captura) TableName() string {
	return "capturas"
}

// CalcularTamanhoFinal aplica a penalidade ao tamanho original
func (c *Captura) CalcularTamanhoFinal() float64 {
	tamanho := c.TamanhoOriginal - c.Penalidade
	if tamanho < 0 {
		return 0
	}
	return tamanho
}

// EstaValidada verifica se a captura foi validada e não anulada
func (c *Captura) EstaValidada() bool {
	return c.Validado && !c.Anulado
}

// ============================================
// RANKING E PREMIAÇÕES
// ============================================

// Ranking representa a classificação em uma etapa
type Ranking struct {
	BaseModel
	EtapaID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"etapa_id"`
	Etapa            *Etapa     `gorm:"foreignKey:EtapaID" json:"etapa,omitempty"`
	InscricaoID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"inscricao_id"`
	Inscricao        *Inscricao `gorm:"foreignKey:InscricaoID" json:"inscricao,omitempty"`
	Posicao          int        `gorm:"not null" json:"posicao"`
	PontuacaoTotal   float64    `gorm:"type:decimal(10,2)" json:"pontuacao_total"`
	MaiorPeixe       float64    `gorm:"type:decimal(10,2)" json:"maior_peixe"`
	QuantidadePeixes int        `json:"quantidade_peixes"`
	Categoria        string     `gorm:"size:30" json:"categoria"` // geral, maior_azul, maior_amarelo, maior_traira
	Premiacao        string     `gorm:"size:200" json:"premiacao,omitempty"`
	ValorPremiacao   float64    `gorm:"type:decimal(10,2)" json:"valor_premiacao,omitempty"`
}

func (Ranking) TableName() string {
	return "rankings"
}

// ============================================
// CONSTANTES E ENUMS
// ============================================

const (
	// Tipos de usuário
	TipoUsuarioAdmin       = "admin"
	TipoUsuarioOrganizador = "organizador"
	TipoUsuarioFiscal      = "fiscal"

	// Status de etapa
	StatusEtapaAberta      = "aberta"
	StatusEtapaEmAndamento = "em_andamento"
	StatusEtapaFinalizada  = "finalizada"
	StatusEtapaCancelada   = "cancelada"

	// Status de pagamento
	StatusPagamentoPendente    = "pendente"
	StatusPagamentoPago        = "pago"
	StatusPagamentoCancelado   = "cancelado"
	StatusPagamentoReembolsado = "reembolsado"

	// Espécies de peixe
	EspecieTucunareAzul    = "tucunare_azul"
	EspecieTucunareAmarelo = "tucunare_amarelo"
	EspecieTraira          = "traira"

	// Categorias de ranking
	CategoriaGeral        = "geral"
	CategoriaMaiorAzul    = "maior_azul"
	CategoriaMaiorAmarelo = "maior_amarelo"
	CategoriaMaiorTraira  = "maior_traira"

	// Limites
	TamanhoMinimoTucunare = 20.0 // cm
	CotaMaximaPeixes      = 4
	PenalidadeMinima      = 0.5 // cm
	PenalidadeMaxima      = 3.0 // cm
)
