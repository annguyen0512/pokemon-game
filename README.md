# Pokémon Game Simulation in Go

## 🌟 Introduction

This project simulates a simple Pokémon game built using **Go (Golang)**. Players can move around a map, catch randomly spawned Pokémon, and manage their Pokémon list. The game emphasizes player interaction and Pokémon management.

---

## ⚙️ Key Features

- **Player Management**: Register new players, update positions, and manage their Pokémon list.
- **Pokémon System**: Randomly spawn Pokémon with various attributes (speed, level, stats).
- **Catch Pokémon**: Players can capture Pokémon appearing on the map.
- **Data Persistence**: Player and Pokémon data are stored in JSON files.
- **Concurrency Handling**: Uses `sync.Mutex` to ensure data consistency in a multithreaded environment.

---

## 🛠️ Technologies Used

- **Language**: Go (Golang)
- **Data Storage**: JSON
- **Built-in Libraries**:
  - `sync` (thread synchronization)
  - `encoding/json` (read/write JSON files)
  - `time` (time management)
