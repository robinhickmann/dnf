package config

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DNS  DNS  `yaml:"dns"`
	HTTP HTTP `yaml:"http"`
	Log  Log  `yaml:"log"`
}

type DNS struct {
	Port    int      `yaml:"port"`
	Binds   []string `yaml:"binds"`
	Zone    Zone     `yaml:"zone"`
	Timeout Timeout  `yaml:"timeout"`
}

type Zone struct {
	Name        string   `yaml:"name"`
	Email       string   `yaml:"email"`
	IPv4        string   `yaml:"ipv4"`
	IPv6        string   `yaml:"ipv6"`
	Nameservers []string `yaml:"nameservers"`
}

type HTTP struct {
	Port    int      `yaml:"port"`
	Binds   []string `yaml:"binds"`
	TLS     TLS      `yaml:"tls"`
	Timeout Timeout  `yaml:"timeout"`
}

type TLS struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type Timeout struct {
	Shutdown time.Duration `yaml:"shutdown"`
	Idle     time.Duration `yaml:"idle"`
	Read     time.Duration `yaml:"read"`
	Write    time.Duration `yaml:"write"`
}

type Log struct {
	Level      string `yaml:"level"`
	Output     string `yaml:"output"`
	TimeFormat string `yaml:"time"`
}

// NewConfig returns a new Config struct from the first available config file.
// Returns an error if no valid config file is found.
func NewConfig(configPath string) (*Config, error) {
	if configPath != "" {
		return loadConfig(configPath)
	}

	for _, path := range []string{"config.yaml", "config.yml"} {
		if _, err := os.Stat(path); err == nil {
			return loadConfig(path)
		}
	}

	return defaultConfig(), nil
}

func loadConfig(path string) (*Config, error) {
	cfg := defaultConfig()

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %s: %w", path, err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(yamlFile))
	decoder.KnownFields(true)

	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %s: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config file: %s: %w", path, err)
	}

	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		DNS: DNS{
			Port:  53,
			Binds: []string{"0.0.0.0"},
			Zone: Zone{
				Name:  "dns.example.com",
				Email: "admin.example.com",
				IPv4:  "127.0.0.1",
				IPv6:  "::1",
				Nameservers: []string{
					"ns1.example.com",
					"ns2.example.com",
				},
			},
			Timeout: Timeout{
				Shutdown: 5 * time.Second,
				Read:     10 * time.Second,
				Write:    10 * time.Second,
			},
		},
		HTTP: HTTP{
			Port:  8080,
			Binds: []string{"0.0.0.0"},
			TLS: TLS{
				Enabled:  true,
				CertFile: "cert.pem",
				KeyFile:  "key.pem",
			},
			Timeout: Timeout{
				Shutdown: 5 * time.Second,
				Idle:     5 * time.Second,
				Read:     10 * time.Second,
				Write:    10 * time.Second,
			},
		},
		Log: Log{
			Level:      "info",
			Output:     "stderr",
			TimeFormat: "unix",
		},
	}
}
