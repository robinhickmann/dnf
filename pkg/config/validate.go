package config

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/rs/zerolog"
)

var timeFormats = map[string]string{
	"unix":      zerolog.TimeFormatUnix,
	"unixms":    zerolog.TimeFormatUnixMs,
	"unixmicro": zerolog.TimeFormatUnixMicro,
	"unixnano":  zerolog.TimeFormatUnixNano,

	"rfc3339":     time.RFC3339,
	"rfc3339nano": time.RFC3339Nano,
	"rfc1123":     time.RFC1123,
	"rfc1123z":    time.RFC1123Z,
	"rfc822":      time.RFC822,
	"rfc822z":     time.RFC822Z,
}

func (c *Config) validate() error {
	if err := c.DNS.validate(); err != nil {
		return err
	}

	if err := c.HTTP.validate(); err != nil {
		return err
	}

	if err := c.Log.validate(); err != nil {
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

	if err := d.Timeout.validate("dns"); err != nil {
		return err
	}

	return nil
}

func (z *Zone) validate() error {
	if err := domain("dns.zone.name", &z.Name); err != nil {
		return err
	}

	if err := email("dns.zone.email", &z.Email); err != nil {
		return err
	}

	if err := host("dns.zone.ipv4", z.IPv4); err != nil {
		return err
	}

	if err := host("dns.zone.ipv6", z.IPv6); err != nil {
		return err
	}

	if err := nameservers("dns.zone.nameservers", z.Nameservers); err != nil {
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

	if err := h.Timeout.validate("http"); err != nil {
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

func (t *Timeout) validate(srv string) error {
	if err := timeout(srv, "timeout.shutdown", t.Shutdown); err != nil {
		return err
	}

	if err := timeout(srv, "timeout.read", t.Read); err != nil {
		return err
	}

	if err := timeout(srv, "timeout.write", t.Write); err != nil {
		return err
	}

	if srv == "http" {
		if err := timeout(srv, "timeout.idle", t.Idle); err != nil {
			return err
		}
	}

	return nil
}

func (l *Log) validate() error {
	// Log level
	if err := require("log.level", l.Level); err != nil {
		return err
	}

	if _, err := zerolog.ParseLevel(l.Level); err != nil {
		return fmt.Errorf("log.level: %q is not a valid level", l.Level)
	}

	// Log output
	if err := require("log.output", l.Output); err != nil {
		return err
	}

	output := strings.ToLower(l.Output)
	if output != "stdout" && output != "stderr" {
		return fmt.Errorf("log.output: %q is not a valid output", l.Output)
	}

	// Log time format
	if err := require("log.time", l.TimeFormat); err != nil {
		return err
	}

	s, ok := timeFormats[l.TimeFormat]
	if !ok {
		return fmt.Errorf("log.time: %q is an unrecognized time format", l.TimeFormat)
	}
	l.TimeFormat = s

	return nil
}

// Helpers

func require(field string, value any) error {
	if reflect.DeepEqual(value, reflect.Zero(reflect.TypeOf(value)).Interface()) {
		return fmt.Errorf("%s field is required", field)
	}
	return nil
}

func timeout(srv, field string, value time.Duration) error {
	if value < time.Second || value > time.Minute {
		return fmt.Errorf("%s.%s field must be at least 1 second and at most 1 minute", srv, field)
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

func port(field string, value int) error {
	if value < 1 || value > 65535 {
		return fmt.Errorf("%s field must be between 1 and 65535", field)
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

func binds(field string, values []string) error {
	if err := require(field, values); err != nil {
		return err
	}

	for i, value := range values {
		if err := host(field, value); err != nil {
			return err
		}

		// format for IPv6
		if strings.Contains(value, ":") {
			values[i] = fmt.Sprintf("[%s]", value)
		}
	}

	return nil
}

func domain(field string, value *string) error {
	if err := require(field, value); err != nil {
		return err
	}

	*value = dns.Fqdn(*value)

	if len(*value) > 255 {
		return fmt.Errorf("%s field must be less than 256 characters", field)
	}

	labels := strings.Split(strings.TrimSuffix(*value, "."), ".")
	for _, label := range labels {
		if len(label) == 0 {
			return fmt.Errorf("%s field must not contain empty labels", field)
		}
		if len(label) > 63 {
			return fmt.Errorf("%s: %q label must be less than 64 characters", field, label)
		}
	}

	if len(labels) < 2 {
		return fmt.Errorf("%s field must contain at least 2 labels", field)
	}

	return nil
}

func email(field string, value *string) error {
	if err := domain(field, value); err != nil {
		return err
	}

	if strings.Contains(*value, "@") {
		return fmt.Errorf("%s field must use RNAME format and cant contain @", field)
	}

	if labels := strings.Split(*value, "."); len(labels) < 3 {
		return fmt.Errorf("%s field must contain at least 3 labels", field)
	}

	return nil
}

func nameservers(field string, values []string) error {
	if err := require(field, values); err != nil {
		return err
	}

	for i, value := range values {
		if err := domain(field, &value); err != nil {
			return err
		}

		values[i] = dns.Fqdn(value)
	}

	return nil
}
