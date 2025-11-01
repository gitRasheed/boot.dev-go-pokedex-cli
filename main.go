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
	"time"

	"github.com/gitRasheed/boot.dev-go-pokedex-cli/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*apiConfig, []string) error
}

type apiConfig struct {
	Next     *string
	Previous *string
}

var commands map[string]cliCommand
var cache *pokecache.Cache

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
		"explore": {
			name:        "explore",
			description: "Explore a location area (explore <area_name>)",
			callback:    commandExplore,
		},
	}

	cfg := &apiConfig{}

	cache = pokecache.NewCache(5 * time.Second)

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
		args := []string{}
		if len(words) > 1 {
			args = words[1:]
		}

		if cmd, ok := commands[cmdName]; ok {
			if err := cmd.callback(cfg, args); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
		} else {
			fmt.Println("Unknown command")
		}
	}
}

func commandExit(cfg *apiConfig, _ []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *apiConfig, _ []string) error {
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

// locationAreaDetail models the detailed location-area response with Pokemon encounters
type locationAreaDetail struct {
	Name              string `json:"name"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

func fetchLocationAreaDetail(url string) (*locationAreaDetail, error) {
	// Try cache first
	if cache != nil {
		if b, ok := cache.Get(url); ok {
			var detail locationAreaDetail
			if err := json.Unmarshal(b, &detail); err == nil {
				return &detail, nil
			}
			// fallthrough to fetch if unmarshal fails
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var detail locationAreaDetail
	if err := json.Unmarshal(body, &detail); err != nil {
		return nil, err
	}

	if cache != nil {
		cache.Add(url, body)
	}

	return &detail, nil
}

func fetchLocationAreas(url string) (*locationAreaList, error) {
	// Try cache first
	if cache != nil {
		if b, ok := cache.Get(url); ok {
			var list locationAreaList
			if err := json.Unmarshal(b, &list); err == nil {
				return &list, nil
			}
			// fallthrough to fetch if unmarshal fails
		}
	}

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

	if cache != nil {
		cache.Add(url, body)
	}

	return &list, nil
}

func commandMap(cfg *apiConfig, _ []string) error {
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

func commandMapBack(cfg *apiConfig, _ []string) error {
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

func commandExplore(cfg *apiConfig, args []string) error {
	if len(args) == 0 {
		fmt.Println("usage: explore <location-area-name>")
		return nil
	}
	name := args[0]
	fmt.Printf("Exploring %s...\n", name)

	base := "https://pokeapi.co/api/v2/location-area/"
	url := base + name + "/"

	detail, err := fetchLocationAreaDetail(url)
	if err != nil {
		return err
	}

	fmt.Println("Found Pokemon:")
	for _, pe := range detail.PokemonEncounters {
		fmt.Printf(" - %s\n", pe.Pokemon.Name)
	}
	return nil
}