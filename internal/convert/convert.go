package convert

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"x07-it/radiod/internal/config"
)

var supportedExt = map[string]struct{}{
	".mp3": {},
	".aac": {},
	".ogg": {},
}

// PrepareCache scans music directories, converts files via ffmpeg and returns map station->tracks.
func PrepareCache(cfg config.Config) (map[string][]string, error) {
	result := make(map[string][]string)

	dirs, err := os.ReadDir(cfg.MusicDir)
	if err != nil {
		return nil, err
	}

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

			if _, err := os.Stat(outputPath); err == nil {
				logrus.Infof("cache exists: %s", outputPath)
			} else {
				logrus.Infof("converting %s", inputPath)
				cmd := exec.Command(cfg.FFMpegPath, "-y", "-i", inputPath, "-b:a", cfg.OutputBitrate, outputPath)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					return nil, err
				}
			}

			result[station] = append(result[station], outputPath)
		}
	}

	return result, nil
}
