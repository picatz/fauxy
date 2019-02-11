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
		go func() {
			for {
				conn, err := server.Accept()
				if err != nil {
					t.Fatal(err)
				}
				go func() {
					defer conn.Close()
					conn.Write([]byte("hello world\n"))
				}()
			}
		}()
	}()

	tcpProxy := NewTCP(
		"127.0.0.1:7777",   // from
		"127.0.0.1:8888",   // to
		NewDefaultConfig(), // config
	)

	tcpProxy.Start()

	wg.Add(1)

	go func() {
		defer wg.Done()
		conn, err := net.Dial("tcp", "127.0.0.1:7777")
		if err != nil {
			t.Fatal(err)
		}
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		fmt.Print("Message received through TCP proxy:", string(message))
	}()

	wg.Wait()

	tcpProxy.Stop()
}
