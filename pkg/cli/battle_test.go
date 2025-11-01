package cli

import (
    "bytes"
    "math/rand"
    "strings"
    "testing"

    "github.com/gitRasheed/boot.dev-go-pokedex-cli/pkg/api"
    "github.com/gitRasheed/boot.dev-go-pokedex-cli/pkg/store"
)

func TestBattleDeterministic(t *testing.T) {
    out := &bytes.Buffer{}
    rng := rand.New(rand.NewSource(1))
    c := &CLI{in: nil, out: out, store: store.NewStore(), rng: rng}

    weak := api.Pokemon{Name: "weak", Stats: map[string]int{"hp": 10, "attack": 1, "defense": 1, "speed": 10}}
    strong := api.Pokemon{Name: "strong", Stats: map[string]int{"hp": 100, "attack": 1000, "defense": 1, "speed": 5}}
    c.store.Add(weak)
    c.store.Add(strong)

    c.cmdBattle([]string{"weak", "strong"})
    got := out.String()
    if !strings.Contains(got, "Battle: weak vs strong") {
        t.Fatalf("unexpected output, missing header: %s", got)
    }
    if !strings.Contains(got, "strong wins!") {
        t.Fatalf("expected strong to win; output:\n%s", got)
    }
}
