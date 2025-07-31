package stream

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"x07-it/radiod/internal/config"
)

// TestAddStation проверяет регистрацию станции.
func TestAddStation(t *testing.T) {
	logrus.Info("проверка AddStation")
	dir := t.TempDir()
	track := filepath.Join(dir, "a.mp3")
	if err := os.WriteFile(track, []byte("aaa"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	cfg := config.Config{OutputBitrate: "320k", BufferSeconds: 0}
	p := NewPlayer(cfg)
	p.AddStation("a", []string{track})
	if p.Get("a") == nil {
		t.Fatal("station not added")
	}
	names := p.StationNames()
	if len(names) != 1 || names[0] != "a" {
		t.Fatalf("unexpected names: %v", names)
	}
}

// TestAddListener проверяет добавление и удаление слушателя.
func TestAddListener(t *testing.T) {
	logrus.Info("проверка AddListener")
	st := &Station{name: "x", listeners: make(map[chan []byte]struct{})}
	ch, remove := st.AddListener()
	if len(st.listeners) != 1 {
		t.Fatal("listener not stored")
	}
	if cap(ch) != 16 {
		t.Fatalf("unexpected cap: %d", cap(ch))
	}
	remove()
	if _, ok := <-ch; ok {
		t.Fatal("channel not closed")
	}
}

// TestBroadcastQueue проверяет порядок доставки данных через очередь.
func TestBroadcastQueue(t *testing.T) {
	logrus.Info("проверка очереди вещания")
	st := &Station{name: "q", listeners: make(map[chan []byte]struct{})}
	ch, _ := st.AddListener()
	st.broadcast([]byte("one"))
	st.broadcast([]byte("two"))
	d1 := <-ch
	d2 := <-ch
	if !bytes.Equal(d1, []byte("one")) || !bytes.Equal(d2, []byte("two")) {
		t.Fatalf("unexpected data: %q %q", d1, d2)
	}
}
