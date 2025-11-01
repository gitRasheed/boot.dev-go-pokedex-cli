package store

import (
    "sync"

    "github.com/gitRasheed/boot.dev-go-pokedex-cli/pkg/api"
)

type Store struct {
    mu sync.Mutex
    entries map[string]api.Pokemon
}

func NewStore() *Store {
    return &Store{entries: make(map[string]api.Pokemon)}
}

func (s *Store) Add(p api.Pokemon) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.entries[p.Name] = p
}

func (s *Store) Get(name string) (api.Pokemon, bool) {
    s.mu.Lock()
    defer s.mu.Unlock()
    p, ok := s.entries[name]
    return p, ok
}

func (s *Store) ListNames() []string {
    s.mu.Lock()
    defer s.mu.Unlock()
    names := make([]string, 0, len(s.entries))
    for n := range s.entries {
        names = append(names, n)
    }
    return names
}
