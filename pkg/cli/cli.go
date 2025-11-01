package cli

import (
    "bufio"
    "fmt"
    "io"
    "math/rand"
    "os"
    "sort"
    "strings"
    "time"

    "github.com/gitRasheed/boot.dev-go-pokedex-cli/internal/pokecache"
    "github.com/gitRasheed/boot.dev-go-pokedex-cli/pkg/api"
    "github.com/gitRasheed/boot.dev-go-pokedex-cli/pkg/store"
)

type CLI struct {
    in  io.Reader
    out io.Writer
    store *store.Store
    pager struct{
        Next *string
        Previous *string
    }
    rng *rand.Rand
}

var ballModifiers = map[string]float64{
    "pokeball": 1.0,
    "greatball": 1.5,
    "ultraball": 2.0,
    "masterball": 100.0,
}

func Run(in io.Reader, out io.Writer) {
    c := &CLI{in: in, out: out, store: store.NewStore()}

    if c.rng == nil {
        c.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
    }

    api.Cache = pokecache.NewCache(5 * 1000000000)

    scanner := bufio.NewScanner(c.in)
    for {
        fmt.Fprint(c.out, "Pokedex > ")
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
        cmd := words[0]
        args := []string{}
        if len(words) > 1 {
            args = words[1:]
        }

        switch cmd {
        case "exit":
            fmt.Fprintln(c.out, "Closing the Pokedex... Goodbye!")
            os.Exit(0)
        case "help":
            c.cmdHelp()
        case "map":
            c.cmdMap()
        case "mapb":
            c.cmdMapBack()
        case "explore":
            c.cmdExplore(args)
        case "catch":
            c.cmdCatch(args)
        case "pokedex":
            c.cmdPokedex()
        case "inspect":
            c.cmdInspect(args)
        case "battle":
            c.cmdBattle(args)
        default:
            fmt.Fprintln(c.out, "Unknown command")
        }
    }
}

func cleanInput(text string) []string {
    s := strings.TrimSpace(text)
    if s == "" {
        return []string{}
    }
    s = strings.ToLower(s)
    return strings.Fields(s)
}

func (c *CLI) cmdHelp() {
    fmt.Fprintln(c.out, "Welcome to the Pokedex!")
    fmt.Fprintln(c.out, "Usage:")
    fmt.Fprintln(c.out)
    fmt.Fprintln(c.out, "  exit                  - Exit the Pokedex")
    fmt.Fprintln(c.out, "  help                  - Show this help message")
    fmt.Fprintln(c.out, "  map                   - Show next page of location areas")
    fmt.Fprintln(c.out, "  mapb                  - Show previous page of location areas")
    fmt.Fprintln(c.out, "  explore <area>        - Explore a location area and list encountered Pokémon")
    fmt.Fprintln(c.out, "  catch <pokemon> [ball]- Catch a Pokémon; optional ball types: pokeball, greatball, ultraball, masterball")
    fmt.Fprintln(c.out, "  pokedex               - List caught Pokémon")
    fmt.Fprintln(c.out, "  inspect <pokemon>     - Show details for a caught Pokémon")
    fmt.Fprintln(c.out, "  battle <p1> <p2>      - Simulate a simple battle between two caught Pokémon")
}

func (c *CLI) cmdMap() {
    url := api.LocationAreaBase
    if c.pager.Next != nil && *c.pager.Next != "" {
        url = *c.pager.Next
    }
    list, err := api.FetchLocationAreas(url)
    if err != nil {
        fmt.Fprintln(c.out, err)
        return
    }
    for _, r := range list.Results {
        fmt.Fprintln(c.out, r.Name)
    }
    c.pager.Next = list.Next
    c.pager.Previous = list.Previous
}

func (c *CLI) cmdMapBack() {
    if c.pager.Previous == nil || *c.pager.Previous == "" {
        fmt.Fprintln(c.out, "you're on the first page")
        return
    }
    list, err := api.FetchLocationAreas(*c.pager.Previous)
    if err != nil {
        fmt.Fprintln(c.out, err)
        return
    }
    for _, r := range list.Results {
        fmt.Fprintln(c.out, r.Name)
    }
    c.pager.Next = list.Next
    c.pager.Previous = list.Previous
}

func (c *CLI) cmdExplore(args []string) {
    if len(args)==0 {
        fmt.Fprintln(c.out, "usage: explore <area>")
        return
    }
    name := args[0]
    fmt.Fprintf(c.out, "Exploring %s...\n", name)
    url := api.LocationAreaBase + name + "/"
    detail, err := api.FetchLocationAreaDetail(url)
    if err != nil {
        fmt.Fprintln(c.out, err)
        return
    }
    fmt.Fprintln(c.out, "Found Pokemon:")
    for _, pe := range detail.PokemonEncounters {
        fmt.Fprintf(c.out, " - %s\n", pe.Pokemon.Name)
    }
}

func (c *CLI) cmdCatch(args []string) {
    if len(args)==0 {
        fmt.Fprintln(c.out, "usage: catch <pokemon> [ball]")
        return
    }
    name := args[0]
    ball := "pokeball"
    if len(args)>1 { ball = strings.ToLower(args[1]) }
    if _, ok := ballModifiers[ball]; !ok { fmt.Fprintf(c.out, "unknown ball '%s', using pokeball\n", ball); ball = "pokeball" }
    fmt.Fprintf(c.out, "Throwing a %s at %s...\n", ball, name)
    p, err := api.FetchPokemon(name)
    if err != nil { fmt.Fprintln(c.out, err); return }
    baseChance := 0.5 - float64(p.BaseExperience)/500.0
    if baseChance < 0.01 { baseChance = 0.01 }
    if baseChance > 0.99 { baseChance = 0.99 }
    chance := baseChance * ballModifiers[ball]
    if chance > 0.9999 { chance = 0.9999 }
    if c.randFloat() < chance {
        fmt.Fprintf(c.out, "%s was caught!\n", p.Name)
        c.store.Add(*p)
        fmt.Fprintln(c.out, "You may now inspect it with the inspect command.")
    } else {
        fmt.Fprintf(c.out, "%s escaped!\n", p.Name)
    }
}

func (c *CLI) randFloat() float64 {
    if c.rng == nil {
        return rand.Float64()
    }
    return c.rng.Float64()
}

func (c *CLI) cmdBattle(args []string) {
    if len(args) < 2 {
        fmt.Fprintln(c.out, "usage: battle <pokemon1> <pokemon2>")
        return
    }
    aName := args[0]
    bName := args[1]

    a, ok := c.store.Get(aName)
    if !ok {
        fmt.Fprintf(c.out, "you have not caught %s\n", aName)
        return
    }
    b, ok := c.store.Get(bName)
    if !ok {
        fmt.Fprintf(c.out, "you have not caught %s\n", bName)
        return
    }

    aHP := a.Stats["hp"]
    bHP := b.Stats["hp"]
    if aHP <= 0 {
        aHP = 10
    }
    if bHP <= 0 {
        bHP = 10
    }

    aAtk := a.Stats["attack"]
    bAtk := b.Stats["attack"]
    aDef := a.Stats["defense"]
    bDef := b.Stats["defense"]
    aSpd := a.Stats["speed"]
    bSpd := b.Stats["speed"]

    fmt.Fprintf(c.out, "Battle: %s vs %s\n", aName, bName)

    attackerIsA := true
    if bSpd > aSpd {
        attackerIsA = false
    }

    round := 1
    for aHP > 0 && bHP > 0 {
        if attackerIsA {
            dmg := c.calcDamage(aAtk, bDef)
            bHP -= dmg
            if bHP < 0 { bHP = 0 }
            fmt.Fprintf(c.out, "%s hits %s for %d damage (%d HP left)\n", aName, bName, dmg, bHP)
        } else {
            dmg := c.calcDamage(bAtk, aDef)
            aHP -= dmg
            if aHP < 0 { aHP = 0 }
            fmt.Fprintf(c.out, "%s hits %s for %d damage (%d HP left)\n", bName, aName, dmg, aHP)
        }
        attackerIsA = !attackerIsA
        round++
        if round > 200 {
            fmt.Fprintln(c.out, "battle ended in a draw")
            return
        }
    }

    if aHP <= 0 && bHP <= 0 {
        fmt.Fprintln(c.out, "It's a draw!")
    } else if bHP <= 0 {
        fmt.Fprintf(c.out, "%s wins!\n", aName)
    } else {
        fmt.Fprintf(c.out, "%s wins!\n", bName)
    }
}

func (c *CLI) calcDamage(atk, def int) int {
    if atk <= 0 { atk = 5 }
    if def < 0 { def = 0 }
    base := atk - def/2
    if base < 1 { base = 1 }
    factor := 0.85 + c.randFloat()*0.3
    dmg := int(float64(base) * factor)
    if dmg < 1 { dmg = 1 }
    return dmg
}

func (c *CLI) cmdPokedex() {
    fmt.Fprintln(c.out, "Your Pokedex:")
    names := c.store.ListNames()
    if len(names)==0 { fmt.Fprintln(c.out, " (empty)") ; return }
    sort.Strings(names)
    for _, n := range names { fmt.Fprintf(c.out, " - %s\n", n) }
}

func (c *CLI) cmdInspect(args []string) {
    if len(args)==0 { fmt.Fprintln(c.out, "usage: inspect <pokemon>"); return }
    name := args[0]
    p, ok := c.store.Get(name)
    if !ok { fmt.Fprintln(c.out, "you have not caught that pokemon"); return }
    if p.Height == 0 && p.Weight == 0 && len(p.Stats) == 0 {
        if fresh, err := api.FetchPokemon(name); err == nil {
            c.store.Add(*fresh)
            p = *fresh
        }
    }
    fmt.Fprintf(c.out, "Name: %s\n", p.Name)
    fmt.Fprintf(c.out, "Height: %d\n", p.Height)
    fmt.Fprintf(c.out, "Weight: %d\n", p.Weight)
    fmt.Fprintln(c.out, "Stats:")
    order := []string{"hp","attack","defense","special-attack","special-defense","speed"}
    for _, k := range order { if v, ok := p.Stats[k]; ok { fmt.Fprintf(c.out, "  -%s: %d\n", k, v) } }
    for k, v := range p.Stats { found:=false; for _, okk := range order { if k==okk { found=true; break } }; if !found { fmt.Fprintf(c.out, "  -%s: %d\n", k, v) } }
    fmt.Fprintln(c.out, "Types:")
    for _, t := range p.Types { fmt.Fprintf(c.out, "  - %s\n", t) }
}

