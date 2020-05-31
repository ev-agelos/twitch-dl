package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
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

func cleanUrl(vodUrl string) string {
	without_spaces := strings.TrimSpace(vodUrl)
	// remove query string too
	videoUrl := strings.Split(without_spaces, "?")[0]
	return videoUrl
}

func fetchQualities(vodUrl string) []string {
	output, _ := exec.Command("streamlink", "-Q", vodUrl).Output()
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

func fetchVod(vodUrl string, player string, destination *string) {
	vodUrl = cleanUrl(vodUrl)
	qualities := fetchQualities(vodUrl)
	quality := readUserQuality(qualities)
	urlParts := strings.Split(vodUrl, "/")
	filename := path.Join(*destination, urlParts[len(urlParts)-1]+".mp4")
	res, err := exec.Command("streamlink", "-o", filename, "-p", player, vodUrl, quality).Output()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(res))
}

func main() {
	destination := flag.String("d", ".", "destination for the vod to be saved")
	flag.Parse()

	fileInfo, err := os.Stat(*destination)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	} else if !fileInfo.IsDir() {
		log.Fatal("Destination is not a directory")
		os.Exit(1)
	}

	if !isCommandInstalled("streamlink") {
		log.Fatal("streamlink is not installed.")
		os.Exit(1)
	}

	player, err := getPlayer()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	for _, vodUrl := range flag.Args() {
		_, err := url.ParseRequestURI(vodUrl)
		if err != nil {
			log.Fatal(err)
			continue
		}

		fetchVod(vodUrl, player, destination)
	}
}
