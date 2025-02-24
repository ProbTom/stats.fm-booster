package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Track struct {
	ID       string
	Name     string
	Artist   string
	Album    string
	Duration int
}

const (
	defaultArtist = "The Weeknd"
	defaultAlbum  = "Hurry Up Tomorrow"
	defaultTrack  = "Cry For Me"
)

var (
	albumTracks = map[string][]Track{
		"3OxfaVgvTxUTy7276t7SPU": {
			{ID: "3AWDeHLc88XogCaCnZQLVI", Name: "Cry For Me", Artist: "The Weeknd", Album: "Hurry Up Tomorrow", Duration: 200000},
			{ID: "4sWQbsLLH2NEbO79DSZCL9", Name: "Big Sleep", Artist: "The Weeknd", Album: "Hurry Up Tomorrow", Duration: 228000},
		},
	}

	artistTracks = map[string][]Track{
		"1Xyo4u8uXC1ZmMpatF05PJ": {},
	}
)

func extractID(input string) (string, string) {
	if strings.Contains(input, "open.spotify.com") {
		trackIndex := strings.Index(input, "track/")
		albumIndex := strings.Index(input, "album/")
		artistIndex := strings.Index(input, "artist/")

		var typePrefix string
		var idStart int
		switch {
		case trackIndex != -1:
			typePrefix = "track"
			idStart = trackIndex + len("track/")
		case albumIndex != -1:
			typePrefix = "album"
			idStart = albumIndex + len("album/")
		case artistIndex != -1:
			typePrefix = "artist"
			idStart = artistIndex + len("artist/")
		default:
			return "", ""
		}

		idEnd := strings.Index(input[idStart:], "?")
		if idEnd == -1 {
			idEnd = len(input)
		} else {
			idEnd += idStart
		}
		id := input[idStart:idEnd]

		return typePrefix, id
	}

	if strings.Contains(input, "/") {
		parts := strings.Split(input, "/")
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}

	return "track", input
}

func addMilliseconds(ts string, msPlayed int) string {
	parsedTime, err := time.Parse("2006-01-02T15:04:05Z", ts)
	if err != nil {
		panic(err)
	}
	updatedTime := parsedTime.Add(time.Millisecond * time.Duration(msPlayed))
	return updatedTime.Format("2006-01-02T15:04:05Z")
}

func getRandomYear(minYear, maxYear int) int {
	return rand.Intn(maxYear-minYear+1) + minYear
}

func getRandomTrackFromAlbum(albumID string) Track {
	tracks, exists := albumTracks[albumID]
	if !exists || len(tracks) == 0 {
		return Track{
			ID:       "defaultTrackID",
			Name:     defaultTrack,
			Artist:   defaultArtist,
			Album:    defaultAlbum,
			Duration: 200000,
		}
	}
	return tracks[rand.Intn(len(tracks))]
}

func getRandomTrackFromArtist(artistID string) Track {
	tracks, exists := artistTracks[artistID]
	if !exists || len(tracks) == 0 {
		return Track{
			ID:       "defaultTrackID",
			Name:     defaultTrack,
			Artist:   defaultArtist,
			Album:    defaultAlbum,
			Duration: 200000,
		}
	}
	return tracks[rand.Intn(len(tracks))]
}

func sanitizeFilename(name string) string {
	if dotIndex := strings.Index(name, "."); dotIndex != -1 {
		name = name[:dotIndex] // Remove everything after the dot
	}
	name = strings.ReplaceAll(name, " ", "_")
	return name
}

func main() {
	rand.Seed(time.Now().UnixNano())

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter Track/Album/Artist: ")
	scanner.Scan()
	input := scanner.Text()

	inputType, id := extractID(input)
	if inputType == "" || id == "" {
		fmt.Println("Invalid input. Please use a valid Spotify link or ID.")
		return
	}

	var selectedTrack Track
	switch inputType {
	case "track":
		selectedTrack = Track{
			ID:       id,
			Name:     defaultTrack,
			Artist:   defaultArtist,
			Album:    defaultAlbum,
			Duration: 200000,
		}
	case "album":
		selectedTrack = getRandomTrackFromAlbum(id)
	case "artist":
		selectedTrack = getRandomTrackFromArtist(id)
	default:
		fmt.Println("Invalid input type. Please use track/, album/, or artist/.")
		return
	}

	fmt.Print("Enter total number of streams: ")
	scanner.Scan()
	totalPlays, err := strconv.Atoi(scanner.Text())
	if err != nil || totalPlays <= 0 {
		fmt.Println("Invalid input for total streams. Using default value of 10.")
		totalPlays = 10
	}

	fmt.Print("Enter start date (e.g., 2015): ")
	scanner.Scan()
	startYear, err := strconv.Atoi(scanner.Text())
	if err != nil || startYear < 2000 || startYear > time.Now().Year() {
		fmt.Println("Invalid start date. Using default value of 2015.")
		startYear = 2015
	}

	fmt.Print("Enter end date (e.g., 2025): ")
	scanner.Scan()
	endYear, err := strconv.Atoi(scanner.Text())
	if err != nil || endYear < startYear || endYear > time.Now().Year()+10 {
		fmt.Println("Invalid end date. Using default value of 2025.")
		endYear = 2025
	}

	dataList := make([]map[string]interface{}, 0)
	currentTS := time.Date(getRandomYear(startYear, endYear), time.Month(rand.Intn(12)+1), rand.Intn(28)+1, rand.Intn(24), rand.Intn(60), rand.Intn(60), 0, time.UTC).Format("2006-01-02T15:04:05Z")

	for i := 0; i < totalPlays; i++ {
		msPlayed := rand.Intn(selectedTrack.Duration)
		updatedTS := addMilliseconds(currentTS, msPlayed)

		streamData := map[string]interface{}{
			"ts":                                currentTS,
			"ms_played":                         msPlayed,
			"master_metadata_track_name":        selectedTrack.Name,
			"master_metadata_album_artist_name": selectedTrack.Artist,
			"master_metadata_album_album_name":  selectedTrack.Album,
			"spotify_track_uri":                 "spotify:track:" + selectedTrack.ID,
		}

		dataList = append(dataList, streamData)
		currentTS = updatedTS
	}

	outputFile, err := json.MarshalIndent(dataList, "", "    ")
	if err != nil {
		panic(err)
	}

	fmt.Print("Do you want to customize the file name? (1 Yes - 0 No): ")
	scanner.Scan()
	customNameChoice, err := strconv.Atoi(scanner.Text())
	if err != nil || customNameChoice != 1 {
		customNameChoice = 0
	}

	filename := "output.json"
	if customNameChoice == 1 {
		fmt.Print("Enter desired file name: ")
		scanner.Scan()
		customName := scanner.Text()
		customName = sanitizeFilename(customName)
		filename = customName + ".json"
	}

	err = writeToFile(filename, outputFile)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Data written to %s\n", filename)
}

func writeToFile(filename string, data []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	err = file.Sync()
	if err != nil {
		return err
	}

	return nil
}