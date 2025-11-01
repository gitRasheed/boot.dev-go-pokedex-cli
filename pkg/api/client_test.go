package api

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestFetchLocationAreas(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/location-area/" {
            w.Write([]byte(`{
  "count": 2,
  "next": "http://example.com/next",
  "previous": null,
  "results": [
    {"name":"area-1","url":"/location-area/area-1/"},
    {"name":"area-2","url":"/location-area/area-2/"}
  ]
}`))
            return
        }
        http.NotFound(w, r)
    }))
    defer ts.Close()

    PokeAPIBase = ts.URL + "/"
    LocationAreaBase = PokeAPIBase + "location-area/"
    Cache = nil

    list, err := FetchLocationAreas(LocationAreaBase)
    if err != nil {
        t.Fatalf("FetchLocationAreas error: %v", err)
    }
    if len(list.Results) != 2 {
        t.Fatalf("expected 2 results, got %d", len(list.Results))
    }
}

func TestFetchLocationAreaDetail(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/location-area/area-1/" {
            w.Write([]byte(`{
  "name": "area-1",
  "pokemon_encounters": [
    {"pokemon": {"name": "p1"}},
    {"pokemon": {"name": "p2"}}
  ]
}`))
            return
        }
        http.NotFound(w, r)
    }))
    defer ts.Close()

    PokeAPIBase = ts.URL + "/"
    LocationAreaBase = PokeAPIBase + "location-area/"
    Cache = nil

    detail, err := FetchLocationAreaDetail(LocationAreaBase + "area-1/")
    if err != nil {
        t.Fatalf("FetchLocationAreaDetail error: %v", err)
    }
    if len(detail.PokemonEncounters) != 2 {
        t.Fatalf("expected 2 encounters, got %d", len(detail.PokemonEncounters))
    }
}

func TestFetchPokemon(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/pokemon/pikachu/" {
            w.Write([]byte(`{
  "name": "pikachu",
  "base_experience": 112,
  "height": 4,
  "weight": 60,
  "stats": [
    {"base_stat": 35, "stat": {"name": "hp"}},
    {"base_stat": 55, "stat": {"name": "attack"}}
  ],
  "types": [
    {"type": {"name": "electric"}}
  ]
}`))
            return
        }
        http.NotFound(w, r)
    }))
    defer ts.Close()

    PokeAPIBase = ts.URL + "/"
    Cache = nil

    p, err := FetchPokemon("pikachu")
    if err != nil {
        t.Fatalf("FetchPokemon error: %v", err)
    }
    if p.Name != "pikachu" {
        t.Fatalf("expected name pikachu, got %s", p.Name)
    }
    if p.BaseExperience != 112 {
        t.Fatalf("expected base_experience 112, got %d", p.BaseExperience)
    }
    if p.Stats["hp"] != 35 {
        t.Fatalf("expected hp 35, got %d", p.Stats["hp"])
    }
}
