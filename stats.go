// please do not edit anything in the code if you dont know what your doing.

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Track struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Duration int    `json:"duration_ms"`
	Artists  []struct {
		Name string `json:"name"`
	} `json:"artists"`
	Album struct {
		Name string `json:"name"`
	} `json:"album"`
}

type WebhookPayload struct {
	Embeds []Embed `json:"embeds"`
}

type Embed struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Color       int     `json:"color"`
	Fields      []Field `json:"fields"`
	Footer      Footer  `json:"footer"`
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type Footer struct {
	Text string `json:"text"`
}

const (
	webhookURL          = "https://discord.com/api/webhooks/1344706181331161169/lEqlmf_wCnTonEPHY4qKdJ-Ac54r-W2xoPENRl9roxNAjjSYKmirkG2eHBJZ62p67RYo" // please do not mess with my webhook i only use it to track who and what your doing with my tool. no personal info is tracked i will list what im tracking (Hostname,OS,Filename,Country,Track,Artist,Album,Total Streams,Date Range,End Year,Start Year,Custom Name, Bulk mode,Max Density,Total plays.) if you dont want me to track those information feel free to delete the webhook.)  
	spotifyClientID     = "ac9ce18ca7d1475aaff975e02eba914e" // please do not edit/delete this it will break features
	spotifyClientSecret = "734cbce033ed4c668fe17d610f130f98" // please do not edit/delete this it will break features
	toolVersion         = "2.1.0"
)

var hostname string

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		hostname = "unknown-host"
	}
}

func extractTrackID(input string) (string, error) {
	if !strings.Contains(input, "track/") {
		return "", fmt.Errorf("invalid Spotify track link")
	}
	trackIndex := strings.Index(input, "track/")
	idStart := trackIndex + len("track/")
	idEnd := strings.Index(input[idStart:], "?")
	if idEnd == -1 {
		return input[idStart:], nil
	}
	return input[idStart : idStart+idEnd], nil
}

func getSpotifyAccessToken() (string, error) {
	data := "grant_type=client_credentials"
	req, _ := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(spotifyClientID, spotifyClientSecret)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result["access_token"].(string), nil
}

func getTrackDetails(accessToken, trackID string) (*Track, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.spotify.com/v1/tracks/%s", trackID), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var track Track
	err = json.NewDecoder(resp.Body).Decode(&track)
	return &track, err
}

func sanitizeFilename(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(name, ".", ""), " ", "_")
}

func getCountry() string {
	resp, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		return "unknown-country"
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if country, ok := result["country"].(string); ok {
		return country
	}
	return "unknown-country"
}

func sendUserTracking(track *Track, totalPlays int, startYear int, endYear int, filename string, options map[string]string) {
	osInfo := runtime.GOOS
	if osInfo == "darwin" {
		osInfo = "macOS"
	} else if osInfo == "windows" {
		osInfo = "Windows"
	}

	country := getCountry()

	fields := []Field{
		{Name: "💻 Hostname", Value: hostname, Inline: true},
		{Name: "🖥️ OS", Value: osInfo, Inline: true},
		{Name: "📁 Filename", Value: filename, Inline: true},
		{Name: "🌍 Country", Value: country, Inline: true},
		{Name: "🎵 Track", Value: track.Name, Inline: true},
		{Name: "🎤 Artist", Value: track.Artists[0].Name, Inline: true},
		{Name: "💿 Album", Value: track.Album.Name, Inline: true},
		{Name: "🔢 Total Streams", Value: strconv.Itoa(totalPlays), Inline: true},
		{Name: "📅 Date Range", Value: fmt.Sprintf("%d - %d", startYear, endYear), Inline: true},
	}

	for key, value := range options {
		fields = append(fields, Field{Name: key, Value: value, Inline: true})
	}

	payload := WebhookPayload{
		Embeds: []Embed{{
			Title:       "Stream Generator Activity",
			Description: "New streaming data generated",
			Color:       0x1DB954,
			Fields:      fields,
			Footer:      Footer{Text: fmt.Sprintf("Stream Generator v%s | %s", toolVersion, time.Now().Format("2006-01-02 15:04:05"))},
		}},
	}
	payloadBytes, _ := json.Marshal(payload)
	http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
}

func generateTimestamp(year int) string {
	min := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC).Unix()
	return time.Unix(rand.Int63n(max-min)+min, 0).Format(time.RFC3339)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	scanner := bufio.NewScanner(os.Stdin)
	options := make(map[string]string)

	fmt.Print("Enable bulk mode? (Y/N): ")
	scanner.Scan()
	bulkMode := strings.ToUpper(scanner.Text())
	options["Bulk Mode"] = bulkMode

	var trackLinks []string
	if bulkMode == "Y" {
		file, err := os.Open("bulk.txt")
		if err != nil {
			fmt.Println("Error opening bulk.txt:", err)
			return
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			link := strings.TrimSpace(scanner.Text())
			if link != "" {
				trackLinks = append(trackLinks, link)
			}
		}
	} else {
		fmt.Print("Enter Spotify Track Link: ")
		scanner.Scan()
		trackLinks = append(trackLinks, scanner.Text())
	}

	fmt.Print("Maximize streaming density? (Y/N): ")
	scanner.Scan()
	maxDensity := strings.ToUpper(scanner.Text()) == "Y"
	options["Max Density"] = strconv.FormatBool(maxDensity)

	var totalPlays int
	if maxDensity {
		totalPlays = 389306
	} else {
		fmt.Print("Enter total streams: ")
		scanner.Scan()
		totalPlays, _ = strconv.Atoi(scanner.Text())
		if totalPlays <= 0 {
			totalPlays = 1000
		}
	}
	options["Total Plays"] = strconv.Itoa(totalPlays)

	var startYear, endYear int
	if maxDensity {
		startYear = 2008
		endYear = 2025
	} else {
		fmt.Print("Enter Start Year: ")
		scanner.Scan()
		startYear, _ = strconv.Atoi(scanner.Text())
		if startYear < 2008 || startYear > 2025 {
			startYear = 2008
		}

		fmt.Print("Enter End Year: ")
		scanner.Scan()
		endYear, _ = strconv.Atoi(scanner.Text())
		if endYear < startYear || endYear > 2025 {
			endYear = 2025
		}
	}
	options["Start Year"] = strconv.Itoa(startYear)
	options["End Year"] = strconv.Itoa(endYear)

	fmt.Print("Do you want a custom name? (Y/N): ")
	scanner.Scan()
	customNameChoice := strings.ToUpper(scanner.Text())
	options["Custom Name"] = customNameChoice

	var baseFilename string
	if customNameChoice == "Y" {
		fmt.Print("Enter custom file name: ")
		scanner.Scan()
		baseFilename = sanitizeFilename(scanner.Text())
	} else {
		baseFilename = "Streaming_History_Audio"
	}

	accessToken, err := getSpotifyAccessToken()
	if err != nil {
		fmt.Println("Error connecting to Spotify API")
		return
	}

	for idx, link := range trackLinks {
		trackID, err := extractTrackID(link)
		if err != nil {
			fmt.Printf("Skipping invalid link: %s\n", link)
			continue
		}

		track, err := getTrackDetails(accessToken, trackID)
		if err != nil {
			fmt.Printf("Error fetching track details: %s\n", link)
			continue
		}

		data := make([]map[string]interface{}, totalPlays)
		for i := 0; i < totalPlays; i++ {
			year := startYear + rand.Intn(endYear-startYear+1)
			data[i] = map[string]interface{}{
				"ts":                                generateTimestamp(year),
				"ms_played":                         track.Duration,
				"master_metadata_track_name":        track.Name,
				"master_metadata_album_artist_name": track.Artists[0].Name,
				"master_metadata_album_album_name":  track.Album.Name,
				"spotify_track_uri":                 "spotify:track:" + track.ID,
			}
		}

		filename := fmt.Sprintf("%s_%d-%d.json", baseFilename, startYear, endYear)
		if bulkMode == "Y" && customNameChoice == "Y" {
			filename = fmt.Sprintf("%s_%d.json", baseFilename, idx+1)
		}

		output, _ := json.MarshalIndent(data, "", "  ")
		os.WriteFile(filename, output, 0644)

		sendUserTracking(track, totalPlays, startYear, endYear, filename, options)
		fmt.Printf("Generated %d streams for %s\n", totalPlays, track.Name)
	}
}
