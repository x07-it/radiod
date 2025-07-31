package main

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

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

	// Setup HTTP routes with Gin and create HTTP server.
	r := server.SetupRouter(player)
	srv := &http.Server{Addr: cfg.Listen, Handler: r}

	go func() {
		for i := 5; i > 0; i-- {
			//fmt. Printf("Осталось %d секунд...\n", i)
			time.Sleep(1 * time.Second)
		}
		openBrowser("http://localhost:7000")
	}()

	// Run server in separate goroutine.
	go func() {
		logrus.Infof("Starting server on %s", cfg.Listen)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("listen and serve")
		}
	}()

	// Wait for termination signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutdown signal received")

	// Gracefully shutdown HTTP server and stop stations.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logrus.WithError(err).Error("server shutdown")
	}
	player.Stop()

	logrus.Info("Server exited")
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
