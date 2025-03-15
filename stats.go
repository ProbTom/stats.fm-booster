package main // please do not edit anything in the code if you dont know what your doing.

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
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
		Name        string `json:"name"`
		ReleaseDate string `json:"release_date"`
		ReleaseYear int    `json:"-"`
	} `json:"album"`
}

type PlaylistTracks struct {
	Items []struct {
		Track Track `json:"track"`
	} `json:"items"`
	Next string `json:"next"`
}

type AlbumTracks struct {
	Items []Track `json:"items"`
	Next  string  `json:"next"`
}

type ArtistAlbums struct {
	Items []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"items"`
	Next string `json:"next"`
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
	toolVersion         = "2.6.0"
)

var (
	hostname   string
	apiCache   = make(map[string][]byte)
	cacheMutex = &sync.Mutex{}
	retryDelay = 5 * time.Second // Initial retry delay
	maxRetries = 3               // Maximum number of retries
)

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		hostname = "unknown-host"
	}
	rand.Seed(time.Now().UnixNano())
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func getSystemStats() (string, string, string) {
	cpuPercent, _ := cpu.Percent(time.Second, false)
	memInfo, _ := mem.VirtualMemory()
	hostInfo, _ := host.Info()
	return fmt.Sprintf("%.1f%%", cpuPercent[0]),
		fmt.Sprintf("%.1f%%", memInfo.UsedPercent),
		fmt.Sprintf("%d hours", int(hostInfo.Uptime/3600))
}

func extractID(input, entity string) string {
	re := regexp.MustCompile(fmt.Sprintf(`/%s/([a-zA-Z0-9]{22})`, entity))
	matches := re.FindStringSubmatch(input)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

func getSpotifyAccessToken() (string, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://accounts.spotify.com/api/token",
		strings.NewReader("grant_type=client_credentials"))
	req.SetBasicAuth(spotifyClientID, spotifyClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get access token: %s, body: %s", resp.Status, string(bodyBytes))
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.AccessToken, nil
}

func makeSpotifyAPIRequest(req *http.Request) ([]byte, error) {
	cacheKey := req.URL.String()
	cacheMutex.Lock()
	if cachedData, exists := apiCache[cacheKey]; exists {
		cacheMutex.Unlock()
		return cachedData, nil
	}
	cacheMutex.Unlock()

	client := &http.Client{}
	var resp *http.Response
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := resp.Header.Get("Retry-After")
			if retryAfter == "" {
				retryAfter = "5"
			}
			waitTime, _ := strconv.Atoi(retryAfter)
			fmt.Printf("Rate limited. Retrying after %d seconds...\n", waitTime)
			time.Sleep(time.Duration(waitTime) * time.Second)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("spotify API error: %s, body: %s", resp.Status, string(bodyBytes))
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		cacheMutex.Lock()
		apiCache[cacheKey] = bodyBytes
		cacheMutex.Unlock()

		return bodyBytes, nil
	}

	return nil, fmt.Errorf("max retries exceeded for request: %s", req.URL.String())
}

func getTrack(accessToken, trackID string) (*Track, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.spotify.com/v1/tracks/%s", trackID), nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	bodyBytes, err := makeSpotifyAPIRequest(req)
	if err != nil {
		return nil, err
	}

	var track Track
	if err := json.Unmarshal(bodyBytes, &track); err != nil {
		return nil, err
	}

	if len(track.Album.ReleaseDate) >= 4 {
		year, _ := strconv.Atoi(track.Album.ReleaseDate[:4])
		track.Album.ReleaseYear = year
	} else {
		track.Album.ReleaseYear = 2008
	}

	return &track, nil
}

func getAlbumTracks(accessToken, albumID string) ([]Track, error) {
	albumReq, _ := http.NewRequest("GET", fmt.Sprintf("https://api.spotify.com/v1/albums/%s", albumID), nil)
	albumReq.Header.Set("Authorization", "Bearer "+accessToken)

	bodyBytes, err := makeSpotifyAPIRequest(albumReq)
	if err != nil {
		return nil, err
	}

	var albumDetails struct {
		Name        string `json:"name"`
		ReleaseDate string `json:"release_date"`
	}
	if err := json.Unmarshal(bodyBytes, &albumDetails); err != nil {
		return nil, err
	}

	var tracks []Track
	url := fmt.Sprintf("https://api.spotify.com/v1/albums/%s/tracks?limit=50", albumID)
	for url != "" {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		bodyBytes, err := makeSpotifyAPIRequest(req)
		if err != nil {
			return nil, err
		}

		var result AlbumTracks
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return nil, err
		}

		for i := range result.Items {
			result.Items[i].Album.Name = albumDetails.Name
			result.Items[i].Album.ReleaseDate = albumDetails.ReleaseDate
			if len(albumDetails.ReleaseDate) >= 4 {
				year, _ := strconv.Atoi(albumDetails.ReleaseDate[:4])
				result.Items[i].Album.ReleaseYear = year
			}
		}

		tracks = append(tracks, result.Items...)
		url = result.Next
	}
	return tracks, nil
}

func getArtistAlbums(accessToken, artistID string) ([]string, error) {
	var albumIDs []string
	url := fmt.Sprintf("https://api.spotify.com/v1/artists/%s/albums?limit=50&include_groups=album,single", artistID)

	for url != "" {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		bodyBytes, err := makeSpotifyAPIRequest(req)
		if err != nil {
			return nil, err
		}

		var result ArtistAlbums
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return nil, err
		}

		for _, album := range result.Items {
			albumIDs = append(albumIDs, album.ID)
		}
		url = result.Next
	}
	return albumIDs, nil
}

func processArtist(accessToken, artistID string, maxTracks int) ([]Track, error) {
	albumIDs, err := getArtistAlbums(accessToken, artistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get albums: %v", err)
	}

	var allTracks []Track
	for _, albumID := range albumIDs {
		tracks, err := getAlbumTracks(accessToken, albumID)
		if err != nil {
			continue
		}
		allTracks = append(allTracks, tracks...)
		if len(allTracks) >= maxTracks {
			break
		}
	}

	if len(allTracks) > maxTracks {
		allTracks = allTracks[:maxTracks]
	}
	return allTracks, nil
}

func processPlaylist(accessToken, playlistID string) ([]Track, error) {
	var tracks []Track
	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?limit=50", playlistID)

	for url != "" {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		bodyBytes, err := makeSpotifyAPIRequest(req)
		if err != nil {
			return nil, err
		}

		var result PlaylistTracks
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return nil, err
		}

		for _, item := range result.Items {
			track, err := getTrack(accessToken, item.Track.ID)
			if err != nil {
				continue
			}
			tracks = append(tracks, *track)
		}
		url = result.Next
	}
	return tracks, nil
}

func processLink(accessToken, link string, maxTracks int) ([]Track, error) {
	switch {
	case strings.Contains(link, "/track/"):
		trackID := extractID(link, "track")
		if trackID == "" {
			return nil, fmt.Errorf("invalid track ID")
		}
		track, err := getTrack(accessToken, trackID)
		if err != nil {
			return nil, err
		}
		return []Track{*track}, nil

	case strings.Contains(link, "/playlist/"):
		playlistID := extractID(link, "playlist")
		if playlistID == "" {
			return nil, fmt.Errorf("invalid playlist ID")
		}
		return processPlaylist(accessToken, playlistID)

	case strings.Contains(link, "/album/"):
		albumID := extractID(link, "album")
		if albumID == "" {
			return nil, fmt.Errorf("invalid album ID")
		}
		return getAlbumTracks(accessToken, albumID)

	case strings.Contains(link, "/artist/"):
		artistID := extractID(link, "artist")
		if artistID == "" {
			return nil, fmt.Errorf("invalid artist ID")
		}
		return processArtist(accessToken, artistID, maxTracks)

	default:
		return nil, fmt.Errorf("unsupported link type")
	}
}

func sanitizeFilename(name string) string {
	return strings.Map(func(r rune) rune {
		if strings.ContainsRune("<>:\"/\\|?*", r) {
			return -1
		}
		return r
	}, name)
}

func getCountry() string {
	resp, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		return "unknown-country"
	}
	defer resp.Body.Close()

	var result struct {
		Country string `json:"country"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Country
}

func sendUserTracking(track Track, totalPlays int, start string, end string, filename string, options map[string]string) {
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

func generateRandomDateRange() (int, int) {
	startYear := 2008 + rand.Intn(18)
	endYear := startYear + rand.Intn(2025-startYear+1)
	if endYear > 2025 {
		endYear = 2025
	}
	return startYear, endYear
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	options := make(map[string]string)

	fmt.Print("Enable bulk mode? (Y/N): ")
	scanner.Scan()
	bulkMode := strings.ToUpper(scanner.Text())

	var links []string
	if bulkMode == "Y" {
		file, _ := os.Open("bulk.txt")
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			links = append(links, strings.TrimSpace(scanner.Text()))
		}
	} else {
		fmt.Print("Enter Spotify Track/Album/Playlist/Artist Link: ")
		scanner.Scan()
		links = append(links, scanner.Text())
	}

	accessToken, err := getSpotifyAccessToken()
	if err != nil {
		fmt.Println("Error: Failed to connect to Spotify API:", err)
		return
	}

	var maxTracks int
	if strings.Contains(links[0], "/artist/") {
		fmt.Print("Enter the number of songs to generate (e.g., 50): ")
		scanner.Scan()
		maxTracks, _ = strconv.Atoi(scanner.Text())
		if maxTracks <= 0 {
			maxTracks = 50
		}
	} else {
		maxTracks = 1000
	}

	var allTracks []Track
	for _, link := range links {
		tracks, err := processLink(accessToken, link, maxTracks)
		if err != nil {
			fmt.Printf("Skipping invalid link: %s (%v)\n", link, err)
			continue
		}
		allTracks = append(allTracks, tracks...)
	}

	if len(allTracks) == 0 {
		fmt.Println("Error: No valid tracks found in any provided links")
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
	fmt.Print("Do you want custom dates? (Y/N): ")
	scanner.Scan()
	customDates := strings.ToUpper(scanner.Text()) == "Y"
	if customDates {
		fmt.Print("Enter start year (2008-2025): ")
		scanner.Scan()
		userStartYear, _ = strconv.Atoi(scanner.Text())
		fmt.Print("Enter end year (2008-2025): ")
		scanner.Scan()
		userEndYear, _ = strconv.Atoi(scanner.Text())

		if userStartYear < 2008 {
			userStartYear = 2008
		}
		if userEndYear > 2025 {
			userEndYear = 2025
		}
		if userStartYear > userEndYear {
			userStartYear, userEndYear = userEndYear, userStartYear
		}
	} else {
		switch {
		case strings.Contains(links[0], "/album/"):
			albumID := extractID(links[0], "album")
			albumReq, _ := http.NewRequest("GET", fmt.Sprintf("https://api.spotify.com/v1/albums/%s", albumID), nil)
			albumReq.Header.Set("Authorization", "Bearer "+accessToken)
			bodyBytes, err := makeSpotifyAPIRequest(albumReq)
			if err != nil {
				fmt.Println("Error fetching album details:", err)
				return
			}

			var albumDetails struct {
				ReleaseDate string `json:"release_date"`
			}
			if err := json.Unmarshal(bodyBytes, &albumDetails); err != nil {
				fmt.Println("Error decoding album details:", err)
				return
			}
			if len(albumDetails.ReleaseDate) >= 4 {
				userStartYear, _ = strconv.Atoi(albumDetails.ReleaseDate[:4])
			} else {
				userStartYear = 2008
			}

		case strings.Contains(links[0], "/artist/"):
			tracks, err := processArtist(accessToken, extractID(links[0], "artist"), 1)
			if err != nil {
				fmt.Println("Error fetching artist tracks:", err)
				return
			}
			if len(tracks) > 0 {
				userStartYear = tracks[0].Album.ReleaseYear
			} else {
				userStartYear = 2008
			}

		case strings.Contains(links[0], "/playlist/"):
			earliestYear := 2025
			for _, track := range allTracks {
				if track.Album.ReleaseYear < earliestYear {
					earliestYear = track.Album.ReleaseYear
				}
			}
			if earliestYear == 2025 {
				earliestYear = 2008
			}
			userStartYear = earliestYear

		default:
			userStartYear = allTracks[0].Album.ReleaseYear
		}

		userEndYear = 2025
		fmt.Printf("Using release year %d as start year\n", userStartYear)
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

	var separateFiles bool
	if strings.Contains(links[0], "/artist/") || strings.Contains(links[0], "/playlist/") || strings.Contains(links[0], "/album/") {
		fmt.Print("Do you want separate files for each track? (Y/N): ")
		scanner.Scan()
		separateFiles = strings.ToUpper(scanner.Text()) == "Y"
	} else {
		separateFiles = false
	}

	var allTracksData []map[string]interface{}
	for _, track := range allTracks {
		currentStreams := totalPlays
		data := make([]map[string]interface{}, currentStreams)

		var startYear, endYear int
		if maxDensity {
			startYear, endYear = generateRandomDateRange()
		} else {
			startYear = userStartYear
			endYear = userEndYear
		}

		for i := 0; i < currentStreams; i++ {
			year := startYear + rand.Intn(endYear-startYear+1)
			if year < 2008 {
				year = 2008
			}
			if year > 2025 {
				year = 2025
			}
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
			if bulkMode == "Y" {
				allTracksData = append(allTracksData, streamData)
			}
		}

		if bulkMode != "Y" {
			if separateFiles {
				filename := fmt.Sprintf("%s_%s_%d-%d.json", baseFilename, sanitizeFilename(track.Name), startYear, endYear)
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
				sendUserTracking(track, currentStreams, fmt.Sprintf("%d", startYear), fmt.Sprintf("%d", endYear), filename, options)
			} else {
				allTracksData = append(allTracksData, data...)
			}
		}
		fmt.Printf("Generated %d streams for %s\n", currentStreams, track.Name)
	}

	if bulkMode == "Y" || !separateFiles {
		var startYear, endYear int
		if maxDensity {
			startYear, endYear = generateRandomDateRange()
		} else {
			startYear = userStartYear
			endYear = userEndYear
		}

		filename := fmt.Sprintf("%s_%d-%d.json", baseFilename, startYear, endYear)
		if !strings.Contains(links[0], "/track/") {
			filename = fmt.Sprintf("%s_combined_%d-%d.json", baseFilename, startYear, endYear)
		}

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
		if len(allTracks) > 0 {
			sendUserTracking(allTracks[0], totalPlays, fmt.Sprintf("%d", startYear), fmt.Sprintf("%d", endYear), filename, options)
		}
	}
}
