package cmd

import (
	"log"

	"github.com/liyanbing/my-gokit/props"
	"github.com/liyanbing/my-gokit/timi/client"
	"github.com/liyanbing/my-gokit/timi/server"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "Timi service cmd",
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "start Timi client",
	Run: func(cmd *cobra.Command, args []string) {
		client.Client()
	},
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start Timi server",
	Run: func(cmd *cobra.Command, args []string) {
		server.Server()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	// root
	rootCmd.PersistentFlags().StringVarP(&props.ConfFilePath, "config-path", "", "./conf/Timi-local.conf", "config file path")
	rootCmd.PersistentFlags().StringVarP(&props.ConsulAddress, "consul-addr", "", "", "consul address")
	rootCmd.PersistentFlags().StringVarP(&props.ConsulSchema, "consul-schema", "", "", "consul schema")
	rootCmd.PersistentFlags().StringVarP(&props.ConsulDataCenter, "consul-data-center", "", "", "consul data center")
	rootCmd.PersistentFlags().Int64VarP(&props.ConsulWaitTime, "consul-wait-time", "", 0, "consul wait time")
	rootCmd.PersistentFlags().StringVarP(&props.ConsulToken, "consul-token", "", "", "consul token")
	rootCmd.PersistentFlags().StringVarP(&props.ConfNode, "config-node", "", "", "config node in consul")
	// server and client
	rootCmd.AddCommand(serverCmd, clientCmd)
}
