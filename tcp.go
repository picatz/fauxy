package fauxy

import (
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

// TCP facilitates the connection(s) from one TCP endpoint to anouther.
type TCP struct {
	from    string
	to      string
	done    chan struct{}
	config  *Config
	timeout time.Duration
}

// NewTCP needs to be documented.
func NewTCP(from, to string, config *Config) Proxy {
	return &TCP{
		from:    from,
		to:      to,
		config:  config,
		timeout: time.Second * 3,
	}
}

// NewTCPWithConfigFile needs to be documented.
func NewTCPWithConfigFile(from, to, configFilename string) (Proxy, error) {
	var config *Config
	var err error
	if configFilename == "" {
		config = NewDefaultConfig()
	} else {
		config, err = NewConfigFromFile(configFilename)
		if err != nil {
			return nil, err
		}
	}

	return &TCP{
		from:    from,
		to:      to,
		config:  config,
		timeout: time.Second * 3,
	}, nil
}

// Stop needs to be documented.
func (p *TCP) Stop() error {
	if p.done == nil {
		return errors.New("tcp server already stopped")
	}
	close(p.done)
	p.done = nil
	return nil
}

// Start needs to be documented.
func (p *TCP) Start() error {
	if p.done == nil {
		p.done = make(chan struct{})
	}
	listener, err := net.Listen("tcp", p.from)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-p.done:
				return
			default:
				connection, err := listener.Accept()
				if err != nil {
					continue
				}
				p.handle(connection)
			}
		}
	}()
	return nil
}

func (p *TCP) handle(connection net.Conn) {
	defer connection.Close()

	// apply config policy

	if p.config != nil {
		if p.config.denyAll {
			return // block connection
		}

		remoteIP := net.ParseIP(strings.Split(connection.RemoteAddr().String(), ":")[0])

		if !p.config.allowAll {
			for _, allowIP := range p.config.Allow {
				if p.config.allowAll || remoteIP.Equal(allowIP) {
					break // allow connection, unless it's also in the deny list
				}
			}

			for _, denyIP := range p.config.Deny {
				if p.config.denyAll || remoteIP.Equal(denyIP) {
					return // block connection
				}
			}
		}
	}

	// handle connection
	remote, err := net.DialTimeout("tcp", p.to, p.timeout)
	//remote, err := net.Dial("tcp", p.to)
	if err != nil {
		return
	}
	defer remote.Close()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go p.copy(remote, connection, wg)
	go p.copy(connection, remote, wg)
	wg.Wait()
}

func (p *TCP) copy(from, to net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	select {
	case <-p.done:
		return
	default:
		if _, err := io.Copy(to, from); err != nil {
			//p.Stop()
			return
		}
	}
}
