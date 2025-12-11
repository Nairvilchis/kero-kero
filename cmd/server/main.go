package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"

	"kero-kero/internal/config"
	"kero-kero/internal/handlers"
	"kero-kero/internal/repository"
	"kero-kero/internal/routes"
	mw "kero-kero/internal/server/middleware"
	"kero-kero/internal/services"
	"kero-kero/internal/whatsapp"
	"kero-kero/pkg/logger"
)

func main() {
	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error cargando configuración: %v\n", err)
		os.Exit(1)
	}

	// Inicializar logger
	logger.Init(cfg.Logging.Level, cfg.Logging.Format)
	log.Info().Str("app", cfg.App.Name).Str("env", cfg.App.Env).Msg("Iniciando Kero-Kero WhatsApp API")

	// Conectar a base de datos
	db, err := repository.NewDatabase(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Error conectando a base de datos")
	}
	defer db.Close()

	// Conectar a Redis
	redisClient, err := repository.NewRedisClient(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Error conectando a Redis")
	}
	defer redisClient.Close()

	// Repositorio de instancias
	instanceRepo := repository.NewInstanceRepository(db)
	webhookRepo := repository.NewWebhookRepository(redisClient)
	msgRepo := repository.NewMessageRepository(db)

	// Inicializar contenedor de WhatsApp (SQLite por defecto)
	// WhatsApp siempre usa SQLite, independientemente del driver de la BD principal
	waDSN := "file:./data/whatsmeow.db?_foreign_keys=on"
	if cfg.Database.Driver == "sqlite" {
		// Si ya estamos usando SQLite, reutilizamos la misma base de datos
		waDSN = cfg.GetDSN() + "?_foreign_keys=on"
	}
	waContainer, err := sqlstore.New(context.Background(), "sqlite3", waDSN, waLog.Stdout("WA-Store", "INFO", true))
	if err != nil {
		log.Fatal().Err(err).Msg("Error inicializando WhatsApp store")
	}

	// Manager de WhatsApp
	waManager := whatsapp.NewManager(waContainer, instanceRepo, msgRepo, redisClient)
	defer waManager.Close()

	// Cargar instancias existentes
	if err := waManager.LoadInstances(context.Background()); err != nil {
		log.Error().Err(err).Msg("Error cargando instancias")
	}

	// Servicios de negocio
	authService := services.NewAuthService(cfg.Security.JWTSecret, cfg.Security.APIKey)
	webhookService := services.NewWebhookService(webhookRepo)
	instanceService := services.NewInstanceService(waManager, instanceRepo, redisClient, webhookService)
	messageService := services.NewMessageService(waManager, msgRepo, redisClient)
	groupService := services.NewGroupService(waManager)
	contactService := services.NewContactService(waManager)
	presenceService := services.NewPresenceService(waManager) // Nuevo servicio de presencia
	privacyService := services.NewPrivacyService(waManager)
	automationService := services.NewAutomationService(waManager, redisClient.Client) // Nuevo servicio de automatización
	chatService := services.NewChatService(waManager, msgRepo)
	statusService := services.NewStatusService(waManager)
	callService := services.NewCallService(waManager)
	wsService := services.NewWebSocketService()
	crmService := services.NewCRMService()
	syncService := services.NewSyncService(waManager, msgRepo, chatService)

	// Iniciar servicios en segundo plano
	go wsService.Run()
	automationService.StartScheduler()

	// Crear e iniciar el Queue Worker
	queueWorker := services.NewQueueWorker(redisClient, messageService, webhookService)
	queueWorker.Start()

	// Configurar servicios en el manager
	waManager.SetWebhookService(webhookService)
	waManager.SetWebSocketService(wsService)
	waManager.SetAutomationService(automationService)

	// Handlers HTTP
	authHandler := handlers.NewAuthHandler(authService)
	instanceHandler := handlers.NewInstanceHandler(instanceService)
	messageHandler := handlers.NewMessageHandler(messageService)
	groupHandler := handlers.NewGroupHandler(groupService)
	contactHandler := handlers.NewContactHandler(contactService)
	presenceHandler := handlers.NewPresenceHandler(presenceService) // Nuevo handler de presencia
	privacyHandler := handlers.NewPrivacyHandler(privacyService)
	automationHandler := handlers.NewAutomationHandler(automationService) // Nuevo handler de automatización
	chatHandler := handlers.NewChatHandler(chatService)
	statusHandler := handlers.NewStatusHandler(statusService)
	callHandler := handlers.NewCallHandler(callService)
	webhookHandler := handlers.NewWebhookHandler(webhookService)
	crmHandler := handlers.NewCRMHandler(crmService)
	syncHandler := handlers.NewSyncHandler(syncService)

	// Router
	r := chi.NewRouter()
	// Middlewares globales
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mw.Logger())
	r.Use(middleware.Recoverer)
	r.Use(mw.CORS(cfg.CORS.AllowedOrigins, cfg.CORS.AllowedMethods, cfg.CORS.AllowedHeaders))
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(mw.NewRateLimiter(cfg.Security.RateLimitReqs, cfg.Security.RateLimitWindow).Middleware())

	// Rutas públicas
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"service\":\"Kero-Kero WhatsApp API\",\"version\":\"2.0.0\",\"status\":\"running\"}")
	})
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		health := map[string]interface{}{
			"status":   "healthy",
			"database": "ok",
			"redis":    "ok",
		}

		ctx := r.Context()

		// Verificar base de datos
		if err := db.Health(ctx); err != nil {
			health["database"] = "error"
			health["database_error"] = err.Error()
			health["status"] = "unhealthy"
		}

		// Verificar Redis
		if err := redisClient.Health(ctx); err != nil {
			health["redis"] = "error"
			health["redis_error"] = err.Error()
			health["status"] = "unhealthy"
		}

		statusCode := http.StatusOK
		if health["status"] == "unhealthy" {
			statusCode = http.StatusServiceUnavailable
		}

		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(health)
	})

	// Endpoint de información del sistema (público)
	r.Get("/system/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		systemInfo := map[string]interface{}{
			"app_name":   cfg.App.Name,
			"app_env":    cfg.App.Env,
			"version":    "2.0.0",
			"database":   cfg.Database.Driver,
			"server_url": fmt.Sprintf("http://localhost:%s", cfg.App.Port),
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(systemInfo)
	})

	// WebSocket endpoint por instancia
	r.Get("/instances/{instanceID}/ws", wsService.HandleConnection)

	// Rutas de autenticación (públicas, sin API Key)
	routes.RegisterAuthRoutes(r, authHandler)

	// Rutas protegidas (API Key o JWT)
	r.Group(func(r chi.Router) {
		if cfg.Security.APIKey != "" {
			r.Use(mw.Auth(cfg.Security.APIKey, authService))
		}
		routes.SetupInstanceRoutes(r, instanceHandler)
		routes.SetupMessageRoutes(r, messageHandler)
		routes.SetupGroupRoutes(r, groupHandler)
		routes.SetupContactRoutes(r, contactHandler)
		routes.RegisterPresenceRoutes(r, presenceHandler) // Nuevo: rutas de presencia
		routes.SetupPrivacyRoutes(r, privacyHandler)
		routes.SetupAutomationRoutes(r, automationHandler) // Nuevo: rutas de automatización
		routes.SetupChatRoutes(r, chatHandler)
		routes.SetupStatusRoutes(r, statusHandler)
		routes.SetupCallRoutes(r, callHandler)
		routes.SetupWebhookRoutes(r, webhookHandler)
		routes.SetupCRMRoutes(r, crmHandler)
		routes.SetupSyncRoutes(r, syncHandler)
	})

	// Servidor HTTP
	srv := &http.Server{Addr: ":" + cfg.App.Port, Handler: r}
	go func() {
		log.Info().Str("port", cfg.App.Port).Msg("Servidor HTTP iniciado")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Error del servidor HTTP")
		}
	}()

	// Esperar señal de terminación
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Apagando servidor...")
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Error().Err(err).Msg("Error durante el shutdown del servidor")
	}

	log.Info().Msg("Servidor detenido correctamente")
}
