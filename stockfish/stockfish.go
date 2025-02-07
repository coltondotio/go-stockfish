package stockfish

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"

	"github.com/coltondotio/go-stockfish/stockfish/internal/resources" // adjust import path as needed
)

type Stockfish struct {
	binaryPath string
	initOnce   sync.Once
	initErr    error
}

func New() (*Stockfish, error) {
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		return &Stockfish{}, nil
	}
	if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		return &Stockfish{}, nil
	}
	return nil, errors.New("unsupported architecture - only macOS ARM64 (Apple Silicon) and Linux x86-64 are supported")
}

func (s *Stockfish) initBinary() error {
	// Create cache directory in user's home
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".cache", "stockfish-binary")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	var binaryName string
	if runtime.GOOS == "darwin" {
		binaryName = "stockfish-macos-m1-apple-silicon"
	} else {
		binaryName = "stockfish-ubuntu-x86-64-avx2"
	}

	binaryPath := filepath.Join(cacheDir, binaryName)

	// Check if binary exists in cache
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// Extract binary from embedded resources
		binary, err := resources.Resources.ReadFile(binaryName)
		if err != nil {
			return fmt.Errorf("failed to read embedded binary: %w", err)
		}

		// Write to cache
		if err := os.WriteFile(binaryPath, binary, 0755); err != nil {
			return fmt.Errorf("failed to write binary to cache: %w", err)
		}
	}

	s.binaryPath = binaryPath
	return nil
}

func (s *Stockfish) GetFenEvaluation(fen string) (float64, error) {
	s.initOnce.Do(func() {
		s.initErr = s.initBinary()
	})

	if s.initErr != nil {
		return 0, fmt.Errorf("failed to initialize stockfish binary: %w", s.initErr)
	}

	cmd := exec.Command(s.binaryPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return 0, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start stockfish: %w", err)
	}

	// Send commands to stockfish
	input := fmt.Sprintf("position fen %s\ngo depth 8\neval\n", fen)
	if _, err := io.WriteString(stdin, input); err != nil {
		return 0, fmt.Errorf("failed to write to stdin: %w", err)
	}
	stdin.Close()

	// Read output and look for evaluation line
	scanner := bufio.NewScanner(stdout)
	evalRegex := regexp.MustCompile(`Final evaluation\s+([+-]?\d+\.\d+)`)

	for scanner.Scan() {
		line := scanner.Text()
		if matches := evalRegex.FindStringSubmatch(line); matches != nil {
			cmd.Process.Kill() // Clean up the process

			// Parse the float value
			var value float64
			_, err := fmt.Sscanf(matches[1], "%f", &value)
			if err != nil {
				return 0, fmt.Errorf("failed to parse evaluation value: %w", err)
			}
			return value, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("error reading stockfish output: %w", err)
	}

	return 0, errors.New("evaluation not found in stockfish output")
}
