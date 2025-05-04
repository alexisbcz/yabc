// Copyright (c) 2025 Alexis Bouchez <alexbcz@proton.me> (https://alexisbouchez.com), MIT License
package cmd

import (
	"os"

	"github.com/alexisbcz/yabc/cmd/posts"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "yabc",
	Short: "A simple CLI to interact with Bluesky",
	Long: `yabc is a simple CLI tool to interact with the Bluesky social network.
It allows users to perform common actions such as posting, browsing feeds,
and managing their accounts directly from the command line.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.AddCommand(posts.NewPostsCommand())
}
