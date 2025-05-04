// Copyright (c) 2025 Alexis Bouchez <alexbcz@proton.me> (https://alexisbouchez.com), MIT License
package posts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/alexisbcz/yabc/internal/bluesky"
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
	yabc posts create --text "Hello world!" --hashtags coding,golang
	yabc posts create --text "Check out this photo" --image path/to/image.jpg`,
		Run: func(cmd *cobra.Command, args []string) {
			if text == "" && imageFile == "" {
				fmt.Println("Error: You must provide either text content or an image to post")
				return
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
			err = createPost(token, content, imageFile)
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

// createPost sends a request to create a new post on Bluesky
func createPost(token *bluesky.DIDResponse, content string, imagePath string) error {
	// Prepare the post record
	record := map[string]interface{}{
		"$type":     "app.bsky.feed.post",
		"text":      content,
		"createdAt": getCurrentTime(),
	}

	// Add image attachment if provided
	if imagePath != "" {
		// Here would be code to upload the image and get its blob reference
		// This implementation would depend on Bluesky's API for image uploads
		// For now we'll just log that this feature isn't implemented
		slog.Info("Image attachment requested but not implemented in this version")
		fmt.Println("Note: Image attachment functionality is not yet implemented")
	}

	// Create the request body
	requestBody := map[string]interface{}{
		"collection": "app.bsky.feed.post",
		"repo":       token.DID,
		"record":     record,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Send the request
	url := fmt.Sprintf("%s/com.atproto.repo.createRecord", API_URL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessJwt))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// getCurrentTime returns the current time in the format required by Bluesky
func getCurrentTime() string {
	// This is a simplified implementation - in production code,
	// use time.Now().Format(time.RFC3339) for proper ISO timestamp
	return time.Now().Format(time.RFC3339)
}
