package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	configutils "github.com/sazonovItas/proxy-manager/pkg/config/utils"
	"github.com/sazonovItas/proxy-manager/pkg/logger"
	slogger "github.com/sazonovItas/proxy-manager/pkg/logger/sl"

	"github.com/sazonovItas/proxy-manager/services/proxy-manager/internal/config"
	"github.com/sazonovItas/proxy-manager/services/proxy-manager/internal/engine"
)

const (
	configPathEnv = "CONFIG_PATH"
)

func main() {
	cfg, err := configutils.LoadConfigFromFile[config.Config](os.Getenv(configPathEnv))
	if err != nil {
		log.Fatalf("faild to load proxy manager config: %s", err.Error())
	}

	logger := logger.NewSlogLogger(
		logger.LogConfig{Environment: "dev", LogLevel: logger.DEBUG},
		os.Stdout,
	)
	logger.Info("init config", "config", *cfg)

	engine, err := engine.NewEngine(cfg.ProxyManager.ProxyImage.Image, engine.DockerClientConfig{
		ApiVersion: cfg.DockerClient.ApiVersion,
		Timeout:    cfg.DockerClient.Timeout,
		Host:       cfg.DockerClient.Host,
	})
	if err != nil {
		logger.Error("failed init engine", slogger.Err(err))
		return
	}

	err = engine.Run(cfg.ProxyManager.Proxies)
	if err != nil {
		logger.Error("failed run proxy manager", slogger.Err(err))
		return
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	<-ctx.Done()

	// TODO: change shutdown timeout with config value
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer func() {
		cancel()

		if shutdownCtx.Err() != nil && !errors.Is(shutdownCtx.Err(), context.Canceled) {
			logger.Warn("proxy manager shutdown with error", slogger.Err(shutdownCtx.Err()))
		}
	}()

	if err := engine.Shutdown(shutdownCtx); err != nil {
		logger.Error("engine shuted down with error", "error", err.Error())
	}

	logger.Info("proxy manager is shuted down")
}
