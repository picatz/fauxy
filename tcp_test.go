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
