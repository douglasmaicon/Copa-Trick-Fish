package models

// ============================================
// TIPOS DE USUÁRIO
// ============================================

const (
	TipoUsuarioAdmin       = "admin"
	TipoUsuarioOrganizador = "organizador"
	TipoUsuarioFiscal      = "fiscal"
)

// GetTiposUsuario retorna todos os tipos de usuário válidos
func GetTiposUsuario() []string {
	return []string{
		TipoUsuarioAdmin,
		TipoUsuarioOrganizador,
		TipoUsuarioFiscal,
	}
}

// ============================================
// STATUS DE ETAPA
// ============================================

const (
	StatusEtapaAberta      = "aberta"
	StatusEtapaEmAndamento = "em_andamento"
	StatusEtapaFinalizada  = "finalizada"
	StatusEtapaCancelada   = "cancelada"
)

// GetStatusEtapa retorna todos os status de etapa válidos
func GetStatusEtapa() []string {
	return []string{
		StatusEtapaAberta,
		StatusEtapaEmAndamento,
		StatusEtapaFinalizada,
		StatusEtapaCancelada,
	}
}

// ============================================
// STATUS DE PAGAMENTO
// ============================================

const (
	StatusPagamentoPendente    = "pendente"
	StatusPagamentoPago        = "pago"
	StatusPagamentoCancelado   = "cancelado"
	StatusPagamentoReembolsado = "reembolsado"
)

// GetStatusPagamento retorna todos os status de pagamento válidos
func GetStatusPagamento() []string {
	return []string{
		StatusPagamentoPendente,
		StatusPagamentoPago,
		StatusPagamentoCancelado,
		StatusPagamentoReembolsado,
	}
}

// ============================================
// ESPÉCIES DE PEIXE
// ============================================

const (
	EspecieTucunareAzul    = "tucunare_azul"
	EspecieTucunareAmarelo = "tucunare_amarelo"
	EspecieTraira          = "traira"
)

// GetEspecies retorna todas as espécies válidas
func GetEspecies() []string {
	return []string{
		EspecieTucunareAzul,
		EspecieTucunareAmarelo,
		EspecieTraira,
	}
}

// GetNomeEspecie retorna o nome amigável da espécie
func GetNomeEspecie(especie string) string {
	nomes := map[string]string{
		EspecieTucunareAzul:    "Tucunaré Azul",
		EspecieTucunareAmarelo: "Tucunaré Amarelo",
		EspecieTraira:          "Traíra",
	}
	return nomes[especie]
}

// ============================================
// CATEGORIAS DE RANKING
// ============================================

const (
	CategoriaGeral        = "geral"
	CategoriaMaiorAzul    = "maior_azul"
	CategoriaMaiorAmarelo = "maior_amarelo"
	CategoriaMaiorTraira  = "maior_traira"
)

// GetCategoriasRanking retorna todas as categorias de ranking
func GetCategoriasRanking() []string {
	return []string{
		CategoriaGeral,
		CategoriaMaiorAzul,
		CategoriaMaiorAmarelo,
		CategoriaMaiorTraira,
	}
}

// ============================================
// REGRAS DO TORNEIO
// ============================================

const (
	// Tamanhos mínimos (em cm)
	TamanhoMinimoTucunare = 20.0

	// Cota de peixes
	CotaMaximaPeixes = 4

	// Penalidades (em cm)
	PenalidadeMinima = 0.5
	PenalidadeMaxima = 3.0

	// Horários
	HorarioLimiteRetorno = "16:00"
)

// ============================================
// ESTADOS BRASILEIROS (UF)
// ============================================

var EstadosBrasileiros = []string{
	"AC", "AL", "AP", "AM", "BA", "CE", "DF", "ES", "GO", "MA",
	"MT", "MS", "MG", "PA", "PB", "PR", "PE", "PI", "RJ", "RN",
	"RS", "RO", "RR", "SC", "SP", "SE", "TO",
}

// ValidarUF verifica se a UF é válida
func ValidarUF(uf string) bool {
	for _, estado := range EstadosBrasileiros {
		if estado == uf {
			return true
		}
	}
	return false
}

// ============================================
// MENSAGENS DE ERRO PADRÃO
// ============================================

const (
	ErrEdicaoNaoEncontrada     = "edição não encontrada"
	ErrEtapaNaoEncontrada      = "etapa não encontrada"
	ErrModalidadeNaoEncontrada = "modalidade não encontrada"
	ErrCompetidorNaoEncontrado = "competidor não encontrado"
	ErrInscricaoNaoEncontrada  = "inscrição não encontrada"
	ErrCapturaNaoEncontrada    = "captura não encontrada"
	ErrUsuarioNaoEncontrado    = "usuário não encontrado"
	ErrReguaNaoEncontrada      = "régua não encontrada"

	ErrCompetidorBanido  = "competidor banido do torneio"
	ErrCompetidorInativo = "competidor inativo"
	ErrLicencaVencida    = "licença de pesca vencida"
	ErrEtapaSemVagas     = "etapa sem vagas disponíveis"
	ErrEtapaFechada      = "etapa não está aberta para inscrições"
	ErrPagamentoPendente = "pagamento pendente"
	ErrReguaNaoDevolvida = "régua não foi devolvida"

	ErrEmailJaCadastrado = "email já cadastrado"
	ErrCPFJaCadastrado   = "CPF já cadastrado"
	ErrSenhaIncorreta    = "senha incorreta"
	ErrTokenInvalido     = "token inválido ou expirado"
	ErrPermissaoNegada   = "permissão negada"

	ErrTamanhoMinimo = "peixe abaixo do tamanho mínimo"
	ErrCotaExcedida  = "cota de peixes excedida"
	ErrVideoInvalido = "vídeo inválido ou não encontrado"
	ErrReguaInvalida = "número de régua inválido"

	ErrDadosInvalidos   = "dados inválidos"
	ErrCampoObrigatorio = "campo obrigatório não informado"
)

// ============================================
// MENSAGENS DE SUCESSO PADRÃO
// ============================================

const (
	MsgCriadoComSucesso     = "criado com sucesso"
	MsgAtualizadoComSucesso = "atualizado com sucesso"
	MsgDeletadoComSucesso   = "deletado com sucesso"
	MsgValidadoComSucesso   = "validado com sucesso"
	MsgAnuladoComSucesso    = "anulado com sucesso"
)

// ============================================
// VALIDAÇÕES
// ============================================

// ValidarEspecie valida se a espécie é válida
func ValidarEspecie(especie string) bool {
	especies := GetEspecies()
	for _, e := range especies {
		if e == especie {
			return true
		}
	}
	return false
}

// ValidarStatusEtapa valida se o status da etapa é válido
func ValidarStatusEtapa(status string) bool {
	statusValidos := GetStatusEtapa()
	for _, s := range statusValidos {
		if s == status {
			return true
		}
	}
	return false
}

// ValidarStatusPagamento valida se o status de pagamento é válido
func ValidarStatusPagamento(status string) bool {
	statusValidos := GetStatusPagamento()
	for _, s := range statusValidos {
		if s == status {
			return true
		}
	}
	return false
}

// ValidarTipoUsuario valida se o tipo de usuário é válido
func ValidarTipoUsuario(tipo string) bool {
	tipos := GetTiposUsuario()
	for _, t := range tipos {
		if t == tipo {
			return true
		}
	}
	return false
}

// ValidarPenalidade valida se a penalidade está dentro dos limites
func ValidarPenalidade(penalidade float64) bool {
	return penalidade >= 0 && penalidade <= PenalidadeMaxima
}
