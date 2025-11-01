package api

import (
    "encoding/json"
    "io"
    "net/http"

    "github.com/gitRasheed/boot.dev-go-pokedex-cli/internal/pokecache"
)

var PokeAPIBase = "https://pokeapi.co/api/v2/"
var LocationAreaBase = PokeAPIBase + "location-area/"
var Cache *pokecache.Cache

type LocationAreaList struct {
    Count    int      `json:"count"`
    Next     *string  `json:"next"`
    Previous *string  `json:"previous"`
    Results  []Result `json:"results"`
}

type Result struct {
    Name string `json:"name"`
    URL  string `json:"url"`
}

type LocationAreaDetail struct {
    Name              string             `json:"name"`
    PokemonEncounters []PokemonEncounter `json:"pokemon_encounters"`
}

type PokemonEncounter struct {
    Pokemon struct {
        Name string `json:"name"`
    } `json:"pokemon"`
}

type Pokemon struct {
    Name           string
    BaseExperience int
    Height         int
    Weight         int
    Stats          map[string]int
    Types          []string
}

func FetchLocationAreas(url string) (*LocationAreaList, error) {
    if Cache != nil {
        if b, ok := Cache.Get(url); ok {
            var list LocationAreaList
            if err := json.Unmarshal(b, &list); err == nil {
                return &list, nil
            }
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

    var list LocationAreaList
    if err := json.Unmarshal(body, &list); err != nil {
        return nil, err
    }

    if Cache != nil {
        Cache.Add(url, body)
    }
    return &list, nil
}

func FetchLocationAreaDetail(url string) (*LocationAreaDetail, error) {
    if Cache != nil {
        if b, ok := Cache.Get(url); ok {
            var detail LocationAreaDetail
            if err := json.Unmarshal(b, &detail); err == nil {
                return &detail, nil
            }
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

    var detail LocationAreaDetail
    if err := json.Unmarshal(body, &detail); err != nil {
        return nil, err
    }

    if Cache != nil {
        Cache.Add(url, body)
    }
    return &detail, nil
}

func FetchPokemon(name string) (*Pokemon, error) {
    url := PokeAPIBase + "pokemon/" + name + "/"

    if Cache != nil {
        if b, ok := Cache.Get(url); ok {
            var p struct {
                Name           string `json:"name"`
                BaseExperience int    `json:"base_experience"`
            }
            if err := json.Unmarshal(b, &p); err == nil {
                return &Pokemon{Name: p.Name, BaseExperience: p.BaseExperience}, nil
            }
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

    var p struct {
        Name           string `json:"name"`
        BaseExperience int    `json:"base_experience"`
        Height         int    `json:"height"`
        Weight         int    `json:"weight"`
        Stats []struct {
            BaseStat int `json:"base_stat"`
            Stat     struct{
                Name string `json:"name"`
            } `json:"stat"`
        } `json:"stats"`
        Types []struct {
            Type struct{
                Name string `json:"name"`
            } `json:"type"`
        } `json:"types"`
    }
    if err := json.Unmarshal(body, &p); err != nil {
        return nil, err
    }

    if Cache != nil {
        Cache.Add(url, body)
    }

    statsMap := make(map[string]int)
    for _, s := range p.Stats {
        statsMap[s.Stat.Name] = s.BaseStat
    }
    types := make([]string, 0, len(p.Types))
    for _, t := range p.Types {
        types = append(types, t.Type.Name)
    }

    return &Pokemon{
        Name: p.Name,
        BaseExperience: p.BaseExperience,
        Height: p.Height,
        Weight: p.Weight,
        Stats: statsMap,
        Types: types,
    }, nil
}
