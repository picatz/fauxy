package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/picatz/fauxy"
	"github.com/spf13/cobra"
)

func main() {
	// handle CTRL+C quit
	cleanup := func() {
		os.Exit(0)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			cleanup()
		}
	}()

	var (
		cmdProxyFrom   string
		cmdProxyTo     string
		cmdProxyConfig string
	)

	var cmdProxy = &cobra.Command{
		Use:   "proxy",
		Short: "Proxy connections from one endpoint to anouther",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			wg := &sync.WaitGroup{}

			var config *fauxy.Config
			var err error

			if cmdProxyConfig != "" {
				config, err = fauxy.NewConfigFromFile(cmdProxyConfig)
				if err != nil {
					fmt.Println("config from file:", cmdProxyConfig, "error:", err)
					os.Exit(1)
				}
			} else {
				config = fauxy.NewDefaultConfig(cmdProxyFrom, cmdProxyTo)
			}

			config.From = cmdProxyFrom
			config.To = cmdProxyTo

			passedFromToCheck := true

			if !strings.Contains(config.From, ":") {
				fmt.Println("invalid or missing '--from' flag")
				passedFromToCheck = false
			}

			if !strings.Contains(config.To, ":") {
				fmt.Println("invalid or missing '--to' flag")
				passedFromToCheck = false
			}

			if !passedFromToCheck {
				os.Exit(1)
			}

			tcpProxy := fauxy.NewTCP(config)

			tcpProxy.Start()

			wg.Add(1)

			wg.Wait()
		},
	}

	cmdProxy.Flags().StringVar(&cmdProxyFrom, "from", "", "Endpoint to listen on (ip:port).")
	cmdProxy.Flags().StringVar(&cmdProxyTo, "to", "", "Endpoint to send connections to (ip:port).")
	cmdProxy.Flags().StringVar(&cmdProxyConfig, "config", "", "Specify filename for config.")

	var rootCmd = &cobra.Command{Use: "fauxy"}
	rootCmd.AddCommand(cmdProxy)
	rootCmd.Execute()
}
