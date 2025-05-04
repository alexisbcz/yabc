package posts

import (
	"fmt"
	_ "image/gif"  // Support gif format
	_ "image/jpeg" // Support jpeg format
	_ "image/png"  // Support png format
	"log/slog"
	"os"

	"github.com/alexisbcz/yabc/internal/bluesky"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var (
	API_URL = "https://bsky.social/xrpc"
)

var (
	text      string
	hashtags  []string
	imageFile string
)

func newCreatePostCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new post on Bluesky",
		Long: `Create a new post on the Bluesky social network.
		
You can include text content, hashtags, and optionally attach an image.
		
Example usage:
    yabc posts create
	yabc posts create --text "Hello world!" --hashtags coding,golang
	yabc posts create --text "Check out this photo" --image path/to/image.jpg`,
		Run: func(cmd *cobra.Command, args []string) {
			if text == "" && imageFile == "" {
				err := huh.NewInput().
					Title("Type text content for the post").
					Placeholder("Hello world!").
					Value(&text).
					Run()
				if err != nil {
					slog.Error("Failed to get user input", "error", err)
					os.Exit(1)
				}
			}

			// Format text with hashtags if provided
			content := text
			for _, tag := range hashtags {
				content += fmt.Sprintf(" #%s", tag)
			}

			// Get authentication token
			token, err := bluesky.GetToken()
			if err != nil {
				slog.Error("Failed to get authentication token", "error", err)
				fmt.Println("Error: Failed to authenticate with Bluesky")
				return
			}

			// Create the post
			err = bluesky.CreatePost(token, content, imageFile)
			if err != nil {
				slog.Error("Failed to create post", "error", err)
				fmt.Println("Error: Failed to create post")
				return
			}

			fmt.Println("Post created successfully!")
		},
	}

	cmd.Flags().StringVarP(&text, "text", "t", "", "Text content for the post")
	cmd.Flags().StringSliceVarP(&hashtags, "hashtags", "a", []string{}, "Comma-separated list of hashtags (without # symbol)")
	cmd.Flags().StringVarP(&imageFile, "image", "i", "", "Path to image file to attach to the post")

	return cmd
}
