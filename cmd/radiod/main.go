package main

import (
	"os/exec" // проверка внешних зависимостей

	"github.com/sirupsen/logrus"
	"x07-it/radiod/internal/config"
	"x07-it/radiod/internal/convert"
	"x07-it/radiod/internal/server"
	"x07-it/radiod/internal/stream"
)

// main is the entry point of the radio server.
// It loads configuration, prepares audio cache and starts the HTTP server.
func main() {
	// Load configuration using Viper.
	cfg, err := config.Load()
	if err != nil {
		logrus.WithError(err).Fatal("cannot load config")
	}

	// Убедиться, что ffmpeg доступен в системе.
	if _, err := exec.LookPath(cfg.FFMpegPath); err != nil {
		logrus.WithError(err).Fatal("ffmpeg not found")
	}

	// Pre-convert audio files and gather track list per station.
	stationTracks, err := convert.PrepareCache(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("prepare cache failed")
	}

	// Create player and start stations.
	player := stream.NewPlayer(cfg)
	for station, tracks := range stationTracks {
		player.AddStation(station, tracks)
	}

	// Setup HTTP routes with Gin.
	r := server.SetupRouter(player)

	logrus.Infof("Starting server on %s", cfg.Listen)
	if err := r.Run(cfg.Listen); err != nil {
		logrus.WithError(err).Fatal("server stopped")
	}
}
