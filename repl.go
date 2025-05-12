package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/vladimirck/pokedexcli/internal/pokecache"
)

type PokemonLocator struct {
	EncounterMethodRates []struct {
		EncounterMethod struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"encounter_method"`
		VersionDetails []struct {
			Rate    int `json:"rate"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"encounter_method_rates"`
	GameIndex int `json:"game_index"`
	ID        int `json:"id"`
	Location  struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"location"`
	Name  string `json:"name"`
	Names []struct {
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
		Name string `json:"name"`
	} `json:"names"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		VersionDetails []struct {
			EncounterDetails []struct {
				Chance          int   `json:"chance"`
				ConditionValues []any `json:"condition_values"`
				MaxLevel        int   `json:"max_level"`
				Method          struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"method"`
				MinLevel int `json:"min_level"`
			} `json:"encounter_details"`
			MaxChance int `json:"max_chance"`
			Version   struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
}

type PokemonData struct {
	BaseExperience         int    `json:"base_experience"`
	Height                 int    `json:"height"`
	ID                     int    `json:"id"`
	IsDefault              bool   `json:"is_default"`
	LocationAreaEncounters string `json:"location_area_encounters"`
	Name                   string `json:"name"`
	Order                  int    `json:"order"`
	Weight                 int    `json:"weight"`
}

type cliCommand struct {
	name        string
	description string
	callback    func(*Config) error
}

type LocationAreas struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type Config struct {
	Next        string
	Previous    string
	cmdArg      string
	cache       *pokecache.Cache
	pokemonData map[string]PokemonData
}

func cleanInput(text string) []string {
	trimedText := strings.ToLower(strings.Trim(text, " \t\n"))
	words := strings.Fields(trimedText)
	return words
}

func commandExit(cfg *Config) error {
	fmt.Printf("Closing the Pokedex... Goodbye!\n")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *Config) error {
	fmt.Printf("Welcome to the Pokedex!\n")
	fmt.Printf("Usage:\n\n")
	fmt.Printf("help: Displays a help message\n")
	fmt.Printf("exit: Exit the Pokedex\n")
	return nil
}

func commandMap(cfg *Config) error {
	locArea := LocationAreas{}

	if len(cfg.Next) == 0 && len(cfg.Previous) != 0 { //we are in the last page
		return fmt.Errorf("we are in the last page already.\n")
	}

	if len(cfg.Next) == 0 && len(cfg.Previous) == 0 { //the first time the functiona is called
		cfg.Next = "https://pokeapi.co/api/v2/location-area/"
	}

	body, ok := cfg.cache.Get(cfg.Next)

	if !ok {
		res, err := http.Get(cfg.Next)
		if err != nil {
			return fmt.Errorf("error contacting the server: %v", err)
		}
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return fmt.Errorf("response status from the server: %s", res.Status)
		}
		body, err = io.ReadAll(res.Body)
		defer res.Body.Close()
		if err != nil {
			return fmt.Errorf("the response could not be read: %v", err)
		}

		cfg.cache.Add(cfg.Next, body)
	}

	if err := json.Unmarshal(body, &locArea); err != nil {
		return fmt.Errorf("the JSON data could not be unmarshalled: %v", err)
	}

	//fmt.Printf("Next: %s\nPrevious: %s\n\n", locArea.Next, locArea.Previous)

	cfg.Next = locArea.Next
	cfg.Previous = locArea.Previous

	for _, area := range locArea.Results {
		fmt.Println(area.Name)
	}

	return nil
}

func commandMapb(cfg *Config) error {
	locArea := LocationAreas{}

	if len(cfg.Next) != 0 && len(cfg.Previous) == 0 { //we are in the last page
		return fmt.Errorf("we are in the first page already.\n")
	}

	if len(cfg.Next) == 0 && len(cfg.Previous) == 0 { //the first time the functiona is called
		return fmt.Errorf("map has never been called.\n")
	}

	body, ok := cfg.cache.Get(cfg.Next)

	if !ok {
		res, err := http.Get(cfg.Previous)
		if err != nil {
			return fmt.Errorf("error contacting the server: %v\n", err)
		}

		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return fmt.Errorf("eesponse status from the server: %s\n", res.Status)
		}

		body, err = io.ReadAll(res.Body)
		defer res.Body.Close()
		if err != nil {
			return fmt.Errorf("the response could not be read: %v\n", err)
		}

		cfg.cache.Add(cfg.Previous, body)
	}

	if err := json.Unmarshal(body, &locArea); err != nil {
		return fmt.Errorf("the JSON data could not be unmarshalled: %v\n", err)
	}

	//fmt.Printf("Next: %s\nPrevious: %s\n\n", locArea.Next, locArea.Previous)

	cfg.Next = locArea.Next
	cfg.Previous = locArea.Previous

	for _, area := range locArea.Results {
		fmt.Println(area.Name)
	}

	return nil
}

func commandExplore(cfg *Config) error {

	if len(cfg.cmdArg) == 0 { //we are in the last page
		return fmt.Errorf("the area name was not given by the user.\n")
	}
	locPokemons := PokemonLocator{}
	url := "https://pokeapi.co/api/v2/location-area/" + cfg.cmdArg

	body, ok := cfg.cache.Get(url)

	if !ok {
		res, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("error contacting the server: %v", err)
		}
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return fmt.Errorf("response status from the server: %s", res.Status)
		}
		body, err = io.ReadAll(res.Body)
		defer res.Body.Close()
		if err != nil {
			return fmt.Errorf("the response could not be read: %v", err)
		}

		cfg.cache.Add(url, body)
	}

	if err := json.Unmarshal(body, &locPokemons); err != nil {
		return fmt.Errorf("the JSON data could not be unmarshalled: %v", err)
	}

	for _, pokemon := range locPokemons.PokemonEncounters {
		fmt.Printf(" - %s\n", pokemon.Pokemon.Name)
	}

	return nil
}

func commandCatch(cfg *Config) error {

	if len(cfg.cmdArg) == 0 { //we are in the last page
		return fmt.Errorf("the pokemon name was not given by the user.\n")
	}
	pokemonData := PokemonData{}
	url := "https://pokeapi.co/api/v2/pokemon/" + cfg.cmdArg
	//fmt.Printf("url: %s\n", url)

	body, ok := cfg.cache.Get(url)

	if !ok {
		res, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("error contacting the server: %v", err)
		}
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return fmt.Errorf("response status from the server: %s", res.Status)
		}
		body, err = io.ReadAll(res.Body)
		defer res.Body.Close()
		if err != nil {
			return fmt.Errorf("the response could not be read: %v", err)
		}

		cfg.cache.Add(url, body)
	}

	if err := json.Unmarshal(body, &pokemonData); err != nil {
		return fmt.Errorf("the JSON data could not be unmarshalled: %v", err)
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", cfg.cmdArg)
	rndNumber := rand.Intn(650)

	if rndNumber > pokemonData.BaseExperience {
		fmt.Printf("%s was caught!\n", cfg.cmdArg)
		cfg.pokemonData[cfg.cmdArg] = pokemonData
		return nil
	}

	fmt.Printf("%s escaped!\n", cfg.cmdArg)

	return nil
}

func commandInspect(cfg *Config) error {
	pokemon, ok := cfg.pokemonData[cfg.cmdArg]

	if !ok {
		return fmt.Errorf("%s is not among the pokemons you have catched", cfg.cmdArg)
	}

	fmt.Printf("Name:\t\t%v\n", pokemon.Name)
	fmt.Printf("Name:\t\t%v\n", pokemon.Height)
	fmt.Printf("Weight:\t\t%v\n", pokemon.Weight)
	fmt.Printf("Order:\t\t%v\n", pokemon.Order)
	fmt.Printf("Base Exp:\t%v\n", pokemon.BaseExperience)

	return nil
}

func commandPokedex(cfg *Config) error {
	if len(cfg.pokemonData) == 0 {
		return fmt.Errorf("no pokemon has been catched!")
	}
	fmt.Printf("Your Pokedex:\n")

	for name, _ := range cfg.pokemonData {
		fmt.Printf(" - %s\n", name)
	}

	return nil
}
