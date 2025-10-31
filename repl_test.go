package main

import "testing"

func TestCleanInput(t *testing.T) {
    cases := []struct {
        input    string
        expected []string
    }{
        {
            input:    "  hello  world  ",
            expected: []string{"hello", "world"},
        },
        {
            input:    "Charmander Bulbasaur PIKACHU",
            expected: []string{"charmander", "bulbasaur", "pikachu"},
        },
        {
            input:    "",
            expected: []string{},
        },
        {
            input:    "   ",
            expected: []string{},
        },
        {
            input:    "  MIXED   case  Test  ",
            expected: []string{"mixed", "case", "test"},
        },
    }

    for _, c := range cases {
        actual := cleanInput(c.input)

        if len(actual) != len(c.expected) {
            t.Errorf("input %q: expected %d words, got %d (actual: %#v)", c.input, len(c.expected), len(actual), actual)
            continue
        }

        for i := range actual {
            if actual[i] != c.expected[i] {
                t.Errorf("input %q: expected word %d to be %q, got %q", c.input, i, c.expected[i], actual[i])
            }
        }
    }
}
