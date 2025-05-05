// Copyright (c) 2025 Alexis Bouchez <alexbcz@proton.me> (https://alexisbouchez.com), MIT License
package posts

import "github.com/spf13/cobra"

func NewPostsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "posts",
		Short: "Manage posts on Bluesky",
	}
	cmd.AddCommand(newCreatePostCommand())

	return cmd
}
