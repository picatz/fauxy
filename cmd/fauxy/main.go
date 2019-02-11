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
			if cmdProxyFrom == "" || !strings.Contains(cmdProxyFrom, ":") {
				fmt.Println("invalid or missing '--from' flag")
				os.Exit(1)
			}

			if cmdProxyTo == "" || !strings.Contains(cmdProxyTo, ":") {
				fmt.Println("invalid or missing '--to' flag")
				os.Exit(1)
			}

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
				config = fauxy.NewDefaultConfig()
			}

			tcpProxy := fauxy.NewTCP(
				cmdProxyFrom, // from
				cmdProxyTo,   // to
				config,       // config
			)

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
