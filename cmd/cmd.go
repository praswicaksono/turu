package cmd

import (
	"github.com/praswicaksono/turu/internal/conf"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "turu",
		Short: "Docker based service discovery registrator",
		Long: `Automate route and service registration to service discovery
		based on docker label`,
	}
)

// Execute executes the root command.
func Execute() error {
	rootCmd.AddCommand(listenCmd)
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(conf.Init)
}
