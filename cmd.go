package main

import (
	"fmt"
	"health-check-on-go/libs/health_check"
	"os"

	"github.com/spf13/cobra"
)

type GlobalContext struct {
	healthCheckContext *health_check.HealthContext
}

var (
	// Used for flags.
	cfgFile string

	globalContext = GlobalContext{
		healthCheckContext: &health_check.HealthContext{},
	}

	rootCmd = &cobra.Command{
		Use:   "myapp",
		Short: "A demo application for Cobra and Viper",
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)

	}
}

func init() {
	// Add the persistent --config flag to the root command.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.myapp.yaml)")
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("serve called")
	},
}

var healthCheckCmd = &cobra.Command{
	Use:   "health-check",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		globalContext.healthCheckContext.RunAll()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(healthCheckCmd)
}
