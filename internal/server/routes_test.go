package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"x07-it/radiod/internal/config"
	"x07-it/radiod/internal/stream"
)

// createTestPlayer создаёт Player с одной станцией и временным треком.
func createTestPlayer(t *testing.T) (*stream.Player, string, string) {
	t.Helper()
	dir := t.TempDir()
	track := filepath.Join(dir, "track.mp3")
	content := []byte("test-audio")
	if err := os.WriteFile(track, content, 0o644); err != nil {
		t.Fatalf("write track: %v", err)
	}
	logrus.WithField("track", track).Info("создан временный трек")
	cfg := config.Config{OutputBitrate: "320k", BufferSeconds: 0}
	p := stream.NewPlayer(cfg)
	p.AddStation("demo", []string{track})
	return p, track, string(content)
}

// TestStationsEndpoint проверяет выдачу списка станций.
func TestStationsEndpoint(t *testing.T) {
	logrus.Info("тестируем /stations")
	gin.SetMode(gin.TestMode)
	p, _, _ := createTestPlayer(t)
	r := SetupRouter(p)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stations", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	}
	expected := "[\"demo\"]"
	if strings.TrimSpace(w.Body.String()) != expected {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

// TestNowPlayingEndpoint проверяет текущий трек станции.
func TestNowPlayingEndpoint(t *testing.T) {
	logrus.Info("тестируем /nowplaying")
	gin.SetMode(gin.TestMode)
	p, track, _ := createTestPlayer(t)
	r := SetupRouter(p)
	// Ожидаем установки NowPlaying.
	st := p.Get("demo")
	deadline := time.Now().Add(time.Second)
	for st.NowPlaying() == "" {
		if time.Now().After(deadline) {
			t.Fatal("now playing not set")
		}
		time.Sleep(10 * time.Millisecond)
	}

	w := httptest.NewRecorder()
	url := "/nowplaying/demo"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	expected := fmt.Sprintf("{\"now\":\"%s\"}", filepath.Base(track))
	if strings.TrimSpace(w.Body.String()) != expected {
		t.Fatalf("body %s", w.Body.String())
	}
}

// TestStreamEndpoint проверяет получение аудиоданных.
func TestStreamEndpoint(t *testing.T) {
	logrus.Info("тестируем /stream")
	gin.SetMode(gin.TestMode)
	p, _, content := createTestPlayer(t)
	r := SetupRouter(p)
	srv := httptest.NewServer(r)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/stream/demo", nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()

	buf := make([]byte, len(content))
	if _, err := io.ReadFull(resp.Body, buf); err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(buf, []byte(content)) {
		t.Fatalf("unexpected data: %q", buf)
	}
}
