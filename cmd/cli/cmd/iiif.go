/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var iiifCmd = &cobra.Command{
	Use:   "iiif",
	Short: "Use the IIIF protocol to scrape a particular image or images, just pass the identifier",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("no IDs passed; nothing to do")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(iiifCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// iiifCmd.PersistentFlags().String("foo", "", "A help for foo")

	f := iiifCmd.Flags()

	f.String("host", "https://api.nga.gov/iiif", "The host to hit for the IIIF request")
	f.String("")
}
