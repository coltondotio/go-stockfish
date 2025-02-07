package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/coltondotio/go-stockfish/stockfish"
)

func main() {
	depth := flag.Int("depth", 10, "search depth")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "Usage: go-stockfish [-depth N] FEN")
		os.Exit(1)
	}

	fen := flag.Arg(0)

	sf, err := stockfish.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Stockfish: %v\n", err)
		os.Exit(1)
	}

	err = sf.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting Stockfish: %v\n", err)
		os.Exit(1)
	}
	defer sf.Close()

	pos, err := sf.GetFenPosition(*depth, fen)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error analyzing position: %v\n", err)
		os.Exit(1)
	}

	json, err := json.MarshalIndent(pos, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(json))
}
