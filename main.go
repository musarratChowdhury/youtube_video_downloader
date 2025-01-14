package main

import (
	"encoding/json"
	"fmt"
	"github.com/kkdai/youtube/v2"
	"log"
	"os/exec"
	"net/http"
	"regexp"
	"strings"
)

// Data structure for JSON response
type Message struct {
	Text string `json:"text"`
}

// Home handler
func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	message := Message{Text: "Welcome to the Go API!"}
	json.NewEncoder(w).Encode(message)
}

func sanitizeFilename(title string) string {
	re := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	sanitized := re.ReplaceAllString(title, "")
	sanitized = strings.TrimSpace(sanitized)
	sanitized = strings.ReplaceAll(sanitized, " ", "_")
	return sanitized + ".mp3"
}
// Health check handler
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
func downloadWithFFmpeg(videoURL string, outputFileName string) error {
	// ffmpeg command to download the video
	cmd := exec.Command("ffmpeg", "-i", videoURL, "-map", "0:a", "-c:a", "libmp3lame", "-b:a", "128k", outputFileName)


	// Execute the command
	err := cmd.Run()
	return err
}
// YouTube video download handler
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	// Get video URL from query parameters
	videoURL := r.URL.Query().Get("url")
	if videoURL == "" {
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	
	client := youtube.Client{}

	// Fetch video information
	video, err := client.GetVideo(videoURL)
	
	
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get video: %v", err), http.StatusInternalServerError)
		return
	}

	// Select format with both video and audio
	format := video.Formats.WithAudioChannels()[0]
	streamUrl, err := client.GetStreamURL(video, &format)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get stream URL: %v", err), http.StatusInternalServerError)
		return
	}
	// Create a file to save the downloaded video
	outputFileName := sanitizeFilename(video.Title)
	err = downloadWithFFmpeg(streamUrl, outputFileName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create file: %v", err), http.StatusInternalServerError)
		return
	}


	message := Message{Text: "Video downloaded successfully as " + outputFileName}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/download", downloadHandler)

	// Start the server on port 8080
	log.Println("Server running on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
