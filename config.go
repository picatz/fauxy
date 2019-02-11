package fauxy

import (
	"encoding/json"
	"io/ioutil"
	"net"
)

// Config needs to be documented.
type Config struct {
	From net.Addr

	Allow []net.IP
	Deny  []net.IP

	allowAll bool
	denyAll  bool
}

// Config needs to be documented.
type configMiddle struct {
	Allow []string
	Deny  []string
}

// NewDefaultConfig needs to be documented.
func NewDefaultConfig() *Config {
	return &Config{
		allowAll: true,
	}
}

// NewConfigFromFile needs to be documented.
func NewConfigFromFile(filename string) (*Config, error) {
	plan, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	configM := &configMiddle{}
	err = json.Unmarshal(plan, configM)
	if err != nil {
		return nil, err
	}

	config := &Config{}

	var (
		allowAll bool
		denyAll  bool
	)

	for _, v := range configM.Allow {
		if v == "*" { // allow all
			configM.Allow = nil
			allowAll = true
		}
	}

	for _, v := range configM.Deny {
		if v == "*" { // allow all
			configM.Deny = nil
			denyAll = true
		}
	}

	out, err := json.Marshal(configM)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(out, config)
	if err != nil {
		return nil, err
	}

	config.allowAll = allowAll
	config.denyAll = denyAll

	return config, nil
}
