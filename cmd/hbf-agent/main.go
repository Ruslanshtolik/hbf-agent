package main

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/yourusername/hbf-agent/internal/agent"
	"github.com/yourusername/hbf-agent/internal/config"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	var (
		configPath  string
		logLevel    string
		showVersion bool
	)

	flag.StringVar(&configPath, "config", "", "path to configuration file")
	flag.StringVar(&logLevel, "log-level", "", "log level override (debug, info, warn, error)")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.Parse()

	if showVersion {
		fmt.Printf("hbf-agent %s (built %s)\n", Version, BuildTime)
		return
	}

	log := logrus.New()

	if configPath != "" {
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Failed to read config file %s: %v", configPath, err)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if logLevel == "" {
		logLevel = cfg.Log.Level
	}
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		log.Fatalf("Invalid log level %q: %v", logLevel, err)
	}
	log.SetLevel(level)
	if cfg.Log.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{})
	}

	log.Infof("Starting hbf-agent %s (built %s)", Version, BuildTime)
	log.Infof("Node %s in datacenter %s", cfg.Agent.NodeID, cfg.Agent.Datacenter)

	a, err := agent.New(cfg, log)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.Start(ctx)
	}()

	var runErr error
	select {
	case runErr = <-errCh:
	case <-ctx.Done():
	}

	if runErr != nil {
		log.Errorf("Agent stopped with error: %v", runErr)
	}

	if err := a.Stop(); err != nil {
		log.Errorf("Error during shutdown: %v", err)
	}

	log.Info("hbf-agent stopped")
}
