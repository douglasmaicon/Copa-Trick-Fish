package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/douglasmaicon/Copa-Trick-Fish/internal/config"
	"github.com/douglasmaicon/Copa-Trick-Fish/internal/database"
	"github.com/douglasmaicon/Copa-Trick-Fish/internal/handlers"
	"github.com/douglasmaicon/Copa-Trick-Fish/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Banner da aplicação
	printBanner()

	// Carregar configurações
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatalf("❌ Erro ao carregar configurações: %v", err)
	}

	// Mostrar configurações
	cfg.PrintConfig()

	// Conectar ao banco de dados
	if err := database.Connect(cfg); err != nil {
		logrus.Fatalf("❌ Erro ao conectar ao banco: %v", err)
	}
	defer database.Close()

	// Executar migrations
	if err := database.Migrate(); err != nil {
		logrus.Fatalf("❌ Erro ao executar migrations: %v", err)
	}

	// Inserir dados iniciais (seed)
	if err := database.Seed(); err != nil {
		logrus.Fatalf("❌ Erro ao executar seed: %v", err)
	}

	// Configurar modo do Gin
	if !cfg.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Criar router
	router := gin.New()

	// Middlewares globais
	setupMiddlewares(router, cfg)

	// Configurar rotas
	setupRoutes(router)

	// Criar servidor HTTP
	srv := &http.Server{
		Addr:           ":" + cfg.Server.Port,
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Iniciar servidor em goroutine
	go func() {
		logrus.Infof("🚀 Servidor iniciado em http://localhost:%s", cfg.Server.Port)
		logrus.Info("📚 Documentação da API: http://localhost:" + cfg.Server.Port + "/api/docs")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("❌ Erro ao iniciar servidor: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("🛑 Desligando servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("❌ Erro ao desligar servidor: %v", err)
	}

	logrus.Info("✅ Servidor desligado com sucesso")
}

// setupMiddlewares configura os middlewares globais
func setupMiddlewares(router *gin.Engine, cfg *config.Config) {
	// Logger customizado
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
		)
	}))

	// Recovery (recupera de panics)
	router.Use(gin.Recovery())

	// CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Middleware de segurança básica
	router.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Next()
	})
}

// setupRoutes configura todas as rotas da API
func setupRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		if err := database.HealthCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now(),
			"db_stats":  database.GetStats(),
		})
	})

	// Rota raiz
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "🎣 Copa Trick-Fish API",
			"version": "1.0.0",
			"docs":    "/api/docs",
		})
	})

	// Grupo de rotas da API
	api := router.Group("/api")
	{
		// Versão da API
		api.GET("/version", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"version": "1.0.0",
				"build":   "development",
			})
		})

		// ============================================
		// ROTAS PÚBLICAS (SEM AUTENTICAÇÃO)
		// ============================================

		// Autenticação
		auth := api.Group("/auth")
		{
			auth.POST("/login/usuario", handlers.LoginUsuario)
			auth.POST("/login/competidor", handlers.LoginCompetidor)
			auth.POST("/refresh", handlers.RefreshToken)
		}

		// Modalidades (público)
		api.GET("/modalidades", handlers.ListarModalidades)

		// Edições (público - apenas leitura)
		api.GET("/edicoes", handlers.ListarEdicoes)
		api.GET("/edicoes/ativa", handlers.BuscarEdicaoAtiva)
		api.GET("/edicoes/:id", handlers.BuscarEdicao)

		// Etapas (público - apenas leitura)
		api.GET("/etapas", handlers.ListarEtapas)
		api.GET("/etapas/:id", handlers.BuscarEtapa)

		// Rankings (público)
		api.GET("/rankings", handlers.ListarRankings)
		api.GET("/rankings/etapa/:id", handlers.BuscarRankingEtapa)

		// ============================================
		// ROTAS AUTENTICADAS (REQUER LOGIN)
		// ============================================

		cfg := config.AppConfig
		autenticado := api.Group("")
		autenticado.Use(middleware.AuthMiddleware(cfg))
		{
			// Perfil do usuário logado
			autenticado.GET("/perfil", handlers.MeuPerfil)

			// Inscrições (competidores podem criar suas próprias)
			autenticado.POST("/inscricoes", handlers.CriarInscricao)
			autenticado.GET("/inscricoes/:id", handlers.BuscarInscricao)

			// Capturas (competidores podem registrar)
			autenticado.POST("/capturas", handlers.CriarCaptura)
			autenticado.GET("/capturas/:id", handlers.BuscarCaptura)
		}

		// ============================================
		// ROTAS DE FISCAL (fiscal, organizador, admin)
		// ============================================

		fiscal := api.Group("/fiscal")
		fiscal.Use(middleware.AuthMiddleware(cfg))
		fiscal.Use(middleware.FiscalOnly())
		{
			// Listar capturas para validação
			fiscal.GET("/capturas", handlers.ListarCapturas)

			// Validar e anular capturas
			fiscal.PUT("/capturas/:id/validar", handlers.ValidarCaptura)
			fiscal.PUT("/capturas/:id/anular", handlers.AnularCaptura)

			// Listar inscrições
			fiscal.GET("/inscricoes", handlers.ListarInscricoes)

			// Devolução de régua
			fiscal.POST("/inscricoes/:id/devolver-regua", handlers.DevolverRegua)
		}

		// ============================================
		// ROTAS DE ORGANIZADOR (organizador, admin)
		// ============================================

		organizador := api.Group("/organizador")
		organizador.Use(middleware.AuthMiddleware(cfg))
		organizador.Use(middleware.OrganizadorOnly())
		{
			// Gerenciar etapas
			organizador.POST("/etapas", handlers.CriarEtapa)
			organizador.PUT("/etapas/:id", handlers.AtualizarEtapa)
			organizador.DELETE("/etapas/:id", handlers.DeletarEtapa)

			// Gerenciar réguas
			organizador.GET("/reguas", handlers.ListarReguas)
			organizador.POST("/reguas/gerar", handlers.GerarReguas)
			organizador.POST("/reguas/sortear", handlers.SortearReguas)
			organizador.DELETE("/reguas/:id", handlers.DeletarRegua)

			// Gerenciar inscrições
			organizador.POST("/inscricoes/:id/confirmar-pagamento", handlers.ConfirmarPagamento)
			organizador.POST("/inscricoes/:id/eliminar", handlers.EliminarCompetidor)

			// Gerenciar rankings
			organizador.POST("/rankings/etapa/:id/gerar", handlers.GerarRanking)
			organizador.DELETE("/rankings/:id", handlers.DeletarRanking)

			// Gerenciar competidores
			organizador.GET("/competidores", handlers.ListarCompetidores)
			organizador.GET("/competidores/:id", handlers.BuscarCompetidor)
			organizador.PUT("/competidores/:id", handlers.AtualizarCompetidor)
			organizador.POST("/competidores/:id/banir", handlers.BanirCompetidor)
			organizador.POST("/competidores/:id/desbanir", handlers.DesbanirCompetidor)
		}

		// ============================================
		// ROTAS DE ADMIN (apenas admin)
		// ============================================

		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(cfg))
		admin.Use(middleware.AdminOnly())
		{
			// Gerenciar edições
			admin.POST("/edicoes", handlers.CriarEdicao)
			admin.PUT("/edicoes/:id", handlers.AtualizarEdicao)
			admin.DELETE("/edicoes/:id", handlers.DeletarEdicao)

			// Registro público de competidores
			admin.POST("/competidores", handlers.CriarCompetidor)

			// Deletar capturas
			admin.DELETE("/capturas/:id", handlers.DeletarCaptura)
		}
	}

	// Rota 404
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Rota não encontrada",
			"path":  c.Request.URL.Path,
		})
	})
}

// printBanner imprime o banner da aplicação
func printBanner() {
	banner := `
╔════════════════════════════════════════════════════════════╗
║                                                            ║
║        🎣 COPA TRICK-FISH - SISTEMA DE GESTÃO 🎣         ║
║                                                            ║
║              Torneio de Pesca Esportiva                   ║
║                     v1.0.0                                ║
║                                                            ║
╚════════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}
