package config

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DNS  DNS  `yaml:"dns"`
	HTTP HTTP `yaml:"http"`
}

type DNS struct {
	Port  int      `yaml:"port"`
	Binds []string `yaml:"binds"`
	Zone  Zone     `yaml:"zone"`
}

type Zone struct {
	Name string `yaml:"name"`
	IPv4 string `yaml:"ipv4"`
	IPv6 string `yaml:"ipv6"`
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
			Port:  5300,
			Binds: []string{"127.0.0.1", "::1"},
			Zone: Zone{
				Name: "dns.example.com",
				IPv4: "127.0.0.1",
				IPv6: "::1",
			},
		},
		HTTP: HTTP{
			Port:  8080,
			Binds: []string{"127.0.0.1", "::1"},
			TLS: TLS{
				Enabled:  true,
				CertFile: "cert.pem",
				KeyFile:  "key.pem",
			},
			Timeout: Timeout{
				Shutdown: 5 * time.Second,
			},
		},
	}
}

// Config Validation

func (c *Config) validate() error {
	if err := c.DNS.validate(); err != nil {
		return err
	}

	if err := c.HTTP.validate(); err != nil {
		return err
	}

	return nil
}

func (d *DNS) validate() error {
	if err := port("dns.port", d.Port); err != nil {
		return err
	}

	if err := binds("dns.binds", d.Binds); err != nil {
		return err
	}

	if err := d.Zone.validate(); err != nil {
		return err
	}

	return nil
}

func (z *Zone) validate() error {
	if err := require("dns.zone.name", z.Name); err != nil {
		return err
	}

	if err := host("dns.zone.ipv4", z.IPv4); err != nil {
		return err
	}

	if err := host("dns.zone.ipv6", z.IPv6); err != nil {
		return err
	}

	return nil
}

func (h *HTTP) validate() error {
	if err := port("http.port", h.Port); err != nil {
		return err
	}

	if err := binds("http.binds", h.Binds); err != nil {
		return err
	}

	if err := h.TLS.validate(); err != nil {
		return err
	}

	if err := h.Timeout.validate(); err != nil {
		return err
	}

	return nil
}

func (t *TLS) validate() error {
	if !t.Enabled {
		return nil
	}

	if err := file("http.tls.cert_file", t.CertFile); err != nil {
		return err
	}

	if err := file("http.tls.key_file", t.KeyFile); err != nil {
		return err
	}

	return nil
}

func (t *Timeout) validate() error {
	if err := timeout("http.timeout.shutdown", t.Shutdown); err != nil {
		return err
	}
	return nil
}

// Validation Helpers

func require(field, value string) error {
	if value == "" {
		return fmt.Errorf("%s field is required", field)
	}
	return nil
}

func binds(field string, values []string) error {
	if len(values) == 0 {
		return fmt.Errorf("%s field is required", field)
	}

	for _, value := range values {
		if err := host(field, value); err != nil {
			return err
		}
	}

	return nil
}

func host(field, value string) error {
	if err := require(field, value); err != nil {
		return err
	}
	if net.ParseIP(value) == nil {
		return fmt.Errorf("%s: %q is not a valid IPv4 or IPv6 address", field, value)
	}
	return nil
}

func port(field string, value int) error {
	if value < 1 || value > 65535 {
		return fmt.Errorf("%s field must be between 1 and 65535", field)
	}
	return nil
}

func timeout(field string, value time.Duration) error {
	if value < time.Second || value > time.Minute {
		return fmt.Errorf("%s field must be at least 1 second and at most 1 minute", field)
	}
	return nil
}

func file(field, value string) error {
	if err := require(field, value); err != nil {
		return err
	}

	stat, err := os.Stat(value)
	if err != nil {
		return fmt.Errorf("%s: %q does not exist", field, value)
	}

	if stat.IsDir() {
		return fmt.Errorf("%s: %q is a directory", field, value)
	}

	return nil
}
