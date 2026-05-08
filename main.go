package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"key_value_engine-nasp/config"
	"key_value_engine-nasp/engine"
	"key_value_engine-nasp/model"
)

func printMenu() {
	fmt.Println("\nDostupne komande:")
	fmt.Println("  PUT <kljuc> <vrednost>  — dodaje/azurira zapis")
	fmt.Println("  GET <kljuc>             — cita vrednost za kljuc")
	fmt.Println("  DELETE <kljuc>          — brise zapis")
	fmt.Println("  QUIT                    — zatvara engine i izlazi")
	fmt.Println()
}

func main() {
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		cfg = config.DefaultConfig()
	}

	eng, err := engine.NewEngine(cfg)
	if err != nil {
		fmt.Printf("Greska pri pokretanju engine-a: %v\n", err)
		os.Exit(1)
	}

	printMenu()

	defer func() {
		if err := eng.Close(); err != nil {
			fmt.Printf("Greska pri zatvaranju engine-a: %v\n", err)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := splitCommand(line)
		if len(parts) == 0 {
			continue
		}

		command := strings.ToUpper(parts[0])

		switch command {
		case "PUT":
			handlePut(eng, parts)
			printMenu()
		case "GET":
			handleGet(eng, parts)
			printMenu()
		case "DELETE":
			handleDelete(eng, parts)
			printMenu()
		case "QUIT", "EXIT":
			fmt.Println("Dovidjenja!")
			return
		default:
			fmt.Printf("Nepoznata komanda: %s\n", command)
			printMenu()
		}
	}
}

func handlePut(eng *engine.Engine, parts []string) {
	if len(parts) < 3 {
		fmt.Println("Upotreba: PUT <kljuc> <vrednost>")
		return
	}

	key := parts[1]
	value := strings.Join(parts[2:], " ")

	err := eng.Put(key, []byte(value))
	if err != nil {
		if err == model.ErrRateLimited {
			fmt.Println("Greska: prekoracen broj dozvoljenih zahteva. Sacekajte.")
		} else {
			fmt.Printf("Greska: %v\n", err)
		}
		return
	}

	fmt.Printf("OK — zapisano: %s = %s\n", key, value)
}

func handleGet(eng *engine.Engine, parts []string) {
	if len(parts) < 2 {
		fmt.Println("Upotreba: GET <kljuc>")
		return
	}

	key := parts[1]

	value, err := eng.Get(key)
	if err != nil {
		switch err {
		case model.ErrKeyNotFound:
			fmt.Printf("Kljuc '%s' nije pronadjen.\n", key)
		case model.ErrDeleted:
			fmt.Printf("Kljuc '%s' nije pronadjen.\n", key)
		case model.ErrRateLimited:
			fmt.Println("Greska: prekoracen broj dozvoljenih zahteva. Sacekajte.")
		default:
			fmt.Printf("Greska: %v\n", err)
		}
		return
	}

	fmt.Printf("%s = %s\n", key, string(value))
}

func handleDelete(eng *engine.Engine, parts []string) {
	if len(parts) < 2 {
		fmt.Println("Upotreba: DELETE <kljuc>")
		return
	}

	key := parts[1]

	err := eng.Delete(key)
	if err != nil {
		switch err {
		case model.ErrKeyNotFound:
			fmt.Printf("Kljuc '%s' nije pronadjen.\n", key)
		case model.ErrRateLimited:
			fmt.Println("Greska: prekoracen broj dozvoljenih zahteva. Sacekajte.")
		default:
			fmt.Printf("Greska: %v\n", err)
		}
		return
	}

	fmt.Printf("OK — obrisan kljuc: %s\n", key)
}

func splitCommand(line string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false

	for _, ch := range line {
		switch {
		case ch == '"':
			inQuotes = !inQuotes
		case ch == ' ' && !inQuotes:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}
