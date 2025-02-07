package stockfish

import (
	"math"
	"testing"
)

func TestGetFenEvaluation(t *testing.T) {
	sf, err := New()
	if err != nil {
		t.Fatalf("Failed to create Stockfish instance: %v", err)
	}

	fen := "r1bqkbnr/ppp1nppp/3p4/3Pp3/4P3/5N2/PPP1BPPP/RNBQK2R b KQkq - 2 5"
	expectedEval := 0.35

	eval, err := sf.GetFenEvaluation(fen)
	if err != nil {
		t.Fatalf("Failed to get evaluation: %v", err)
	}

	if math.Abs(eval-expectedEval) > 0.1 {
		t.Errorf("Evaluation %f not within 0.1 of expected %f", eval, expectedEval)
	}
}
