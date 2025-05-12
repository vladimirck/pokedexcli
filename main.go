package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/vladimirck/pokedexcli/internal/pokecache"
)

func main() {

	cfg := Config{
		Next:        "",
		Previous:    "",
		cache:       pokecache.NewCache(time.Second * 5),
		pokemonData: map[string]PokemonData{},
	}
	scanner := bufio.NewScanner(os.Stdin)
	var text string
	cmds := make(map[string]cliCommand)

	cmds = map[string]cliCommand{
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
			description: "Displays a list of the next 20 locations of the Pokemon world",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays a list of the previous 20 locations of the Pokemon world",
			callback:    commandMapb,
		},
		"explore": {
			name:        "explore",
			description: "Explore and area and displays the pokemon found in it",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Try to catch a pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Print the stats of catched pokemons",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "List all catched pokemons",
			callback:    commandPokedex,
		},
	}
	for {
		fmt.Print("Pokedex > ")
		if scanner.Scan() {
			text = scanner.Text()
		}
		words := cleanInput(text)
		if len(words) == 0 {
			continue
		}
		if len(words) == 2 {
			cfg.cmdArg = words[1]
		}
		cmd, ok := cmds[words[0]]
		if !ok {
			fmt.Printf("Not a valid command\n")
			continue
		}

		err := cmd.callback(&cfg)
		if err != nil {
			fmt.Printf("an error ocurred during the execution of %s command: %v\n", cmd.name, err)
		}
	}
}
