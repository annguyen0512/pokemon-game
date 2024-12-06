package main

import (
	"POKEMON-GAME-POKEBAT/pkg/player"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Player struct {
	ID       int               `json:"id"`
	Name     string            `json:"name"`
	Pokemons []CapturedPokemon `json:"pokemon_list"`
}

type CapturedPokemon struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Type       []string `json:"type"`
	BaseExp    int      `json:"base_exp"`
	HP         int      `json:"hp"`
	EV         float64  `json:"ev"`
	Level      int      `json:"level"`
	CurrentExp int      `json:"current_exp"`
	Speed      int      `json:"speed"`
	Attack     int      `json:"attack"`
	Defense    int      `json:"defense"`
	SpecialAtk int      `json:"special_atk"`
	SpecialDef int      `json:"special_def"`
}

type gamer struct {
	name        string
	fighterList map[int]player.CapturedPokemon
	fighter     player.CapturedPokemon
	addr        *net.UDPAddr
}

var (
	players      = make(map[string]player.Player)
	gamers       = make(map[string]*gamer) //Store pointers to gamers
	mutex        sync.Mutex
	battleActive bool
	serverConn   *net.UDPConn
	divider      = "________________________-"
)

func parseInput(input string, maxChoices int) []int {
	choices := []int{}
	for _, char := range input {
		if char != ' ' {
			choice, err := strconv.Atoi(string(char))
			//Check valid selection input
			if err != nil || choice < 1 || choice > maxChoices {
				continue
			}
			choices = append(choices, choice)
		}
	}
	return choices
}

func switchTurn(attacker, defender *gamer) {
	*attacker, *defender = *defender, *attacker
}

func wait(i int) {
	time.Sleep(time.Duration(i) * time.Second)
}

func choosePokemon(g gamer, p player.Player, conn *net.UDPConn) []player.CapturedPokemon {
	//Create list to store chosen pokemons
	var chosenPokemons []player.CapturedPokemon

	//Display player's name and pokemon list
	sendMessage(conn, g.addr, fmt.Sprintf("Player: %s\n", g.name))
	for _, pokemon := range p.Pokemons {
		sendMessage(conn, g.addr, showPokemonProfile(pokemon))
	}

	sendMessage(conn, g.addr, "Select three Pokemon (Please enter the pokemon ID separated by space): ")

	//Receive and analysis selections from client
	message := receiveMessage(conn)
	fmt.Println("Received message:", message)
	choices := parseInput(message, 3)

	//Append each selected pokemon into chosen list
	for _, choice := range choices {
		chosenPokemons = append(chosenPokemons, p.Pokemons[choice-1])
	}

	fmt.Println(chosenPokemons)

	return chosenPokemons
}

func showPokemonProfile(pokemon player.CapturedPokemon) string {
	profile := fmt.Sprintf("%d. Name: %s | ", pokemon.ID, pokemon.Name)
	profile += fmt.Sprintf("Type: %v | ", pokemon.Type)
	profile += fmt.Sprintf("Base Exp: %d |", pokemon.BaseExp)
	profile += fmt.Sprintf("HP: %d | ", pokemon.HP)
	profile += fmt.Sprintf("EV: %.1f | ", pokemon.EV)
	profile += fmt.Sprintf("Level: %d | ", pokemon.Level)
	profile += fmt.Sprintf("Current Exp: %d\n", pokemon.CurrentExp)
	profile += fmt.Sprintf("Speed: %d | ", pokemon.Speed)
	profile += fmt.Sprintf("Attack: %d | ", pokemon.Attack)
	profile += fmt.Sprintf("Defense: %d | ", pokemon.Defense)
	profile += fmt.Sprintf("Special Atk: %d | ", pokemon.SpecialAtk)
	profile += fmt.Sprintf("Special Def: %d\n", pokemon.SpecialDef)
	profile += "\n"
	return profile
}

// Send message to client
func sendMessage(conn *net.UDPConn, addr *net.UDPAddr, message string) {
	conn.WriteToUDP([]byte(message), addr)
}

// Receive message from client
func receiveMessage(conn *net.UDPConn) string {
	buffer := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("Error reading from UDP:", err)
		return ""
	}
	return string(buffer[:n])
}

func handleClient(conn *net.UDPConn, addr *net.UDPAddr, playerName string) {
	//Ensure multiple client can connect to server
	mutex.Lock()
	defer mutex.Unlock()

	p, exist := players[playerName]
	if !exist {
		sendMessage(conn, addr, "ERROR: No player found with name: "+playerName)
	}

	g := &gamer{
		name:        playerName,
		fighterList: make(map[int]player.CapturedPokemon),
		addr:        addr,
	}

	choosePokemons := choosePokemon(*g, p, conn)
	for _, pokemon := range choosePokemons {
		g.fighterList[pokemon.ID] = pokemon
	}

	g.fighter = choosePokemons[0]
	gamers[playerName] = g
	sendMessage(conn, addr, "SUCCESS: You have registered with name: "+playerName)
	fmt.Println("Player registered: " + playerName)
}

func selectFighter(g *gamer, conn *net.UDPConn) bool {
	//Variable to verify if there are available fighters
	available := false

	for _, pokemon := range g.fighterList {
		if pokemon.HP > 0 {
			available = true
			sendMessage(conn, g.addr, showPokemonProfile(pokemon))
		}
	}

	//If user don't have any available pokemon
	if !available {
		sendMessage(conn, g.addr, "You don't have any available fighter left!")
		return false
	}

	//If user have available pokemon
	for {
		sendMessage(conn, g.addr, "Select your fighter by ID: ")
		input := receiveMessage(conn)
		id, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			sendMessage(conn, g.addr, "BAD INPUT: Invalid input. Please enter valid pokemon ID.")
			continue
		}

		selectedPokemon, ok := g.fighterList[id]
		//Check if the selected pokemon is valid
		if !ok || selectedPokemon.HP <= 0 {
			sendMessage(conn, g.addr, "BAD SELECTION: Invalid ID or your selected pokemon has fainted.")
			continue
		}

		g.fighter = selectedPokemon
		sendMessage(conn, g.addr, fmt.Sprintf("Selected fighter: %s\n", g.fighter.Name))
		return true
	}
}

func startBattle(gamer1, gamer2 *gamer) {
	fmt.Println("Battle begins!")

	sendMessage(serverConn, gamer1.addr, "Two players connected. The battle is starting!\n")
	sendMessage(serverConn, gamer2.addr, "Two players connected. The battle is starting!\n")

	var attacker, defender *gamer
	if gamer1.fighter.Speed >= gamer2.fighter.Speed {
		attacker = gamer1
		defender = gamer2
	} else {
		attacker = gamer2
		defender = gamer1
	}

	i := 0
	for {
		i++
		sendMessage(serverConn, defender.addr, divider)
		sendMessage(serverConn, attacker.addr, divider)
		sendMessage(serverConn, attacker.addr, fmt.Sprintf("Turn %d, attacker %s: \n", i, attacker.name))
		sendMessage(serverConn, defender.addr, fmt.Sprintf("Turn %d, attacker %s: \n", i, attacker.name))

		attack(attacker, defender)

		sendMessage(serverConn, attacker.addr, "Turn result: ")
		sendMessage(serverConn, attacker.addr, showPokemonProfile(defender.fighter))

		sendMessage(serverConn, defender.addr, "Turn result: ")
		sendMessage(serverConn, defender.addr, showPokemonProfile(defender.fighter))

		sendMessage(serverConn, defender.addr, divider)
		sendMessage(serverConn, attacker.addr, divider)

		//If fighter of defender runs out of blood
		if defender.fighter.HP <= 0 {
			sendMessage(serverConn, defender.addr, divider)
			sendMessage(serverConn, attacker.addr, divider)

			sendMessage(serverConn, defender.addr, fmt.Sprintf("%s's %s fainted!\n, you have to switch your fighter!", defender.name, defender.fighter.Name))
			sendMessage(serverConn, attacker.addr, "The opponent's fighter is fainted!, wait for them to switch the fighter!")

			status := selectFighter(defender, serverConn)
			//Attacker win the battle
			if !status {
				distributedExperiencePoints(attacker, defender, serverConn)
				sendMessage(serverConn, defender.addr, "END BATTLE: YOU LOST!!!")
				sendMessage(serverConn, defender.addr, "END BATTLE: YOU WIN!!!")
				break
			}
		}

		sendMessage(serverConn, attacker.addr, fmt.Sprintf("%s, do you want to switch your fighter? (Y/N)", attacker.name))

		switchTurn(attacker, defender)
	}
	sendMessage(serverConn, attacker.addr, "BATTLE ENDED!")
}

func distributedExperiencePoints(winner, loser *gamer, conn *net.UDPConn) {
	totalExp := 0

	//Total EXP from loser's team
	for _, pokemon := range loser.fighterList {
		totalExp += pokemon.CurrentExp
	}

	//EXP for each pokemon
	expPerPokemon := totalExp / (3 * len(loser.fighterList))
	sendMessage(conn, winner.addr, fmt.Sprintf("RECEIVE EXP: Each pokemon of %s will get %d bonus exp!", winner.name, expPerPokemon))

	//Distribute EXP for each pokemon of winner's team
	for _, pokemon := range winner.fighterList {
		pokemon.CurrentExp += expPerPokemon
		//Check if pokemon have enough EXP to level up
		if pokemon.CurrentExp > pokemon.BaseExp {
			pokemon.CurrentExp = pokemon.BaseExp
		}
	}
}

func attack(attacker, defender *gamer) {
	// Determine if it's a special attack or not
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	isSpecialAttack := r.Intn(2) == 0

	// Print attacking and defending pokemon information
	sendMessage(serverConn, attacker.addr, "ATTACKING: ")
	sendMessage(serverConn, attacker.addr, showPokemonProfile(attacker.fighter))
	sendMessage(serverConn, defender.addr, "ATTACKING: ")
	sendMessage(serverConn, defender.addr, showPokemonProfile(attacker.fighter))
	wait(2)

	sendMessage(serverConn, defender.addr, "DEFENDING: ")
	sendMessage(serverConn, defender.addr, showPokemonProfile(defender.fighter))
	sendMessage(serverConn, attacker.addr, "DEFENDING: ")
	sendMessage(serverConn, attacker.addr, showPokemonProfile(defender.fighter))
	wait(2)

	var damage int
	if isSpecialAttack {
		damage = int(float64(attacker.fighter.SpecialAtk) - float64(defender.fighter.SpecialDef))
	} else {
		damage = attacker.fighter.Attack - defender.fighter.Defense
	}

	//Ensure damage at least 1
	if damage < 1 {
		damage = 1
	}

	//Apply damage to defender's HP
	defender.fighter.HP -= damage

	//Update defender's fighterLisy entry with new HP
	defender.fighterList[defender.fighter.ID] = defender.fighter

	// Inform players about the attack and damage dealt
	if isSpecialAttack {
		sendMessage(serverConn, attacker.addr, fmt.Sprintf("%s used a special attack!\n", attacker.fighter.Name))
		sendMessage(serverConn, defender.addr, fmt.Sprintf("%s used a special attack!\n", attacker.fighter.Name))
	} else {
		sendMessage(serverConn, attacker.addr, fmt.Sprintf("%s used a normal attack!\n", attacker.fighter.Name))
		sendMessage(serverConn, defender.addr, fmt.Sprintf("%s used a normal attack!\n", attacker.fighter.Name))
	}

	sendMessage(serverConn, attacker.addr, divider)
	sendMessage(serverConn, defender.addr, divider)
	sendMessage(serverConn, attacker.addr, fmt.Sprintf("Damage dealt: %d\n", damage))
	sendMessage(serverConn, attacker.addr, fmt.Sprintf("%s's HP: %d\n", defender.fighter.Name, defender.fighter.HP))
	sendMessage(serverConn, defender.addr, fmt.Sprintf("Damage dealt: %d\n", damage))
	sendMessage(serverConn, defender.addr, fmt.Sprintf("%s's HP: %d\n", defender.fighter.Name, defender.fighter.HP))

	wait(3)
	// Prompt the attacker to switch their fighter
	sendMessage(serverConn, attacker.addr, fmt.Sprintf("%s, do you want to switch your fighter? (Y/N)", attacker.name))
	response := receiveMessage(serverConn)
	if strings.TrimSpace(strings.ToUpper(response)) == "Y" {
		if selectFighter(attacker, serverConn) {
			sendMessage(serverConn, attacker.addr, "You have switched your fighter.")
		} else {
			sendMessage(serverConn, attacker.addr, "Failed to switch fighter. Continue with the current fighter.")
		}
	}
}

func handleLogin(conn *net.UDPConn, addr *net.UDPAddr, message string) {
	mutex.Lock()
	defer mutex.Unlock()

	username := strings.Split(message, ":")[1]
	if _, exists := players[username]; exists {
		// Username already exists, notify the client
		successMessage := "SUCCESS: You have registered as " + username
		sendMessage(conn, addr, successMessage)
		fmt.Println("Success message sent to", addr, ":", successMessage)
		sendMessage(conn, addr, "Welcome to the Pokemon Battle Server!\n")
		go handleClient(conn, addr, username)
	} else {
		// Notify the client that no player found with this name
		errorMessage := "FAILED: no player found " + username
		sendMessage(conn, addr, errorMessage)
		fmt.Println("Error message sent to", addr, ":", errorMessage)
	}
}

func handleLogout(message string) {
	mutex.Lock()
	defer mutex.Unlock()

	username := strings.Split(message, ":")[1]
	delete(players, username)
	fmt.Println("Client", username, "logged out")
}

func loadPlayerData() {
	file, err := os.Open("../../player.json")
	if err != nil {
		fmt.Println("Error opening player data file:", err)
		return
	}
	defer file.Close()

	var playersData []player.Player

	if err := json.NewDecoder(file).Decode(&playersData); err != nil {
		fmt.Println("Error decoding player data:", err)
		return
	}
	for _, pd := range playersData {
		players[pd.Name] = pd
	}

	fmt.Printf("Loaded %d players\n", len(players))
}

func handleMessage(msg string, senderAddr *net.UDPAddr) {
	// Handle any other messages here
}

func main() {
	loadPlayerData()
	//UPD address
	serverAddr, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}

	//UDP listener
	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}

	fmt.Println("Servet started, waiting for player...")

	//Buffer for incoming messages
	buf := make([]byte, 1024)

	for {
		n, addr, _ := serverConn.ReadFromUDP(buf)
		message := string(buf[:n])
		fmt.Println("Receive message from client: ", message)

		if strings.HasPrefix(message, "LOGIN:") {
			handleLogin(serverConn, addr, message)
		} else if strings.HasPrefix(message, "LOGOUT:") {
			handleLogout(message)
		} else {
			mutex.Lock()
			if len(gamers) == 2 && !battleActive {
				battleActive = true
				var gamer1, gamer2 *gamer
				for _, g := range gamers {
					if gamer1 == nil {
						gamer1 = g
					} else {
						gamer2 = g
					}
				}
				mutex.Unlock()

				go startBattle(gamer1, gamer2)

				mutex.Lock()
				gamers = make(map[string]*gamer)
				battleActive = false
				mutex.Unlock()
			} else {
				mutex.Unlock()
			}
			go handleMessage(message, addr)
		}
	}
}
