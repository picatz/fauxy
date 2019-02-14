package fauxy

import (
	"errors"
	"io"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// TCP facilitates the connection(s) from one TCP endpoint to anouther.
type TCP struct {
	Config *Config
	done   chan struct{}

	workerPool []tcpWorker
	jobChannel chan *net.TCPConn
	quit       chan bool
}

// tcpWorker represents the worker that executes the connection
type tcpWorker struct {
	workerPool chan chan *net.TCPConn
	jobChannel chan *net.TCPConn
	quit       chan bool
}

func newTCPWorker(workerPool chan chan *net.TCPConn) tcpWorker {
	return tcpWorker{
		workerPool: workerPool,
		jobChannel: make(chan *net.TCPConn, runtime.NumCPU()),
		quit:       make(chan bool)}
}

func setupConfig(config *Config) {
	if config.Policies.Timeout <= 0 {
		config.Policies.Timeout = time.Millisecond * 1000
	}
}

func setupLogger(config *Config) {
	if config.Log.Stdout {
		log.SetOutput(os.Stdout)
	}
	if config.Log.Stderr {
		log.SetOutput(os.Stderr)
	}
	formatter := &logrus.TextFormatter{
		FullTimestamp: true,
	}
	logrus.SetFormatter(formatter)
}

// NewTCP needs to be documented.
func NewTCP(config *Config) Proxy {
	setupConfig(config)
	setupLogger(config)
	return &TCP{
		Config:     config,
		done:       make(chan struct{}),
		jobChannel: make(chan *net.TCPConn, runtime.NumCPU()),
		quit:       make(chan bool, runtime.NumCPU()),
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
	setupConfig(config)
	setupLogger(config)
	return &TCP{
		Config:     config,
		done:       make(chan struct{}),
		jobChannel: make(chan *net.TCPConn, runtime.NumCPU()),
		quit:       make(chan bool),
	}, nil
}

// Stop needs to be documented.
func (p *TCP) Stop() {
	log.Warn("Stopping proxy")
	allWorkersStopped := make(chan bool)
	go func() {
		for i := 1; i <= runtime.NumCPU(); i++ {
			p.quit <- true
		}
		allWorkersStopped <- true
	}()
	<-allWorkersStopped
}

// Start needs to be documented.
func (p *TCP) Start() error {
	log.WithFields(log.Fields{
		"config": p.Config,
	}).Info("Starting proxy")
	if p.done == nil {
		p.done = make(chan struct{})
	}

	addr, err := net.ResolveTCPAddr("tcp", p.Config.From)
	if err != nil {
		panic(err)
	}

	listener, err := net.ListenTCP("tcp", addr)

	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-p.done:
				return
			default:
				connection, err := listener.AcceptTCP()
				if err != nil {
					log.WithFields(log.Fields{
						"error": err.Error(),
					}).Warn("Error while accepting new TCP connections")
					continue
				}
				p.jobChannel <- connection
			}
		}
	}()
	for i := 1; i <= runtime.NumCPU(); i++ {
		log.Info("Starting worker", i)
		go func() {
			for {
				select {
				case connection := <-p.jobChannel:
					log.Info("Handling connection")
					connection.SetNoDelay(!p.Config.Policies.Nagle)
					connection.SetKeepAlive(p.Config.Policies.KeepAlive)
					p.handle(connection)
				case <-p.quit:
					log.Warn("Stopping worker")
					return
				}
			}
		}()
	}
	return nil
}

func (p *TCP) handle(connection *net.TCPConn) {
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
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Warn("Error during proxy connection")
		return
	}
	defer remote.Close()

	stop := make(chan struct{})

	go p.copy(connection, remote, stop)
	go p.copy(remote, connection, stop)

	select {
	case <-stop:
		return
	}
}

func (p *TCP) meetsConnectionPolicies(connection net.Conn) bool {
	if p.Config.Policies.DenyAll {
		connection.Close()
		log.WithFields(log.Fields{
			"ip": connection.RemoteAddr().String(),
		}).Info("Denied connection")
		return false
	}

	if p.Config.Policies.AllowAll {
		log.WithFields(log.Fields{
			"ip": connection.RemoteAddr().String(),
		}).Info("Allowed connection")
		return true
	}

	remoteIP := net.ParseIP(strings.Split(connection.RemoteAddr().String(), ":")[0])

	for _, allowIP := range p.Config.Policies.Allow {
		if remoteIP.Equal(allowIP) {
			log.WithFields(log.Fields{
				"ip": connection.RemoteAddr().String(),
			}).Info("Allowed connection")
			return true
		}
	}

	// if explicitly allowing only certain IPs, deny
	// everything that doesn't match
	if len(p.Config.Policies.Allow) > 0 {
		log.WithFields(log.Fields{
			"ip": connection.RemoteAddr().String(),
		}).Info("Denied connection")
		return false
	}

	for _, denyIP := range p.Config.Policies.Deny {
		if remoteIP.Equal(denyIP) {
			connection.Close()
			log.WithFields(log.Fields{
				"ip": connection.RemoteAddr().String(),
			}).Info("Denied connection")
			return false
		}
	}

	// allow all traffic by default
	log.WithFields(log.Fields{
		"ip": connection.RemoteAddr().String(),
	}).Info("Allowed connection")
	return true
}

func (p *TCP) copy(from, to net.Conn, stop chan struct{}) (int64, error) {
	defer to.Close()
	defer from.Close()
	written, err := io.Copy(to, from)
	if err != nil {
		log.WithFields(log.Fields{
			"written":     written,
			"error":       err,
			"source":      from.RemoteAddr().String(),
			"destination": to.RemoteAddr().String(),
		}).Warn("Error while copying bytes")
		return written, err
	}
	log.WithFields(log.Fields{
		"written":     written,
		"source":      from.RemoteAddr().String(),
		"destination": to.RemoteAddr().String(),
	}).Info("Copied bytes")
	stop <- struct{}{}
	return written, err
}
