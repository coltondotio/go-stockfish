package stockfish

import (
	"errors"
	"runtime"
)

type Stockfish interface {
	Start() error
	Close() error
	GetFenPosition(depth int, fen string) (Position, error)
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

type Position struct {
	IsMateScore      bool
	MateScore        int
	IsCentipawnScore bool
	CentipawnScore   int
}
