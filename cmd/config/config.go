package config

import (
	"flag"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

const (
	defaultHost = ":8080"
	defaultBase = "http://localhost:8080"
)

type Configuration struct {
	ServeAddress string
	BaseURL      string
}

func Load() (*Configuration, error) {
	var cfg Configuration
	flag.StringVar(&cfg.ServeAddress, "a", defaultHost, "Address to listen on")
	flag.StringVar(&cfg.BaseURL, "b", defaultBase, "Base URL for shorted links")
	flag.Parse()
	err := cfg.Validate()
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Configuration) Validate() error {
	if err := c.validateServeAddress(); err != nil {
		return err
	}
	if err := c.validateBaseURL(); err != nil {
		return err
	}
	return nil
}

func (c *Configuration) validateServeAddress() error {
	addr := c.ServeAddress
	if !strings.Contains(addr, ":") {
		addr = ":" + addr
	}

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid address: %s", addr)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid port: %s", portStr)
	}

	if host != "localhost" && host != "" && net.ParseIP(host) == nil {
		return fmt.Errorf("invalid host: %s", host)
	}

	return nil
}

func (c *Configuration) validateBaseURL() error {
	baseURL := c.BaseURL
	if !strings.Contains(baseURL, "://") {
		baseURL = "http://" + baseURL
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("scheme must be http or https")
	}
	if u.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	hostname := u.Hostname()
	if hostname == "" {
		return fmt.Errorf("invalid hostname")
	}

	if u.Path != "" && u.Path != "/" {
		return fmt.Errorf("URL must not contain a path (got %q)", u.Path)
	}
	if u.RawQuery != "" {
		return fmt.Errorf("URL must not contain query parameters")
	}
	if u.Fragment != "" {
		return fmt.Errorf("URL must not contain a fragment")
	}

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	if strings.HasSuffix(u.Host, ":") {
		return fmt.Errorf("invalid host - trailing colon without port (use 'host:port' or ':port')")
	}

	if prt := u.Port(); prt != "" {
		if p, err := strconv.Atoi(prt); err != nil || p < 1 || p > 65535 {
			return fmt.Errorf("port must be 1-65535")
		}
	}
	c.BaseURL = baseURL
	return nil
}
