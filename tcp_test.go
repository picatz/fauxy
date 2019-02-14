package fauxy

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"testing"
)

func TestTCP(t *testing.T) {
	wg := &sync.WaitGroup{}

	go func() {
		server, err := net.Listen("tcp", "127.0.0.1:8888")
		if err != nil {
			t.Fatal(err)
		}
		for {
			conn, err := server.Accept()
			if err != nil {
				t.Fatal(err)
			}
			buff := make([]byte, 14)
			conn.Read(buff)
			fmt.Println("server read:", string(buff))
			go func() {
				defer conn.Close()
				conn.Write([]byte("hello world\n"))
			}()
		}
	}()

	config, err := NewConfigFromFile("tmp/tcp_test.json")

	if err != nil {
		t.Fatal(err)
	}

	tcpProxy := NewTCP(config)

	tcpProxy.Start()

	wg.Add(1)

	go func() {
		defer wg.Done()
		conn, err := net.Dial("tcp", "127.0.0.1:7777")
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		conn.Write([]byte("anybody there?"))
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		fmt.Print("Message received through TCP proxy:", string(message))
	}()

	wg.Wait()

	tcpProxy.Stop()
}

func TestTCP_Stop(t *testing.T) {
	type fields struct {
		Config *Config
		done   chan struct{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &TCP{
				Config: tt.fields.Config,
				done:   tt.fields.done,
			}
			if err := p.Stop(); (err != nil) != tt.wantErr {
				t.Errorf("TCP.Stop() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTCP_Start(t *testing.T) {
	type fields struct {
		Config *Config
		done   chan struct{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &TCP{
				Config: tt.fields.Config,
				done:   tt.fields.done,
			}
			if err := p.Start(); (err != nil) != tt.wantErr {
				t.Errorf("TCP.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
