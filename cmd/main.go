package main

import (
	"context"
	clickrepo "github.com/ilam072/shortener/internal/click/repo/postgres"
	clickrest "github.com/ilam072/shortener/internal/click/rest"
	clickservice "github.com/ilam072/shortener/internal/click/service"
	"github.com/ilam072/shortener/internal/config"
	"github.com/ilam072/shortener/internal/link/cache"
	linkrepo "github.com/ilam072/shortener/internal/link/repo/postgres"
	linkrest "github.com/ilam072/shortener/internal/link/rest"
	linkservice "github.com/ilam072/shortener/internal/link/service"
	"github.com/ilam072/shortener/internal/middleware"
	"github.com/ilam072/shortener/internal/validator"
	"github.com/ilam072/shortener/pkg/db"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/zlog"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Initialize logger
	zlog.Init()

	// Context
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// Initialize config
	cfg := config.MustLoad()

	// Connect to DB
	DB, err := db.OpenDB(cfg.DB)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to connect to DB")
	}

	// Connect to Redis
	redisClient := redis.New(cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB)

	if err = redisClient.Ping(ctx).Err(); err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to ping redis")
	}

	// Initialize validator
	v := validator.New()

	// Initialize cache
	linkCache := cache.New(redisClient)

	// Initialize link and click repositories
	clickRepo := clickrepo.New(DB)
	linkRepo := linkrepo.New(DB)

	// Initialize link and click services
	link := linkservice.New(linkRepo, linkCache)
	click := clickservice.New(clickRepo)

	// Initialize link and click handlers
	linkHandler := linkrest.NewLinkHandler(link, click, v)
	clickHandler := clickrest.NewClickHandler(click)

	// Initialize Gin engine
	engine := ginext.New("")
	engine.Use(ginext.Logger())
	engine.Use(ginext.Recovery())
	engine.Use(middleware.TimeoutMiddleware(2 * time.Second))

	apiGroup := engine.Group("/api")
	apiGroup.POST("/shorten", linkHandler.CreateLink)
	apiGroup.GET("/s/:alias", linkHandler.Redirect)
	apiGroup.GET("/analytics/:alias", clickHandler.GetAnalytics)

	// Initialize and start http server
	server := &http.Server{
		Addr:    cfg.Server.HTTPPort,
		Handler: engine,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			zlog.Logger.Fatal().Err(err).Msg("failed to listen start http server")
		}
	}()

	<-ctx.Done()

	// Graceful shutdown
	withTimeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := server.Shutdown(withTimeout); err != nil {
		zlog.Logger.Error().Err(err).Msg("server shutdown failed")
	}

	if err := DB.Master.Close(); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to close master database")
	}
}
