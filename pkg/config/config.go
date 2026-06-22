package config

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Flags struct {
	ConfigPath string
	VersionCmd bool
}

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
	Timeout struct {
		Shutdown time.Duration `yaml:"shutdown"`
	} `yaml:"timeout"`
}

type TLS struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// ParseFlags returns a new Flags struct with all the parsed cli flags.
func ParseFlags() *Flags {
	var flags Flags

	flag.StringVar(&flags.ConfigPath, "config", "", "Path to the config file")
	flag.BoolVar(&flags.VersionCmd, "version", false, "Print the version and exit")
	flag.Parse()

	return &flags
}

// NewConfig returns a new Config struct from the first available config file.
// Returns an error if no valid config file is found.
func NewConfig(configPath ...string) (*Config, error) {
	var err error

	for _, path := range configPath {
		if _, err = os.Stat(path); err == nil {
			config, err := decodeConfig(path)
			if err != nil {
				return nil, err
			}

			err = config.validate()
			if err != nil {
				return nil, fmt.Errorf("failed to validate config file: %s: %w", path, err)
			}

			return config, nil
		}
	}

	return nil, fmt.Errorf("no valid config file found")
}

func decodeConfig(configPath string) (*Config, error) {
	var cfg Config

	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %s: %w", configPath, err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(yamlFile))
	decoder.KnownFields(true)

	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %s: %w", configPath, err)
	}

	return &cfg, nil
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

	if err := timeout("http.timeout.shutdown", h.Timeout.Shutdown); err != nil {
		return err
	}

	if err := h.TLS.validate(); err != nil {
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
