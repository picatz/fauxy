package fauxy

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"time"
)

// Config needs to be documented
type Config struct {
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	Policies struct {
		AllowAll bool          `json:"allowAll,omitempty"`
		DenyAll  bool          `json:"denyAll,omitempty"`
		Allow    []net.IP      `json:"allow,omitempty"`
		Deny     []net.IP      `json:"deny,omitempty"`
		Timeout  time.Duration `json:"timeout,omitempty"`
		Nagle    bool          `json:"nagle,omitempty"`
	} `json:"policies,omitempty"`
	Hexdump bool `json:"hexdump,omitempty"`
	Monitor struct {
		BytesCopied bool `json:"bytes_copied,omitempty"`
	} `json:"monitor,omitempty"`
	Log struct {
		Stderr bool `json:"stderr,omitempty"`
		Stdout bool `json:"stdout,omitempty"`
		File   bool `json:"file,omitempty"`
	} `json:"log,omitempty"`
}

// NewDefaultConfig needs to be documented.
func NewDefaultConfig(from, to string) *Config {
	return &Config{
		From: from,
		To:   to,
	}
}

// NewConfigFromFile needs to be documented.
func NewConfigFromFile(filename string) (*Config, error) {
	file, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	config := &Config{}

	err = json.Unmarshal(file, config)

	if err != nil {
		return nil, err
	}

	return config, nil
}
