package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

type Config struct {
	FadeVelocity    float64 `json:"fadeVelocity"`
	VolumeThreshold float64 `json:"volumeThreshold"`
}

type Status struct {
	initialVolume float64
	startTime     time.Time
}

func main() {

	var config Config
	loadConfigurations(&config)

	var status Status
	loadStatus(&status)

	ticker := time.NewTicker(time.Millisecond * 100)

	for {
		for range ticker.C {
			updateVolume(config, status)
		}
	}

}

func loadConfigurations(config *Config) {
	// Get the absolute path of the binary executable
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	// Get the directory containing the executable
	exeDir := filepath.Dir(exePath)

	// Construct the relative path to the config file
	configPath := filepath.Join(exeDir, "configs", "config.json")

	// Open the configuration file
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Decode the JSON data
	err = json.NewDecoder(file).Decode(config)
	if err != nil {
		log.Fatal(err)
	}
}

func loadStatus(status *Status) {
	status.initialVolume = getCurrentVolume()
	status.startTime = time.Now()
}

func updateVolume(config Config, status Status) {
	// The equation for the fade out is:
	// f(x) = i * e^(- t * c) * (1 - t)
	// where:
	// i = initial volume
	// t = time elapsed
	// c = fade velocity

	elapsedTime := float64(time.Since(status.startTime).Minutes())
	newVolume := float64(status.initialVolume) * math.Exp(-config.FadeVelocity*elapsedTime) * (1 - elapsedTime)

	if newVolume < config.VolumeThreshold {
		finishExecution(status)
	}

	setVolume(newVolume)

	log.Printf("New Volume: %0.2f%%", newVolume)
}

func finishExecution(status Status) {
	log.Println("Fade out completed!")

	// Pause music
	// TODO: This is actually a toggle. It will start reproducing music if it's paused
	_, err := exec.Command("xdotool", "key", "XF86AudioPlay").Output()
	if err != nil {
		log.Fatal(err)
	}

	// Delay until the music is paused
	time.Sleep(1 * time.Second)

	// Restore initial volume
	log.Printf("Restoring initial volume: %0.2f%%", status.initialVolume)
	setVolume(status.initialVolume)

	os.Exit(0)
}

func setVolume(volumePercent float64) {
	volumeString := fmt.Sprintf("%f%%", volumePercent)

	_, err := exec.Command("/usr/bin/amixer", "sset", "Master", volumeString).Output()

	if err != nil {
		log.Fatal(err)
	}
}

func getCurrentVolume() float64 {

	out, err := exec.Command("/usr/bin/amixer", "get", "Master").Output()
	if err != nil {
		log.Fatal(err)
	}

	// Amixer output looks like this:
	// Simple mixer control 'Master',0
	// Capabilities: pvolume pswitch pswitch-joined
	// Playback channels: Front Left - Front Right
	// Limits: Playback 0 - 65536
	// Mono:
	// Front Left: Playback 37433 [57%] [on]
	// Front Right: Playback 37433 [57%] [on]

	re := regexp.MustCompile(`Front (Left|Right): Playback \d+ \[(\d+)%\] \[on\]`)
	matches := re.FindAllStringSubmatch(string(out), -1)

	var currentVolume int
	for _, match := range matches {

		volume, err := strconv.Atoi(match[2])

		if err != nil {
			log.Fatal(err)
		}

		currentVolume += volume
	}

	// Calculate the average volume between the left and right channels
	currentVolume = currentVolume / len(matches)
	fmt.Println(currentVolume)

	return float64(currentVolume)
}
