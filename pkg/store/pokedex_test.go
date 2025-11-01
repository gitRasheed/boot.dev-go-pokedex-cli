package store

import (
    "testing"

    "github.com/gitRasheed/boot.dev-go-pokedex-cli/pkg/api"
)

func TestStoreAddGet(t *testing.T) {
    s := NewStore()
    p := api.Pokemon{Name: "bob"}
    s.Add(p)

    got, ok := s.Get("bob")
    if !ok {
        t.Fatalf("expected to find pokemon 'bob' in store")
    }
    if got.Name != "bob" {
        t.Fatalf("expected name bob, got %q", got.Name)
    }
}

func TestStoreListNames(t *testing.T) {
    s := NewStore()
    s.Add(api.Pokemon{Name: "a"})
    s.Add(api.Pokemon{Name: "b"})
    s.Add(api.Pokemon{Name: "c"})

    names := s.ListNames()
    if len(names) != 3 {
        t.Fatalf("expected 3 names, got %d", len(names))
    }

    m := map[string]bool{}
    for _, n := range names { m[n] = true }
    for _, want := range []string{"a", "b", "c"} {
        if !m[want] {
            t.Fatalf("expected name %q in list", want)
        }
    }
}
