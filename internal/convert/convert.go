package convert

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"x07-it/radiod/internal/config"
)

var supportedExt = map[string]struct{}{
	".mp3": {},
	".aac": {},
	".ogg": {},
}

// PrepareCache scans music directories, converts files via ffmpeg and returns map station->tracks.
func PrepareCache(cfg config.Config) (map[string][]string, error) {
	// result collects converted file paths per station.
	result := make(map[string][]string)
	var mu sync.Mutex

	dirs, err := os.ReadDir(cfg.MusicDir)
	if err != nil {
		return nil, err
	}

	// errgroup manages goroutines and propagates the first error.
	g, ctx := errgroup.WithContext(context.Background())
	// semaphore limits number of concurrent ffmpeg processes.
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		station := d.Name()
		srcDir := filepath.Join(cfg.MusicDir, station)
		dstDir := filepath.Join(cfg.CacheDir, station)
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return nil, err
		}

		files, err := os.ReadDir(srcDir)
		if err != nil {
			return nil, err
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(f.Name()))
			if _, ok := supportedExt[ext]; !ok {
				continue
			}
			inputPath := filepath.Join(srcDir, f.Name())
			outName := strings.TrimSuffix(f.Name(), ext) + "." + cfg.OutputFormat
			outputPath := filepath.Join(dstDir, outName)

			// Capture variables for goroutine.
			st := station
			in := inputPath
			out := outputPath

			if err := sem.Acquire(ctx, 1); err != nil {
				return nil, err
			}

			g.Go(func() error {
				defer sem.Release(1)

				if _, err := os.Stat(out); err == nil {
					logrus.Infof("cache exists: %s", out)
				} else {
					logrus.Infof("converting %s", in)
					cmd := exec.CommandContext(ctx, cfg.FFMpegPath, "-y", "-i", in, "-b:a", cfg.OutputBitrate, out)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					if err := cmd.Run(); err != nil {
						return err
					}
				}

				mu.Lock()
				result[st] = append(result[st], out)
				mu.Unlock()
				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}
