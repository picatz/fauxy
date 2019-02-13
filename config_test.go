package fauxy

import "testing"

func TestNewDefaultConfig(t *testing.T) {
	config := NewDefaultConfig("127.0.0.1:80", "127.0.0.1:8080")
	if config.From != "127.0.0.1:80" {
		t.Error("got unexpected `From` value in config")
	}
	if config.To != "127.0.0.1:8080" {
		t.Error("got unexpected `To` value in config")
	}
}

func TestNewConfigFromFile(t *testing.T) {
	config, err := NewConfigFromFile("tmp/example.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Error(config)
}
