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
