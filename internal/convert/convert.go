package convert

import (
	"context"
	"io/fs"
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

	// errgroup manages goroutines and propagates the first error.
	g, ctx := errgroup.WithContext(context.Background())
	// semaphore limits number of concurrent ffmpeg processes.
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	// Walk through the music directory tree and schedule conversion for every audio file.
	err := filepath.WalkDir(cfg.MusicDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if _, ok := supportedExt[ext]; !ok {
			return nil
		}

		// Determine relative directory to build cache path and station names.
		rel, err := filepath.Rel(cfg.MusicDir, path)
		if err != nil {
			return err
		}
		relDir := filepath.Dir(rel)
		outName := strings.TrimSuffix(d.Name(), ext) + "." + cfg.OutputFormat
		outputPath := filepath.Join(cfg.CacheDir, relDir, outName)
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return err
		}

		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}

		in := path
		out := outputPath
		dir := relDir

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

			// Track belongs to every directory in its path; register it for each station.
			if dir != "." && dir != "" {
				mu.Lock()
				parts := strings.Split(dir, string(filepath.Separator))
				for _, st := range parts {
					result[st] = append(result[st], out)
				}
				mu.Unlock()
			}
			return nil
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}
