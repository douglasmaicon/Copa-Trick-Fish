package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Config armazena todas as configurações da aplicação
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Storage  StorageConfig
}

// ServerConfig - configurações do servidor HTTP
type ServerConfig struct {
	Port        string
	Environment string // development, staging, production
	LogLevel    string
}

// DatabaseConfig - configurações do banco de dados
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	TimeZone string
}

// JWTConfig - configurações de autenticação JWT
type JWTConfig struct {
	Secret     string
	Expiration int // em horas
}

// StorageConfig - configurações de armazenamento de arquivos
type StorageConfig struct {
	Type      string // local, s3, cloudinary
	LocalPath string
	BucketURL string
}

var AppConfig *Config

// Load carrega as configurações das variáveis de ambiente
func Load() (*Config, error) {
	// Tenta carregar o arquivo .env (não dá erro se não existir)
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port:        getEnv("PORT", "8080"),
			Environment: getEnv("ENVIRONMENT", "development"),
			LogLevel:    getEnv("LOG_LEVEL", "info"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "trickfish"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			TimeZone: getEnv("DB_TIMEZONE", "America/Sao_Paulo"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "change-me-in-production"),
			Expiration: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
		},
		Storage: StorageConfig{
			Type:      getEnv("STORAGE_TYPE", "local"),
			LocalPath: getEnv("STORAGE_LOCAL_PATH", "./uploads"),
			BucketURL: getEnv("STORAGE_BUCKET_URL", ""),
		},
	}

	// Validações críticas
	if config.Server.Environment == "production" {
		if config.JWT.Secret == "change-me-in-production" {
			return nil, fmt.Errorf("JWT_SECRET deve ser configurado em produção")
		}
		if config.Database.SSLMode == "disable" {
			logrus.Warn("⚠️  SSL está desabilitado no banco de dados em produção!")
		}
	}

	// Configurar nível de log
	setLogLevel(config.Server.LogLevel)

	AppConfig = config
	logrus.Info("✅ Configurações carregadas com sucesso")

	return config, nil
}

// GetDSN retorna a string de conexão do PostgreSQL
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		c.Host,
		c.User,
		c.Password,
		c.DBName,
		c.Port,
		c.SSLMode,
		c.TimeZone,
	)
}

// IsDevelopment verifica se está em ambiente de desenvolvimento
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction verifica se está em ambiente de produção
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// getEnv obtém variável de ambiente ou retorna valor padrão
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt obtém variável de ambiente como inteiro
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		logrus.Warnf("Não foi possível converter %s para int, usando valor padrão: %d", key, defaultValue)
		return defaultValue
	}
	return value
}

// setLogLevel configura o nível de log
func setLogLevel(level string) {
	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Formato do log
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}

// PrintConfig imprime as configurações (ocultando dados sensíveis)
func (c *Config) PrintConfig() {
	logrus.Info("==================== CONFIGURAÇÕES ====================")
	logrus.Infof("Ambiente: %s", c.Server.Environment)
	logrus.Infof("Porta: %s", c.Server.Port)
	logrus.Infof("Log Level: %s", c.Server.LogLevel)
	logrus.Infof("Database Host: %s:%s", c.Database.Host, c.Database.Port)
	logrus.Infof("Database Name: %s", c.Database.DBName)
	logrus.Infof("Storage Type: %s", c.Storage.Type)
	logrus.Info("======================================================")
}
