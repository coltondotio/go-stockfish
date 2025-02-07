package stockfish

import (
	"errors"
	"runtime"
)

type Stockfish interface {
	GetFenEvaluation(fen string) (float64, error)
}

func New() (Stockfish, error) {
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		return &stockfishImpl{}, nil
	}
	if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		return &stockfishImpl{}, nil
	}
	return nil, errors.New("unsupported architecture - only macOS ARM64 (Apple Silicon) and Linux x86-64 are supported")
}
