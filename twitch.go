package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func isCommandInstalled(command string) bool {
	cmd := exec.Command("/bin/sh", "-c", "command", "-v", command)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

func getPlayer() (string, error) {
	if isCommandInstalled("mpv") {
		return "mpv", nil
	}
	if isCommandInstalled("vlc") {
		return "vlc", nil
	}
	return "", fmt.Errorf("Mpv player is not installed, nor vlc.")
}

func cleanUrl(url string) string {
	without_spaces := strings.TrimSpace(url)
	// remove query string too
	video_url := strings.Split(without_spaces, "?")[0]
	return video_url
}

func fetchQualities(url string) []string {
	output, _ := exec.Command("streamlink", "-Q", url).Output()
	qualities := string(output)
	qualities = strings.Replace(qualities, " (worst)", "", -1)
	qualities = strings.Replace(qualities, " (best)", "", -1)
	return strings.Split(qualities, ", ")[1:]
}

func readUserQuality(qualities []string) string {
	fmt.Println("Select quality: ")
	for index, quality := range qualities {
		fmt.Println(index+1, quality)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		input = strings.TrimSpace(input)
		selection, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Input was not an integer")
			continue
		}
		if selection < 1 || selection > len(qualities) {
			fmt.Println("Input was not within the available range")
			continue
		}
		return qualities[selection-1]
	}
}

func main() {
	if !isCommandInstalled("streamlink") {
		log.Fatal("streamlink is not installed.")
		os.Exit(1)
	}

	player, err := getPlayer()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	for _, url := range os.Args[1:] {
		videoUrl := cleanUrl(url)
		qualities := fetchQualities(videoUrl)
		quality := readUserQuality(qualities)
		urlParts := strings.Split(videoUrl, "/")
		filename := urlParts[len(urlParts)-1] + ".mp4"
		res, err := exec.Command("streamlink", "-o", filename, "-p", player, videoUrl, quality).Output()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(res))
	}
}
