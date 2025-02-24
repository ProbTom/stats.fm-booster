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
	Duration int
}

const (
	defaultTrack = "Cry For Me"
)

func extractTrackID(input string) string {
	if strings.Contains(input, "open.spotify.com") {
		trackIndex := strings.Index(input, "track/")
		if trackIndex == -1 {
			return ""
		}
		idStart := trackIndex + len("track/")
		idEnd := strings.Index(input[idStart:], "?")
		if idEnd == -1 {
			return input[idStart:]
		}
		return input[idStart : idStart+idEnd]
	}

	if strings.Contains(input, "/") {
		parts := strings.Split(input, "/")
		if len(parts) == 2 {
			return parts[1]
		}
	}

	return input
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
	fmt.Print("Enter Track ID or Link: ")
	scanner.Scan()
	input := scanner.Text()

	trackID := extractTrackID(input)
	if trackID == "" {
		fmt.Println("Invalid input. Please use a valid Spotify track link or ID.")
		return
	}

	selectedTrack := Track{
		ID:       trackID,
		Name:     defaultTrack,
		Duration: 200000,
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
			"spotify_track_uri":                 "spotify:track:" + selectedTrack.ID,
		}

		dataList = append(dataList, streamData)
		currentTS = updatedTS
	}

	outputFile, err := json.MarshalIndent(dataList, "", "    ")
	if err != nil {
		panic(err)
	}

	fmt.Print("Do you want to customize the file name? (Y/N): ")
	scanner.Scan()
	customNameChoice := strings.ToUpper(scanner.Text())
	if customNameChoice != "Y" {
		customNameChoice = "N"
	}

	filename := "output.json"
	if customNameChoice == "Y" {
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
