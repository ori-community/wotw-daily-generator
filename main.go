package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type GoalModeEntry struct {
	Name        string
	Probability int
}

type HeaderIncludeEntry struct {
	HeaderNames []string
	Probability int
}

type ProbabilitySettings struct {
	GoalMode       []GoalModeEntry
	UseRandomSpawn int
	HeaderGroups   [][]HeaderIncludeEntry
}

type Settings struct {
	EnableKii       bool
	Probabilities   ProbabilitySettings
	SeedgenPath     string
	DisableSpoilers bool
}

func chance(percentage int) bool {
	return rand.Intn(100) < percentage
}

func IfElse(condition bool, a interface{}, b interface{}) interface{} {
	if condition {
		return a
	}
	return b
}

func main() {
	var date = time.Now().Format("2006-01-02")
	rand.Seed(time.Now().UnixNano())

	var settingsFile, _ = os.ReadFile("settings.json")

	var settings Settings
	err := json.Unmarshal(settingsFile, &settings)

	if err != nil {
		log.Fatal(err)
	}

	// Difficulty
	var difficulty = "gorlek"
	if settings.EnableKii && chance(40) {
		difficulty = "kii"
	}

	// Spawn
	var randomSpawn = chance(settings.Probabilities.UseRandomSpawn)

	// Goal mode
	var goalMode = "trees"
	var goalModeRng = rand.Intn(100)
	var lastTotalProbability = 0
	for _, goalModeEntry := range settings.Probabilities.GoalMode {
		if goalModeRng > lastTotalProbability && goalModeRng <= (lastTotalProbability+goalModeEntry.Probability) {
			goalMode = goalModeEntry.Name
			break
		}

		lastTotalProbability += goalModeEntry.Probability
	}

	// Headers
	var headers []string
	for _, headerGroup := range settings.Probabilities.HeaderGroups {
		for _, headerEntry := range headerGroup {
			if chance(headerEntry.Probability) {
				for _, headerName := range headerEntry.HeaderNames {
					headers = append(headers, headerName)
				}
				break
			}
		}
	}

	// Generate command
	var command = "seed --verbose" +
		" --difficulty " + difficulty +
		" --preset qol" +
		" --goals " + goalMode +
		" --headers " + strings.Join(headers, " ")

	if randomSpawn {
		command += " --spawn random"
	}

	if settings.DisableSpoilers {
		command += " -r"
	}

	command += " -- " + date

	println(command)

	var seedgen = exec.Command(settings.SeedgenPath, strings.Split(command, " ")...)

	seedgen.Dir = filepath.Dir(settings.SeedgenPath)
	seedgen.Stdout = os.Stdout
	seedgen.Stderr = os.Stderr
	err = seedgen.Run()

	if err != nil {
		log.Fatal(err)
	}

	println()
	println()

	println("**Daily seed " + date + "**")
	println("Goal mode: *" + goalMode + "*")
	println("Difficulty: *" + difficulty + "*")
	println("Random spawn: *" + IfElse(randomSpawn, "yes", "no").(string) + "*")
	println("Headers: *" + strings.Join(headers, ", ") + "*")

	println()
	println()

	fmt.Print("Press [Enter] to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
