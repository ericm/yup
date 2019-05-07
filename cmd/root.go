package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "yup",
	Short: "Yup is an easy to use AUR Helper",
	Long: `Yup is an easy to use AUR Helper with
				support for ncurses and search sorting
				Made by Eric Moynihan https://moynihan.io 
				/ https://github.com/ericm`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Root Command")
	},
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("Error in command execution: %s", err)
	}

	return nil
}
