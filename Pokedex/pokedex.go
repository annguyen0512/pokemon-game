package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Pokemon struct {
	Name       string   `json:"name"`
	Type       []string `json:"type"`
	BaseExp    int      `json:"base_exp"`
	Exp        int      `json:"exp"`
	Level      int      `json:"level"`
	EV         float64  `json:"ev"`
	Attributes struct {
		HP           int `json:"hp"`
		Attack       int `json:"attack"`
		Defense      int `json:"defense"`
		Speed        int `json:"speed"`
		SpAttack     int `json:"sp_attack"`
		SpDefense    int `json:"sp_defense"`
		DmgWhenAtked int `json:"dmg_when_atked"`
	} `json:"attributes"`
}

func main() {
	url := "https://pokemondb.net/pokedex/all"
	pokemonList := fetchData(url)
	fetchBaseExp(pokemonList)

	saveToJSON(pokemonList)
}

func fetchData(url string) []Pokemon {
	//GET the url
	res, err := http.Get(url)
	if err != nil {
		fmt.Println("failed to fetch HTML from PokemonDB: ", err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("failed to parse HTML from PokemonDB: ", err)
	}

	var pokemons []Pokemon

	//Go through each <tr> tag
	doc.Find("table.data-table tbody tr").Each(func(i int, s *goquery.Selection) {
		pokemon := Pokemon{}

		//Get name
		pokemon.Name = s.Find("td.cell-name a.ent-name").Text()

		//Get type
		typeSelection := s.Find("td.cell-icon a")
		typeSelection.Each(func(j int, typeLink *goquery.Selection) {
			pokemon.Type = append(pokemon.Type, strings.TrimSpace(typeLink.Text()))
		})

		pokemon.BaseExp = 0
		pokemon.Exp = 0
		pokemon.Level = 1
		pokemon.EV = 0.5

		//Get attributes
		s.Find("td.cell-num").Each(func(k int, attrSelection *goquery.Selection) {
			attrText := strings.TrimSpace(attrSelection.Text())
			attrValue, err := strconv.Atoi(attrText)
			if err != nil {
				log.Printf("Error parsing attribute %d for %s: %v", k, pokemon.Name, err)
				return
			}
			switch k {
			case 2:
				pokemon.Attributes.HP = attrValue
			case 3:
				pokemon.Attributes.Attack = attrValue
			case 4:
				pokemon.Attributes.Defense = attrValue
			case 5:
				pokemon.Attributes.SpAttack = attrValue
			case 6:
				pokemon.Attributes.SpDefense = attrValue
			case 7:
				pokemon.Attributes.Speed = attrValue
			}
		})

		pokemons = append(pokemons, pokemon)
	})

	return pokemons
}

func fetchBaseExp(pokemons []Pokemon) error {
	url := "https://bulbapedia.bulbagarden.net/wiki/List_of_Pokémon_by_effort_value_yield_(Generation_IX)"
	res, err := http.Get(url)
	if err != nil {
		fmt.Println("failed to GET URL: ", err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("failed to get doc: ", err)
	}

	// Map to store BaseExp values by Pokemon name for quick lookup
	baseExpMap := make(map[string]int)

	// Find each Pokemon row in the table
	doc.Find("table.sortable tbody tr").Each(func(i int, s *goquery.Selection) {
		// // Find Pokemon name
		pokemonName := strings.TrimSpace(s.Find("td").Eq(2).Find("a").Text())
		baseExpStr := strings.TrimSpace(s.Find("td").Eq(3).Text())
		//evStr := strings.TrimSpace(s.Find("td").Eq(10).Text())
		baseExp, err := strconv.Atoi(baseExpStr)
		if err != nil {
			log.Printf("Error parsing BaseExp: %v", err)
			return
		}
		baseExpMap[pokemonName] = baseExp
	})

	// Assign BaseExp to corresponding Pokémon structs
	for i := range pokemons {
		if baseExp, ok := baseExpMap[pokemons[i].Name]; ok {
			pokemons[i].BaseExp = baseExp
		}

	}

	return nil
}

func saveToJSON(data []Pokemon) {
	file, err := os.Create("pokedex.json")
	if err != nil {
		fmt.Println("Can not create JSON file: ", err)
		return
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ")

	if err := encoder.Encode(data); err != nil {
		fmt.Println("Can not write data to JSON file: ", err)
		return
	}

	fmt.Println("Save data to JSON file successfully")
}
