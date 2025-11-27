package database

import (
	"fmt"
	"time"

	"github.com/douglasmaicon/Copa-Trick-Fish/internal/config"
	"github.com/douglasmaicon/Copa-Trick-Fish/internal/models"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect estabelece conexão com o banco de dados PostgreSQL
func Connect(cfg *config.Config) error {
	var err error

	// Configurar logger do GORM
	gormLogger := logger.Default
	if cfg.IsDevelopment() {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// String de conexão
	dsn := cfg.Database.GetDSN()

	// Tentar conectar com retry
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogger,
			NowFunc: func() time.Time {
				// Usar timezone configurado
				loc, _ := time.LoadLocation(cfg.Database.TimeZone)
				return time.Now().In(loc)
			},
		})

		if err == nil {
			break
		}

		logrus.Warnf("⚠️  Tentativa %d/%d de conexão com banco falhou: %v", i+1, maxRetries, err)
		time.Sleep(time.Second * 2)
	}

	if err != nil {
		return fmt.Errorf("falha ao conectar ao banco após %d tentativas: %w", maxRetries, err)
	}

	// Configurar pool de conexões
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("erro ao obter conexão SQL: %w", err)
	}

	// Configurações de pool (importante para performance)
	sqlDB.SetMaxIdleConns(10)           // Conexões idle no pool
	sqlDB.SetMaxOpenConns(100)          // Máximo de conexões abertas
	sqlDB.SetConnMaxLifetime(time.Hour) // Tempo máximo de vida da conexão

	// Testar conexão
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("erro ao fazer ping no banco: %w", err)
	}

	logrus.Info("✅ Conectado ao banco de dados PostgreSQL")
	return nil
}

// Migrate executa as migrations automáticas
func Migrate() error {
	logrus.Info("🔄 Executando migrations...")

	// Habilitar extensão para UUID no PostgreSQL
	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		logrus.Warnf("⚠️  Não foi possível criar extensão uuid-ossp: %v", err)
	}

	// AutoMigrate cria/atualiza as tabelas automaticamente
	err := DB.AutoMigrate(
		&models.Usuario{},
		&models.Edicao{},
		&models.Modalidade{},
		&models.Etapa{},
		&models.Competidor{},
		&models.Regua{},
		&models.Inscricao{},
		&models.Captura{},
		&models.Ranking{},
	)

	if err != nil {
		return fmt.Errorf("erro ao executar migrations: %w", err)
	}

	logrus.Info("✅ Migrations executadas com sucesso")
	return nil
}

// Seed insere dados iniciais no banco
func Seed() error {
	logrus.Info("🌱 Inserindo dados iniciais (seed)...")

	// Verificar se já existem dados
	var countModalidades int64
	DB.Model(&models.Modalidade{}).Count(&countModalidades)

	if countModalidades > 0 {
		logrus.Info("ℹ️  Banco já possui dados, pulando seed")
		return nil
	}

	// Criar modalidades padrão
	modalidades := []models.Modalidade{
		{Nome: "Embarcada", Descricao: "Competição em barcos e lanchas", Ordem: 1},
		{Nome: "Caiaque", Descricao: "Competição em caiaques com remo e/ou pedal", Ordem: 2},
		{Nome: "Casais", Descricao: "Competição em duplas (casais)", Ordem: 3},
		{Nome: "Feminino", Descricao: "Competição exclusiva feminina", Ordem: 4},
		{Nome: "Infantil", Descricao: "Competição infantil", Ordem: 5},
	}

	for _, modalidade := range modalidades {
		if err := DB.Create(&modalidade).Error; err != nil {
			return fmt.Errorf("erro ao criar modalidade %s: %w", modalidade.Nome, err)
		}
		logrus.Infof("  ✓ Modalidade criada: %s", modalidade.Nome)
	}

	// Criar usuário admin padrão
	senhaHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("erro ao gerar hash da senha: %w", err)
	}

	admin := models.Usuario{
		Nome:  "Administrador",
		Email: "admin@trickfish.com.br",
		Senha: string(senhaHash),
		Tipo:  models.TipoUsuarioAdmin,
		Ativo: true,
	}

	if err := DB.Create(&admin).Error; err != nil {
		return fmt.Errorf("erro ao criar usuário admin: %w", err)
	}
	logrus.Info("  ✓ Usuário admin criado (email: admin@trickfish.com.br, senha: admin123)")

	// Criar edição exemplo
	edicaoAtual := models.Edicao{
		Ano:       time.Now().Year(),
		Nome:      fmt.Sprintf("Copa Trick-Fish %d", time.Now().Year()),
		Descricao: "Edição atual do torneio de pesca esportiva",
		Ativa:     true,
	}

	if err := DB.Create(&edicaoAtual).Error; err != nil {
		return fmt.Errorf("erro ao criar edição: %w", err)
	}
	logrus.Infof("  ✓ Edição criada: %s", edicaoAtual.Nome)

	// Associar todas as modalidades à edição
	if err := DB.Model(&edicaoAtual).Association("Modalidades").Append(&modalidades); err != nil {
		return fmt.Errorf("erro ao associar modalidades: %w", err)
	}

	logrus.Info("✅ Seed concluído com sucesso!")
	logrus.Info("============================================")
	logrus.Info("📧 Credenciais de acesso:")
	logrus.Info("   Email: admin@trickfish.com.br")
	logrus.Info("   Senha: admin123")
	logrus.Info("   ⚠️  ALTERE A SENHA EM PRODUÇÃO!")
	logrus.Info("============================================")

	return nil
}

// Close fecha a conexão com o banco
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("erro ao obter conexão SQL: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("erro ao fechar conexão: %w", err)
	}

	logrus.Info("✅ Conexão com banco de dados fechada")
	return nil
}

// HealthCheck verifica se o banco está saudável
func HealthCheck() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("erro ao obter conexão SQL: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("banco não responde: %w", err)
	}

	return nil
}

// GetStats retorna estatísticas da conexão
func GetStats() map[string]interface{} {
	sqlDB, _ := DB.DB()
	stats := sqlDB.Stats()

	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
	}
}
