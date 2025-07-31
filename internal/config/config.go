package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config holds server configuration parameters loaded from file or environment.
type Config struct {
	FFMpegPath    string `mapstructure:"ffmpeg_path"`
	OutputFormat  string `mapstructure:"output_format"`
	OutputBitrate string `mapstructure:"output_bitrate"`
	MusicDir      string `mapstructure:"music_dir"`
	CacheDir      string `mapstructure:"cache_dir"`
	BufferSeconds int    `mapstructure:"buffer_seconds"`
	Listen        string `mapstructure:"listen"`
}

// Load reads configuration from config.yaml and environment variables.
func Load() (Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	// Default values allow server to run out of the box.
	v.SetDefault("ffmpeg_path", "ffmpeg")
	v.SetDefault("output_format", "mp3")
	v.SetDefault("output_bitrate", "96k")
	v.SetDefault("music_dir", "./music")
	v.SetDefault("cache_dir", "./.cache")
	v.SetDefault("buffer_seconds", 7)
	v.SetDefault("listen", ":7000")

	v.AutomaticEnv()

	// Read configuration file if present.
	if err := v.ReadInConfig(); err != nil {
		// ignore error if config file not found
		logrus.WithError(err).Debug("config file not read")
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
