package stream

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"x07-it/radiod/internal/config"
)

// Station represents a single radio station with listeners and track list.
type Station struct {
	name      string
	tracks    []string
	current   string
	listeners map[chan []byte]struct{}
	mu        sync.Mutex
	buffer    time.Duration
	bps       int // bytes per second for throttling
	stop      chan struct{}
	stopOnce  sync.Once
}

// Player manages multiple stations.
type Player struct {
	cfg      config.Config
	stations map[string]*Station
}

// NewPlayer creates a player with given configuration.
func NewPlayer(cfg config.Config) *Player {
	return &Player{cfg: cfg, stations: make(map[string]*Station)}
}

// AddStation registers new station and starts playback goroutine.
func (p *Player) AddStation(name string, tracks []string) {
	st := &Station{
		name:      name,
		tracks:    tracks,
		listeners: make(map[chan []byte]struct{}),
		buffer:    time.Duration(p.cfg.BufferSeconds) * time.Second,
		bps:       parseBitrate(p.cfg.OutputBitrate),
		stop:      make(chan struct{}),
	}
	p.stations[name] = st
	go st.loop()
}

// StationNames returns list of all station identifiers.
func (p *Player) StationNames() []string {
	res := make([]string, 0, len(p.stations))
	for k := range p.stations {
		res = append(res, k)
	}
	return res
}

// Get returns station by name.
func (p *Player) Get(name string) *Station { return p.stations[name] }

// Stop stops playback for all stations.
func (p *Player) Stop() {
	for _, st := range p.stations {
		st.Stop()
	}
}

// NowPlaying returns current track name for station.
func (s *Station) NowPlaying() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.current
}

// AddListener attaches a new listener channel to station.
func (s *Station) AddListener() (chan []byte, func()) {
	ch := make(chan []byte, 16)
	s.mu.Lock()
	s.listeners[ch] = struct{}{}
	s.mu.Unlock()

	remove := func() {
		s.mu.Lock()
		delete(s.listeners, ch)
		close(ch)
		s.mu.Unlock()
	}
	return ch, remove
}

// Stop signals station loop to exit.
func (s *Station) Stop() {
	s.stopOnce.Do(func() { close(s.stop) })
}

// loop continuously plays tracks and broadcasts to listeners.
func (s *Station) loop() {
	for {
		for _, track := range s.tracks {
			select {
			case <-s.stop:
				logrus.WithField("station", s.name).Info("station stopped")
				return
			default:
			}

			s.mu.Lock()
			s.current = filepath.Base(track)
			s.mu.Unlock()

			logrus.WithField("station", s.name).WithField("track", track).Info("playing track")
			f, err := os.Open(track)
			if err != nil {
				logrus.WithError(err).Error("open track")
				continue
			}
			bufSize := s.bps * int(s.buffer.Seconds())
			if bufSize <= 0 {
				bufSize = 32 * 1024
			}
			reader := bufio.NewReaderSize(f, bufSize)
			buf := make([]byte, 32*1024)
			for {
				select {
				case <-s.stop:
					f.Close()
					logrus.WithField("station", s.name).Info("station stopped")
					return
				default:
				}
				n, err := reader.Read(buf)
				if n > 0 {
					s.broadcast(buf[:n])
					sleep := time.Duration(int64(time.Second) * int64(n) / int64(s.bps))
					if sleep > 0 {
						time.Sleep(sleep)
					}
				}
				if err == io.EOF {
					break
				}
				if err != nil {
					logrus.WithError(err).Error("read track")
					break
				}
			}
			f.Close()
		}
	}
}

// broadcast sends data chunk to all listeners.
func (s *Station) broadcast(data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for ch := range s.listeners {
		select {
		case ch <- data:
		default:
		}
	}
}
