package fauxy

import "testing"

func TestNewConfigFromFile(t *testing.T) {
	config, err := NewConfigFromFile("tmp/deny_all.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Error(config)
}
