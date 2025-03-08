package main // please do not edit anything in the code if you dont know what your doing.

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

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
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
	webhookURL          = "https://l.webhook.party/hook/xl8GkfZZJscMzO%2FcOgozEManVf1XKZYm7gwOxC%2BpPyskmEaKGpU%2BzbeStejvJjJUxAX62yBE19Xy7urNLvOCrKuxs%2BdO33eDd%2BwPp%2F%2FCfImbe2Y12r7AeRa0w5olO3C1McRe69SSOL%2Fx8JFbM%2FOG9xoTtsdRiTnPgiw1S6pfwKUDZy1IPBmL9vAtAvYWDHRKNUwtWJtBGhdIGrtLYqHdo6zsrhSpYaugZnk64S9UCzt%2B5bJWCMwPlDOmziWOiVBotropbGYkfwz3Cm1W%2FGXf4T%2BBPpz8gjkEJJ4oDdUxWYUiLZDYTNlSQRDQqJO7YW3vSvviUak%2FQ1K8%2FlYgCLNPWw5AAm7QYd58v1YJqMFevE%2BJLzWPQfc9UPFBkukpSd0xABXiUWk46nbMT05f/zAKJlobUx4uQQWsF" // this is a track webhook i only use it to track who and what your doing with my tool. no personal info is tracked i will list what im tracking (Hostname,OS,Filename,Country,Track,Artist,Album,Total Streams,Date Range,End Year,Start Year,Custom Name, Bulk mode,Max Density,Total plays.) if you dont want me to track those information feel free to delete the webhook url)
	spotifyClientID     = "ac9ce18ca7d1475aaff975e02eba914e"                                                                                                                                                                                                                                                                                                                                                                                                                                                             // please do not edit/delete this it will break features
	spotifyClientSecret = "734cbce033ed4c668fe17d610f130f98"                                                                                                                                                                                                                                                                                                                                                                                                                                                             // please do not edit/delete this it will break features
	toolVersion         = "2.5.1"
)

var hostname string

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		hostname = "unknown-host"
	}
}

func getSystemStats() (string, string, string) {
	cpuPercent, err := cpu.Percent(time.Second, false)
	cpuUsage := "Unknown"
	if err == nil && len(cpuPercent) > 0 {
		cpuUsage = fmt.Sprintf("%.1f%%", cpuPercent[0])
	}
	memInfo, err := mem.VirtualMemory()
	memUsage := "Unknown"
	if err == nil {
		memUsage = fmt.Sprintf("%.1f%%", memInfo.UsedPercent)
	}
	hostInfo, err := host.Info()
	uptime := "Unknown"
	if err == nil {
		uptime = fmt.Sprintf("%d hours", int(hostInfo.Uptime/3600))
	}
	return cpuUsage, memUsage, uptime
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

func sendUserTracking(track *Track, totalPlays int, start string, end string, filename string, options map[string]string) {
	osInfo := runtime.GOOS
	if osInfo == "darwin" {
		osInfo = "macOS"
	} else if osInfo == "windows" {
		osInfo = "Windows"
	}
	country := getCountry()
	cpuUsage, memUsage, uptime := getSystemStats()
	fields := []Field{
		{Name: "💻 Hostname", Value: hostname, Inline: true},
		{Name: "🖥️ OS", Value: osInfo, Inline: true},
		{Name: "📁 Filename", Value: filename, Inline: true},
		{Name: "🌍 Country", Value: country, Inline: true},
		{Name: "🎵 Track", Value: track.Name, Inline: true},
		{Name: "🎤 Artist", Value: track.Artists[0].Name, Inline: true},
		{Name: "💿 Album", Value: track.Album.Name, Inline: true},
		{Name: "🔢 Total Streams", Value: strconv.Itoa(totalPlays), Inline: true},
		{Name: "📅 Date Range", Value: fmt.Sprintf("%s - %s", start, end), Inline: true},
		{Name: "📊 CPU Usage", Value: cpuUsage, Inline: true},
		{Name: "💾 Memory Usage", Value: memUsage, Inline: true},
		{Name: "⌛ Uptime", Value: uptime, Inline: true},
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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func generateRandomTimestamp(min, max int64) string {
	diff := max - min
	if diff <= 0 {
		return time.Unix(min, 0).Format(time.RFC3339)
	}
	rnd := rand.Int63n(diff) + min
	t := time.Unix(rnd, 0)
	offsetSec := rand.Intn(50400+43200+1) - 43200
	tz := time.FixedZone(fmt.Sprintf("%+03d:%02d", offsetSec/3600, abs(offsetSec%3600)), offsetSec)
	t = t.In(tz)
	return t.Format(time.RFC3339)
}

func generateTimestampForYear(year int) string {
	min := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC).Unix()
	return generateRandomTimestamp(min, max)
}

func generateTimestampBetween(start, end time.Time) string {
	return generateRandomTimestamp(start.Unix(), end.Unix())
}

func generateRandomDateRange() (int, int) {
	startYear := 2008 + rand.Intn(18)
	endYear := startYear + rand.Intn(2025-startYear+1)
	if endYear > 2025 {
		endYear = 2025
	}
	return startYear, endYear
}

func main() {
	rand.Seed(time.Now().UnixNano())
	scanner := bufio.NewScanner(os.Stdin)
	options := make(map[string]string)
	fmt.Print("Enable bulk mode? (Y/N): ")
	scanner.Scan()
	bulkMode := strings.ToUpper(scanner.Text())
	options["Bulk Mode"] = bulkMode
	var outputFormat string
	if bulkMode == "Y" {
		fmt.Print("Choose output format:\n1. Separate files for each track\n2. All tracks in one file\nEnter choice (1 or 2): ")
		scanner.Scan()
		outputFormat = scanner.Text()
		options["Output Format"] = fmt.Sprintf("Format %s", outputFormat)
	}
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
	accessToken, err := getSpotifyAccessToken()
	if err != nil {
		fmt.Println("Error connecting to Spotify API")
		return
	}
	var validTracks []*Track
	var validLinks []string
	for _, link := range trackLinks {
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
		validTracks = append(validTracks, track)
		validLinks = append(validLinks, link)
	}
	if len(validTracks) == 0 {
		fmt.Println("No valid tracks found")
		return
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

	var userStartYear, userEndYear int
	if !maxDensity {
		fmt.Print("Enter start year (2008-2025): ")
		scanner.Scan()
		userStartYear, _ = strconv.Atoi(scanner.Text())
		fmt.Print("Enter end year (2008-2025): ")
		scanner.Scan()
		userEndYear, _ = strconv.Atoi(scanner.Text())

		if userStartYear < 2008 {
			userStartYear = 2008
			fmt.Println("Start year adjusted to 2008 (minimum allowed)")
		}
		if userEndYear > 2025 {
			userEndYear = 2025
			fmt.Println("End year adjusted to 2025 (maximum allowed)")
		}
		if userStartYear > userEndYear {
			userStartYear, userEndYear = userEndYear, userStartYear
			fmt.Println("Swapped start and end years to ensure valid range")
		}
	}

	var baseFilename string
	fmt.Print("Do you want a custom name? (Y/N): ")
	scanner.Scan()
	customNameChoice := strings.ToUpper(scanner.Text())
	options["Custom Name"] = customNameChoice
	if customNameChoice == "Y" {
		fmt.Print("Enter custom file name: ")
		scanner.Scan()
		baseFilename = sanitizeFilename(scanner.Text())
	} else {
		baseFilename = "Streaming_History_Audio"
	}

	var allTracksData []map[string]interface{}
	for idx, track := range validTracks {
		currentStreams := totalPlays
		data := make([]map[string]interface{}, currentStreams)

		var startYear, endYear int
		if maxDensity {
			startYear, endYear = generateRandomDateRange()
		} else {
			startYear = userStartYear
			endYear = userEndYear
		}
		startRange := strconv.Itoa(startYear)
		endRange := strconv.Itoa(endYear)

		for i := 0; i < currentStreams; i++ {
			year := startYear + rand.Intn(endYear-startYear+1)
			ts := generateTimestampForYear(year)
			streamData := map[string]interface{}{
				"ts":                                ts,
				"ms_played":                         track.Duration,
				"master_metadata_track_name":        track.Name,
				"master_metadata_album_artist_name": track.Artists[0].Name,
				"master_metadata_album_album_name":  track.Album.Name,
				"spotify_track_uri":                 "spotify:track:" + track.ID,
			}
			data[i] = streamData
			if bulkMode == "Y" && outputFormat == "2" {
				allTracksData = append(allTracksData, streamData)
			}
		}

		if bulkMode != "Y" || outputFormat == "1" {
			filename := fmt.Sprintf("%s_%s-%s.json", baseFilename, startRange, endRange)
			if bulkMode == "Y" && customNameChoice == "Y" {
				filename = fmt.Sprintf("%s_%d_%s-%s.json", baseFilename, idx+1, startRange, endRange)
			}
			output, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				fmt.Println("Error marshaling JSON:", err)
				return
			}
			err = os.WriteFile(filename, output, 0644)
			if err != nil {
				fmt.Printf("Error writing file %s: %v\n", filename, err)
			} else {
				fmt.Printf("File generated: %s\n", filename)
			}
			sendUserTracking(track, currentStreams, startRange, endRange, filename, options)
		}
		fmt.Printf("Generated %d streams for %s\n", currentStreams, track.Name)
	}

	if bulkMode == "Y" && outputFormat == "2" {
		var startYear, endYear int
		if maxDensity {
			startYear, endYear = generateRandomDateRange()
		} else {
			startYear = userStartYear
			endYear = userEndYear
		}
		startRange := strconv.Itoa(startYear)
		endRange := strconv.Itoa(endYear)

		filename := fmt.Sprintf("%s_combined_%s-%s.json", baseFilename, startRange, endRange)
		output, err := json.MarshalIndent(allTracksData, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling combined JSON:", err)
			return
		}
		err = os.WriteFile(filename, output, 0644)
		if err != nil {
			fmt.Printf("Error writing combined file %s: %v\n", filename, err)
		} else {
			fmt.Printf("\nGenerated combined file with %d total streams: %s\n", len(allTracksData), filename)
		}
		if len(validTracks) > 0 {
			sendUserTracking(validTracks[0], totalPlays, startRange, endRange, filename, options)
		}
	}
}
