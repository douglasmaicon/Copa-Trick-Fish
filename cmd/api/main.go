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

		// TODO: Adicionar rotas de recursos aqui
		// Exemplo:
		// edicoes := api.Group("/edicoes")
		// {
		//     edicoes.GET("", handlers.ListarEdicoes)
		//     edicoes.GET("/:id", handlers.BuscarEdicao)
		// }
		modalidades := api.Group("/modalidades")
		{
			modalidades.GET("", handlers.ListarModalidades)
		}

		// Edições
		edicoes := api.Group("/edicoes")
		{
			edicoes.GET("", handlers.ListarEdicoes)
			edicoes.GET("/ativa", handlers.BuscarEdicaoAtiva)
			edicoes.GET("/:id", handlers.BuscarEdicao)
			edicoes.POST("", handlers.CriarEdicao)
			edicoes.PUT("/:id", handlers.AtualizarEdicao)
			edicoes.DELETE("/:id", handlers.DeletarEdicao)
		}

		// Etapas
		etapas := api.Group("/etapas")
		{
			etapas.GET("", handlers.ListarEtapas)
			etapas.GET("/:id", handlers.BuscarEtapa)
			etapas.POST("", handlers.CriarEtapa)
			etapas.PUT("/:id", handlers.AtualizarEtapa)
			etapas.DELETE("/:id", handlers.DeletarEtapa)
		}

		// Competidores
		competidores := api.Group("/competidores")
		{
			competidores.GET("", handlers.ListarCompetidores)
			competidores.GET("/:id", handlers.BuscarCompetidor)
			competidores.POST("", handlers.CriarCompetidor)
			competidores.PUT("/:id", handlers.AtualizarCompetidor)
			competidores.POST("/:id/banir", handlers.BanirCompetidor)
			competidores.POST("/:id/desbanir", handlers.DesbanirCompetidor)
		}

		// Inscrições
		inscricoes := api.Group("/inscricoes")
		{
			inscricoes.GET("", handlers.ListarInscricoes)
			inscricoes.GET("/:id", handlers.BuscarInscricao)
			inscricoes.POST("", handlers.CriarInscricao)
			inscricoes.POST("/:id/confirmar-pagamento", handlers.ConfirmarPagamento)
			inscricoes.POST("/:id/eliminar", handlers.EliminarCompetidor)
			inscricoes.POST("/:id/devolver-regua", handlers.DevolverRegua)
		}

		// Réguas
		reguas := api.Group("/reguas")
		{
			reguas.GET("", handlers.ListarReguas)
			reguas.POST("/gerar", handlers.GerarReguas)
			reguas.POST("/sortear", handlers.SortearReguas)
			reguas.DELETE("/:id", handlers.DeletarRegua)
		}

		// Capturas
		capturas := api.Group("/capturas")
		{
			capturas.GET("", handlers.ListarCapturas)
			capturas.GET("/:id", handlers.BuscarCaptura)
			capturas.POST("", handlers.CriarCaptura)
			capturas.PUT("/:id/validar", handlers.ValidarCaptura)
			capturas.PUT("/:id/anular", handlers.AnularCaptura)
			capturas.DELETE("/:id", handlers.DeletarCaptura)
		}

		// Rankings
		rankings := api.Group("/rankings")
		{
			rankings.GET("", handlers.ListarRankings)
			rankings.GET("/etapa/:id", handlers.BuscarRankingEtapa)
			rankings.POST("/etapa/:id/gerar", handlers.GerarRanking)
			rankings.DELETE("/:id", handlers.DeletarRanking)
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
