package fauxy

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

// TCP facilitates the connection(s) from one TCP endpoint to anouther.
type TCP struct {
	Config *Config
	done   chan struct{}
}

// NewTCP needs to be documented.
func NewTCP(config *Config) Proxy {
	if config.Log.Stdout {
		log.SetOutput(os.Stdout)
	}
	if config.Log.Stderr {
		log.SetOutput(os.Stderr)
	}
	return &TCP{
		Config: config,
		done:   make(chan struct{}),
	}
}

// NewTCPWithConfigFile needs to be documented.
func NewTCPWithConfigFile(from, to, configFilename string) (Proxy, error) {
	var config *Config
	var err error
	if configFilename == "" {
		return nil, errors.New("no filename given")
	}
	config, err = NewConfigFromFile(configFilename)
	if err != nil {
		return nil, err
	}

	return &TCP{
		Config: config,
		done:   make(chan struct{}),
	}, nil
}

// Stop needs to be documented.
func (p *TCP) Stop() error {
	log.Warn("Stopping proxy")
	if p.done == nil {
		return errors.New("tcp server already stopped")
	}
	close(p.done)
	p.done = nil
	return nil
}

// Start needs to be documented.
func (p *TCP) Start() error {
	log.WithFields(log.Fields{
		"config": p.Config,
	}).Info("Starting proxy")
	if p.done == nil {
		p.done = make(chan struct{})
	}
	listener, err := net.Listen("tcp", p.Config.From)
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
	log.Info("Started proxy")
	return nil
}

func (p *TCP) handle(connection net.Conn) {
	defer connection.Close()

	if !p.meetsConnectionPolicies(connection) {
		log.WithFields(log.Fields{
			"ip":       connection.RemoteAddr().String(),
			"policies": p.Config.Policies,
		}).Warn("Failed to meet policies")
		return
	}

	remote, err := net.DialTimeout("tcp", p.Config.To, p.Config.Policies.Timeout)
	if err != nil {
		return
	}
	defer remote.Close()

	wg := &sync.WaitGroup{}
	wg.Add(2)
	if p.Config.Monitor.From {
		fmt.Println("monitoring from")
		go func() {
			written, err := p.copy(connection, remote, wg)
			if err != nil {
				log.WithFields(log.Fields{
					"error":       err.Error(),
					"source":      remote.RemoteAddr().String(),
					"destination": connection.RemoteAddr().String(),
				}).Info("monitorFrom")
			}
			log.WithFields(log.Fields{
				"written":     written,
				"source":      remote.RemoteAddr().String(),
				"destination": connection.RemoteAddr().String(),
			}).Info("monitorFrom")
		}()
	} else {
		go p.copy(connection, remote, wg)
	}
	if p.Config.Monitor.To {
		fmt.Println("monitoring to")
		go func() {
			written, err := p.copy(remote, connection, wg)
			if err != nil {
				log.WithFields(log.Fields{
					"error":       err.Error(),
					"source":      remote.RemoteAddr().String(),
					"destination": connection.RemoteAddr().String(),
				}).Info("monitorFrom")
			}
			log.WithFields(log.Fields{
				"written":     written,
				"destination": connection.RemoteAddr().String(),
				"source":      remote.RemoteAddr().String(),
			}).Info("monitorTo")
		}()
	} else {
		go p.copy(remote, connection, wg)
	}

	wg.Wait()

	log.WithFields(log.Fields{
		"ip": connection.RemoteAddr().String(),
	}).Info("Processed connection")
}

func (p *TCP) meetsConnectionPolicies(connection net.Conn) bool {
	if p.Config.Policies.DenyAll {
		connection.Close()
		return false
	}

	if p.Config.Policies.AllowAll {
		return true
	}

	remoteIP := net.ParseIP(strings.Split(connection.RemoteAddr().String(), ":")[0])

	for _, allowIP := range p.Config.Policies.Allow {
		if remoteIP.Equal(allowIP) {
			return true
		}
	}

	// if explicitly allowing only certain IPs, deny
	// everything that doesn't match
	if len(p.Config.Policies.Allow) > 0 {
		return false
	}

	for _, denyIP := range p.Config.Policies.Deny {
		if remoteIP.Equal(denyIP) {
			connection.Close()
			return false
		}
	}

	// allow all traffic by default
	return true
}

func (p *TCP) copy(from, to net.Conn, wg *sync.WaitGroup) (int64, error) {
	defer wg.Done()
	select {
	case <-p.done:
		return int64(0), nil
	default:
		return io.Copy(to, from)
	}
}
