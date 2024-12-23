package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Pokemon struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Type       []string `json:"type"`
	BaseExp    int      `json:"base_exp"`
	Speed      int      `json:"speed"`
	Attack     int      `json:"attack"`
	Defense    int      `json:"defense"`
	SpecialAtk int      `json:"special_atk"`
	SpecialDef int      `json:"speical_def"`
	HP         int      `json:"hp"`
	EV         float64  `json:"ev"`
}

func main() {
	pokemonURL := "https://pokemondb.net/pokedex/national"

	pokemons, err := fetchData(pokemonURL)
	if err != nil {
		fmt.Println("Error fetching data:", err)
		return
	}

	file, err := json.MarshalIndent(pokemons, "", "  ")
	if err != nil {
		fmt.Println("Error encoding pokemon data:", err)
		return
	}

	err = os.WriteFile("pokedex.json", file, 0644)
	if err != nil {
		fmt.Println("Error saving data to JSON file:", err)
		return
	}

	fmt.Println("Pokemon data has been saved successfully.")
}

func fetchData(url string) ([]Pokemon, error) {
	//GET request
	res, err := http.Get(url)
	if err != nil {
		fmt.Println("failed to fetch HTML from PokemonDB: ", err)
		return nil, err
	}
	defer res.Body.Close()

	//Check the status
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	//Create empty pokemon list
	var pokemons []Pokemon
	var count = 1

	//Go through over each list item in the list containing Pokemon data
	doc.Find(".infocard").Each(func(i int, s *goquery.Selection) {
		if len(pokemons) > 200 {
			return
		}
		var pokemon Pokemon

		//Get pokemon ID String
		idStr, _ := s.Find(".infocard-cell-data").First().Attr("data-sprite")
		//Convert ID String into Integer
		id, _ := strconv.Atoi(strings.TrimPrefix(idStr, "/sprites/"))
		pokemon.ID = id

		//Get Pokemon name
		pokemon.Name = strings.TrimSpace(s.Find(".ent-name").Text())

		//Get all Pokemon types
		s.Find(".itype").Each(func(i int, s *goquery.Selection) {
			pokemon.Type = append(pokemon.Type, strings.TrimSpace(s.Text()))
		})
		pokemon.ID = count
		count++

		//Get Pokemon detail page URL
		detailURL, _ := s.Find("a").First().Attr("href")
		fullDetailURL := "https://pokemondb.net" + detailURL

		//Fetch Pokemon detail data from detail page
		if err := FetchDetails(fullDetailURL, &pokemon); err != nil {
			fmt.Printf("Error fetching details for %s: %v\n", pokemon.Name, err)
			return
		}
		fmt.Println("Append pokemon: ", pokemon.Name)
		pokemons = append(pokemons, pokemon)
	})

	return pokemons, nil
}

func FetchDetails(URL string, pokemon *Pokemon) error {
	//GET request
	res, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	//Check the status code
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	var baseExpFound, evFound, statsFound bool

	//Find all table from class "vitals-table"
	doc.Find(".vitals-table").Each(func(i int, s *goquery.Selection) {
		s.Find("tbody tr").Each(func(i int, row *goquery.Selection) {
			attrName := strings.TrimSpace(row.Find("th").Text())
			attrValue := strings.TrimSpace(row.Find("td").Text())

			switch attrName {
			case "Base Exp.":
				pokemon.BaseExp, _ = strconv.Atoi(attrValue)
				baseExpFound = true
			case "EV yield":
				evStr := strings.Fields(attrValue)[0]
				pokemon.EV, _ = strconv.ParseFloat(evStr, 64)
				evFound = true
			default:
				//Check if attrValue contains stats
				statParts := strings.Split(attrValue, "\n")
				if len(statParts) == 3 {
					statValue, _ := strconv.Atoi(statParts[0])
					statsFound = true
					switch attrName {
					case "HP":
						pokemon.HP = statValue
					case "Attack":
						pokemon.Attack = statValue
					case "Defense":
						pokemon.Defense = statValue
					case "Sp. Atk":
						pokemon.SpecialAtk = statValue
					case "Sp. Def":
						pokemon.SpecialDef = statValue
					case "Speed":
						pokemon.Speed = statValue
					}
				}
			}
		})
		//Check statFound
		if statsFound {
			return
		}
	})
	//Check all data is fetch successfully
	if !baseExpFound || !evFound || !statsFound {
		return fmt.Errorf("all required data are not fetched successfully")
	}

	return nil
}
