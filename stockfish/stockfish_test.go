package stockfish

import (
	"testing"
)

func TestGetFenPosition(t *testing.T) {
	sf, err := New()
	if err != nil {
		t.Fatalf("Failed to create Stockfish instance: %v", err)
	}

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
