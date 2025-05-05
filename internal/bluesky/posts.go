// Copyright (c) 2025 Alexis Bouchez <alexbcz@proton.me> (https://alexisbouchez.com), MIT License
package bluesky

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CreatePost sends a request to create a new post on Bluesky
func CreatePost(token *DIDResponse, content string, imagePath string) error {
	// Prepare the post record
	record := map[string]interface{}{
		"$type":     "app.bsky.feed.post",
		"text":      content,
		"createdAt": getCurrentTime(),
	}

	// Add image attachment if provided
	if imagePath != "" {
		fmt.Println("Uploading image:", imagePath)

		blobResp, err := uploadImage(token, imagePath)
		if err != nil {
			return fmt.Errorf("failed to upload image: %w", err)
		}

		// Get image dimensions for aspect ratio if possible
		width, height, err := getImageDimensions(imagePath)
		if err != nil {
			slog.Warn("Could not determine image dimensions", "error", err)
			fmt.Println("Warning: Could not determine image dimensions, aspect ratio won't be specified")
		}

		// Prepare the image embed
		imageEmbed := map[string]interface{}{
			"alt": "Attached image", // Default alt text
			"image": map[string]interface{}{
				"$type":    "blob",
				"ref":      map[string]string{"$link": blobResp.Blob.Ref.Link},
				"mimeType": blobResp.Blob.MimeType,
				"size":     blobResp.Blob.Size,
			},
		}

		// Add aspect ratio if we have dimensions
		if width > 0 && height > 0 {
			imageEmbed["aspectRatio"] = map[string]int{
				"width":  width,
				"height": height,
			}
		}

		// Add the image to the post record
		record["embed"] = map[string]interface{}{
			"$type":  "app.bsky.embed.images",
			"images": []map[string]interface{}{imageEmbed},
		}
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
		var errResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			slog.Error("API error response", "response", errResp)
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode the response to get the post details
	var postResp PostCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&postResp); err == nil {
		slog.Info("Post created", "uri", postResp.URI, "cid", postResp.CID)
	}

	return nil
}

// getCurrentTime returns the current time in the format required by Bluesky
func getCurrentTime() string {
	// Bluesky prefers the "Z" format over "+00:00"
	return time.Now().UTC().Format(time.RFC3339Nano)[:23] + "Z"
}

// uploadImage uploads an image to Bluesky and returns a blob reference
func uploadImage(token *DIDResponse, imagePath string) (*UploadBlobResponse, error) {
	// Check if file exists and is accessible
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("image file does not exist: %s", imagePath)
	} else if err != nil {
		return nil, fmt.Errorf("cannot access image file: %w", err)
	}

	// Read the file content instead of keeping it open
	imgData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file: %w", err)
	}

	// Check file size - Bluesky has a 1MB limit
	if len(imgData) > 1000000 {
		return nil, fmt.Errorf("image file size too large: %d bytes (1,000,000 bytes maximum)", len(imgData))
	}

	// Determine MIME type
	mimeType := getMimeType(imagePath)
	if mimeType == "" {
		// Default to a common image type if we can't determine
		mimeType = "image/jpeg"
	}

	slog.Info("Uploading image", "path", imagePath, "size", len(imgData), "mimeType", mimeType)

	// According to the Bluesky docs, we should send the raw image bytes directly, not as multipart
	url := fmt.Sprintf("%s/com.atproto.repo.uploadBlob", API_URL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(imgData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessJwt))
	req.Header.Set("Content-Type", mimeType)

	// Send the request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body for better error messages
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Try to parse error response
		var errResp map[string]interface{}
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			slog.Error("API error response", "response", errResp)
			if message, ok := errResp["message"].(string); ok {
				return nil, fmt.Errorf("API error: %s", message)
			}
		}
		return nil, fmt.Errorf("unexpected status code: %d - %s", resp.StatusCode, string(respBody))
	}

	// Decode the response
	var blobResp UploadBlobResponse
	if err := json.Unmarshal(respBody, &blobResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w - body: %s", err, string(respBody))
	}

	if blobResp.Blob.Ref.Link == "" {
		return nil, fmt.Errorf("invalid response: missing blob reference link - body: %s", string(respBody))
	}

	slog.Info("Image uploaded successfully", "blob_link", blobResp.Blob.Ref.Link, "size", len(imgData))
	return &blobResp, nil
}

// getMimeType tries to determine the MIME type of a file based on its extension
func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		// Try to infer from file content if needed
		return ""
	}
}

// getImageDimensions tries to determine the width and height of an image file
func getImageDimensions(imagePath string) (int, int, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to open image for dimension detection: %w", err)
	}
	defer file.Close()

	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image dimensions: %w", err)
	}

	return img.Width, img.Height, nil
}

// Updated response structure
type UploadBlobResponse struct {
	Blob struct {
		Type     string  `json:"$type,omitempty"`
		Ref      RefLink `json:"ref,omitempty"`
		MimeType string  `json:"mimeType"`
		Size     int64   `json:"size"`
	} `json:"blob"`
}

// PostCreateResponse is the response from creating a post
type PostCreateResponse struct {
	URI    string `json:"uri"`
	CID    string `json:"cid"`
	Record struct {
		Type      string `json:"$type"`
		Text      string `json:"text"`
		CreatedAt string `json:"createdAt"`
	} `json:"record"`
}

// Updated to match the actual API response structure
type BlobReference struct {
	Type     string  `json:"$type,omitempty"`
	Ref      RefLink `json:"ref,omitempty"`
	MimeType string  `json:"mimeType"`
	Size     int64   `json:"size"`
}

type RefLink struct {
	Link string `json:"$link"`
}
