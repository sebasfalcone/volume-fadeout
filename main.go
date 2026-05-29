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

// Load the configuration from the config file
func loadConfigurations(config *Config) {
	currentDir, err := os.Getwd()
	log.Printf("Current directory: %s", currentDir)
	if err != nil {
		log.Fatal(err)
	}

	configPath := filepath.Join(currentDir, "configs", "config.json")

	file, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(config)
	if err != nil {
		log.Fatal(err)
	}
}

// Get the initial volume and start time
func loadStatus(status *Status) {
	status.initialVolume = getAverageVolume()
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

// Finish the execution of the program, restoring the initial volume and pausing the music
func finishExecution(status Status) {
	log.Println("Fade out completed!")

	_, err := exec.Command("playerctl", "pause").Output()
	if err != nil {
		log.Panic("Unable to pause music: " + err.Error())
	}

	// Delay until the music is paused
	time.Sleep(1 * time.Second)

	// Restore initial volume
	setVolume(status.initialVolume)

	os.Exit(0)
}

func setVolume(volumePercent float64) {
	volumeString := fmt.Sprintf("%f%%", volumePercent)
	log.Printf("Setting volume to: %s", volumeString)

	_, err := exec.Command("/usr/bin/amixer", "-D", "pulse", "set", "Master", volumeString).Output()

	if err != nil {
		log.Panic("Unable to set volume: " + err.Error())
	}
}

// Get the average volume of the left and right channels
func getAverageVolume() float64 {

	out, err := exec.Command("/usr/bin/amixer", "get", "Master").Output()
	if err != nil {
		log.Panic("Unable to retrieve volume: " + err.Error())
	}

	// Amixer output looks like this:
	// Simple mixer control 'Master',0
  	// 	Capabilities: pvolume pswitch pswitch-joined
  	// 	Playback channels: Front Left - Front Right
  	// 	Limits: Playback 0 - 65536
  	// 	Mono:
  	// 	Front Left: Playback 39312 [60%] [on]
  	// 	Front Right: Playback 39312 [60%] [on]

	left := regexp.MustCompile(`Front Left: Playback \d+ \[(\d+)%\]`)
	matchLeft := left.FindStringSubmatch(string(out))
	if matchLeft == nil {
		log.Panic("Unable to find left channel volume")
	}

	right := regexp.MustCompile(`Front Right: Playback \d+ \[(\d+)%\]`)
	matchRight := right.FindStringSubmatch(string(out))
	if matchRight == nil {
		log.Panic("Unable to find right channel volume")
	}

	leftVolume, err := strconv.Atoi(matchLeft[1])
	log.Printf("Left channel volume: %d", leftVolume)
	if err != nil {
		log.Panic("Unable to parse left channel volume")
	}

	rightVolume, err := strconv.Atoi(matchRight[1])
	log.Printf("Right channel volume: %d", rightVolume)
	if err != nil {
		log.Panic("Unable to parse right channel volume")
	}

	currentVolume := (leftVolume + rightVolume) / 2

	log.Printf("Current volume: %d", currentVolume)
	return float64(currentVolume)
}
