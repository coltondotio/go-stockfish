package stockfish

import (
	"testing"
	"time"
)

func TestGetFenPosition(t *testing.T) {
	sf, err := New(Options{})
	if err != nil {
		t.Fatalf("Failed to create Stockfish instance: %v", err)
	}
	err = sf.Start()
	if err != nil {
		t.Fatalf("Failed to start Stockfish: %v", err)
	}

	defer sf.Close()

	fen := "rnbqk1nr/1ppp1ppp/p7/4p2Q/2B1P3/b7/PPPP1PPP/RNB1K1NR w KQkq - 0 4"

	pos, err := sf.GetFenPosition(15, fen)
	if err != nil {
		t.Fatalf("Failed to get position: %v", err)
	}

	if !pos.IsMateScore {
		t.Error("Expected mate score, got centipawn score")
	}

	if pos.MateScore != 1 {
		t.Errorf("Expected mate in 1, got mate in %d", pos.MateScore)
	}
}

func TestGetFenPositionBenchmark(t *testing.T) {
	sf, err := New(Options{})
	if err != nil {
		t.Fatalf("Failed to create Stockfish instance: %v", err)
	}
	err = sf.Start()
	if err != nil {
		t.Fatalf("Failed to start Stockfish: %v", err)
	}
	defer sf.Close()

	fens := []string{
		"rnbqk1nr/ppp2ppp/4p3/3p4/1bPP4/2N5/PP2PPPP/R1BQKBNR w KQkq - 2 4",
		"rnbqk1nr/ppp2ppp/4p3/3p4/1bPP4/P1N5/1P2PPPP/R1BQKBNR b KQkq - 0 4",
		"rnbqk1nr/ppp2ppp/4p3/3p4/2PP4/P1b5/1P2PPPP/R1BQKBNR w KQkq - 0 5",
		"rnbqk1nr/ppp2ppp/4p3/3p4/2PP4/P1P5/4PPPP/R1BQKBNR b KQkq - 0 5",
		"rnbqk2r/ppp2ppp/4pn2/3p4/2PP4/P1P5/4PPPP/R1BQKBNR w KQkq - 1 6",
		"rnbqk2r/ppp2ppp/4pn2/3p2B1/2PP4/P1P5/4PPPP/R2QKBNR b KQkq - 2 6",
		"rnbqk2r/ppp2ppp/4pn2/6B1/2pP4/P1P5/4PPPP/R2QKBNR w KQkq - 0 7",
		"rnbqk2r/ppp2ppp/4pn2/6B1/2pP4/P1P1P3/5PPP/R2QKBNR b KQkq - 0 7",
		"rnb1k2r/ppp2ppp/4pn2/3q2B1/2pP4/P1P1P3/5PPP/R2QKBNR w KQkq - 1 8",
		"rnb1k2r/ppp2ppp/4pB2/3q4/2pP4/P1P1P3/5PPP/R2QKBNR b KQkq - 0 8",
		"rnb1k2r/ppp2p1p/4pp2/3q4/2pP4/P1P1P3/5PPP/R2QKBNR w KQkq - 0 9",
		"rnb1k2r/ppp2p1p/4pp2/3q4/2pP4/P1P1P3/4NPPP/R2QKB1R b KQkq - 1 9",
		"rnb1k1r1/ppp2p1p/4pp2/3q4/2pP4/P1P1P3/4NPPP/R2QKB1R w KQq - 2 10",
		"rnb1k1r1/ppp2p1p/4pp2/3q4/2pP1N2/P1P1P3/5PPP/R2QKB1R b KQq - 3 10",
		"rnb1k1r1/ppp2p1p/4pp2/q7/2pP1N2/P1P1P3/5PPP/R2QKB1R w KQq - 4 11",
		"rnb1k1r1/ppp2p1p/4pp2/q7/2pP1N2/P1P1P3/2Q2PPP/R3KB1R b KQq - 5 11",
		"rn2k1r1/pppb1p1p/4pp2/q7/2pP1N2/P1P1P3/2Q2PPP/R3KB1R w KQq - 6 12",
		"rn2k1r1/pppb1p1p/4pp2/q7/2BP1N2/P1P1P3/2Q2PPP/R3K2R b KQq - 0 12",
		"rn2k1r1/ppp2p1p/2b1pp2/q7/2BP1N2/P1P1P3/2Q2PPP/R3K2R w KQq - 1 13",
		"rn2k1r1/ppp2p1p/2b1pp2/q7/2BP1N2/P1P1P3/2Q2PPP/R4RK1 b q - 2 13",
		"rn2k1r1/ppp2p1p/2b2p2/q3p3/2BP1N2/P1P1P3/2Q2PPP/R4RK1 w q - 0 14",
		"rn2k1r1/ppp2p1p/2b2p2/q3P3/2B2N2/P1P1P3/2Q2PPP/R4RK1 b q - 0 14",
		"rn2k1r1/ppp2p1p/2b5/q3p3/2B2N2/P1P1P3/2Q2PPP/R4RK1 w q - 0 15",
		"rn2k1r1/ppp2p1Q/2b5/q3p3/2B2N2/P1P1P3/5PPP/R4RK1 b q - 0 15",
		"rn2kr2/ppp2p1Q/2b5/q3p3/2B2N2/P1P1P3/5PPP/R4RK1 w q - 1 16",
		"rn2kr2/ppp2p1Q/2b3N1/q3p3/2B5/P1P1P3/5PPP/R4RK1 b q - 2 16",
		"rn2kr2/ppp2p1Q/6N1/q3p3/2B1b3/P1P1P3/5PPP/R4RK1 w q - 3 17",
		"rn2kr2/ppp2B1Q/6N1/q3p3/4b3/P1P1P3/5PPP/R4RK1 b q - 0 17",
		"rn2k3/ppp2r1Q/6N1/q3p3/4b3/P1P1P3/5PPP/R4RK1 w q - 0 18",
		"rn2k1Q1/ppp2r2/6N1/q3p3/4b3/P1P1P3/5PPP/R4RK1 b q - 1 18",
		"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 ",
		"rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2",
		"rnbqkbnr/pppp1ppp/8/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2",
		"rnbqk1nr/pppp1ppp/3b4/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3",
		"rnbqk1nr/pppp1ppp/3b4/4p3/3PP3/5N2/PPP2PPP/RNBQKB1R b KQkq d3 0 3",
		"rnbqk1nr/pppp1ppp/3b4/8/3pP3/5N2/PPP2PPP/RNBQKB1R w KQkq - 0 4",
		"rnbqk1nr/pppp1ppp/3b4/8/3NP3/8/PPP2PPP/RNBQKB1R b KQkq - 0 4",
		"rnbqk1nr/pp1p1ppp/3b4/2p5/3NP3/8/PPP2PPP/RNBQKB1R w KQkq c6 0 ",
		"rnbqk1nr/pp1p1ppp/3b4/2p2N2/4P3/8/PPP2PPP/RNBQKB1R b KQkq - 1 5",
		"rnbqk1nr/pp1p1ppp/8/2p1bN2/4P3/8/PPP2PPP/RNBQKB1R w KQkq - 2 6",
		"rnbqk1nr/pp1p1ppp/8/2p1bN2/4P3/2N5/PPP2PPP/R1BQKB1R b KQkq - 3 6",
		"rnbqk1nr/pp1p1p1p/6p1/2p1bN2/4P3/2N5/PPP2PPP/R1BQKB1R w KQkq - 0 7",
		"rnbqk1nr/pp1p1p1p/3N2p1/2p1b3/4P3/2N5/PPP2PPP/R1BQKB1R b KQkq - 1 7",
		"rnbqk1nr/pp1p1p1p/3b2p1/2p5/4P3/2N5/PPP2PPP/R1BQKB1R w KQkq - 0 8",
		"rnbqk1nr/pp1p1p1p/3Q2p1/2p5/4P3/2N5/PPP2PPP/R1B1KB1R b KQkq - 0 8",
	}

	start := time.Now()

	for i, fen := range fens {
		_, err := sf.GetFenPosition(20, fen)
		if err != nil {
			t.Fatalf("Failed to get position on iteration %d: %v", i, err)
		}
	}

	elapsed := time.Since(start)
	t.Logf("Time taken for 50 position evaluations: %v", elapsed)
}

func TestBadPosition(t *testing.T) {
	sf, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}

	err = sf.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer sf.Close()

	fen := "rnbqkbnr/pppp2pp/4pp2/8/1Q6/3P4/PPP1PPPP/RNB1KBNR b KQkq - 1 3"
	pos, err := sf.GetFenPosition(10, fen)
	if err != nil {
		t.Fatal(err)
	}

	if !pos.IsCentipawnScore {
		t.Fatal("Expected centipawn score")
	}

	if pos.CentipawnScore > -500 {
		t.Errorf("Expected position evaluation to be less than -500, got %d", pos.CentipawnScore)
	}
}

func TestMateInOne(t *testing.T) {
	sf, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}

	err = sf.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer sf.Close()

	fen := "rnbqkbnr/2pppppp/1p6/p6Q/2B1P3/8/PPPP1PPP/RNB1K1NR w KQkq - 0 4"
	pos, err := sf.GetFenPosition(10, fen)
	if err != nil {
		t.Fatal(err)
	}

	if !pos.IsMateScore {
		t.Fatal("Expected mate score")
	}

	if pos.MateScore != 1 {
		t.Errorf("Expected mate in 1, got mate in %d", pos.MateScore)
	}
}

func TestCheckmate(t *testing.T) {
	sf, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}

	err = sf.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer sf.Close()

	fen := "rnbqkbnr/2pppQpp/1p6/p7/2B1P3/8/PPPP1PPP/RNB1K1NR b KQkq - 0 4"
	pos, err := sf.GetFenPosition(10, fen)
	if err != nil {
		t.Fatal(err)
	}

	if !pos.IsMateScore {
		t.Fatal("Expected mate score")
	}

	if pos.MateScore != 0 {
		t.Errorf("Expected mate in 0 (checkmate), got mate in %d", pos.MateScore)
	}
}

func TestBop(t *testing.T) {
	sf, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}

	err = sf.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer sf.Close()

	fen := "rn2k1Q1/ppp2r2/6N1/q3p3/4b3/P1P1P3/5PPP/R4RK1 b q - 1 18"

	for i := 0; i < 100; i++ {
		_, err := sf.GetFenPosition(10, fen)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestDraw(t *testing.T) {
	sf, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}

	err = sf.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer sf.Close()

	fen := "8/5K2/8/8/8/8/2Q5/k7 b - - 32 74"
	pos, err := sf.GetFenPosition(10, fen)
	if err != nil {
		t.Fatal(err)
	}

	if pos.IsMateScore {
		t.Error("Expected non-mate score for drawn position")
	}

	if !pos.IsCentipawnScore {
		t.Error("Expected centipawn score for drawn position")
	}

	if pos.CentipawnScore != 0 {
		t.Errorf("Expected centipawn score of 0 for drawn position, got %d", pos.CentipawnScore)
	}
}
