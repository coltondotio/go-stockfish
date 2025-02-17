package main

import (
	"fmt"
	"log"

	"github.com/coltondotio/go-stockfish/stockfish"
)

func main() {
	// Initialize stockfish
	sf, err := stockfish.New(stockfish.Options{
		Debug: false,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = sf.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer sf.Close()

	// Example FEN string representing a chess position
	fen := "r1bqkbnr/ppp1nppp/3p4/3Pp3/4P3/5N2/PPP1BPPP/RNBQK2R b KQkq - 2 5"

	// Get evaluation
	eval, err := sf.GetFenPosition(10, fen)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Position evaluation: %.2f\n", eval)
}
