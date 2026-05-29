package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/brianpowell/on-air/internal/config"
)

func brokerOptions(cfg config.MQTT) (broker string, tlsConfig *tls.Config, err error) {
	u, err := url.Parse(cfg.Broker)
	if err != nil {
		return "", nil, fmt.Errorf("parse mqtt.broker: %w", err)
	}
	if u.Host == "" {
		return "", nil, fmt.Errorf("parse mqtt.broker: missing host")
	}

	if !cfg.MQTTTLSEnabled() {
		return cfg.Broker, nil, nil
	}

	// Paho only TLS-wraps ssl/tls/mqtts schemes; tcp:// is always plain MQTT.
	scheme := strings.ToLower(u.Scheme)
	switch scheme {
	case "ssl", "tls", "mqtts", "mqtt+ssl", "tcps":
		if scheme == "ssl" {
			u.Scheme = "tls"
		}
	default:
		return "", nil, fmt.Errorf("mqtt.broker: TLS required but scheme is %q (use ssl:// or tls://)", u.Scheme)
	}

	tlsConfig = &tls.Config{
		MinVersion:         tls.VersionTLS12,
		ServerName:         u.Hostname(),
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
	}

	if cfg.TLS.CAFile != "" {
		pool, err := loadCertPool(cfg.TLS.CAFile)
		if err != nil {
			return "", nil, err
		}
		tlsConfig.RootCAs = pool
	}

	return u.String(), tlsConfig, nil
}

func loadCertPool(path string) (*x509.CertPool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read mqtt.tls.ca_file: %w", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("mqtt.tls.ca_file: no certificates found in %s", path)
	}

	return pool, nil
}
