package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*apiConfig) error
}

type apiConfig struct {
	Next     *string
	Previous *string
}

var commands map[string]cliCommand

func main() {
	commands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Show next 20 location areas",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Show previous 20 location areas",
			callback:    commandMapBack,
		},
	}

	cfg := &apiConfig{}

	scanner := bufio.NewScanner(os.Stdin)

	// REPL loop
	for {
		fmt.Print("Pokedex > ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Fprintln(os.Stderr, "error reading input:", err)
			}
			break
		}

		words := cleanInput(scanner.Text())
		if len(words) == 0 {
			continue
		}

		cmdName := words[0]
		if cmd, ok := commands[cmdName]; ok {
			if err := cmd.callback(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
		} else {
			fmt.Println("Unknown command")
		}
	}
}

func commandExit(cfg *apiConfig) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *apiConfig) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()

	names := make([]string, 0, len(commands))
	for k := range commands {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, n := range names {
		cmd := commands[n]
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

// locationAreaList models the PokeAPI response for location-area list
type locationAreaList struct {
	Count    int `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

func fetchLocationAreas(url string) (*locationAreaList, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var list locationAreaList
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

func commandMap(cfg *apiConfig) error {
	base := "https://pokeapi.co/api/v2/location-area/"
	url := base
	if cfg.Next != nil && *cfg.Next != "" {
		url = *cfg.Next
	}

	list, err := fetchLocationAreas(url)
	if err != nil {
		return err
	}

	for _, r := range list.Results {
		fmt.Println(r.Name)
	}

	cfg.Next = list.Next
	cfg.Previous = list.Previous
	return nil
}

func commandMapBack(cfg *apiConfig) error {
	if cfg.Previous == nil || *cfg.Previous == "" {
		fmt.Println("you're on the first page")
		return nil
	}

	list, err := fetchLocationAreas(*cfg.Previous)
	if err != nil {
		return err
	}

	for _, r := range list.Results {
		fmt.Println(r.Name)
	}

	cfg.Next = list.Next
	cfg.Previous = list.Previous
	return nil
}

// trims, lowercases and splits the input into words.
func cleanInput(text string) []string {
	s := strings.TrimSpace(text)
	if s == "" {
		return []string{}
	}

	s = strings.ToLower(s)

	parts := strings.Fields(s)
	if len(parts) == 0 {
		return []string{}
	}
	return parts
}