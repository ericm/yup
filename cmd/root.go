package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var syncF string

// Root command for running the cli
var rootCmd = &cobra.Command{
	Use:   "yup",
	Short: "Yup is an easy to use AUR Helper",
	Long: `Yup is an easy to use AUR Helper with
				support for ncurses and search sorting
				Made by Eric Moynihan https://moynihan.io 
				/ https://github.com/ericm`,
	Run:     root,
	Version: "v0.0.1",
}

func init() {
	// Sync flag
	// rootCmd.Flags().StringVarP(&syncF, "sync", "S", "", "usage:  yup {-S --sync} [options] [package(s)]")
}

// Execute the root command
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("Error in command execution: %s", err)
	}

	return nil
}

func root(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		// Sync argument(s)
		sync(args)
	} else {
		// Send straight to pacman
	}
}
