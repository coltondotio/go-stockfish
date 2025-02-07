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
	"strconv"
	"strings"
	"sync"

	"github.com/coltondotio/go-stockfish/stockfish/internal/resources"
)

type stockfishImpl struct {
	binaryPath string
	initOnce   sync.Once
	initErr    error
}

func (s *stockfishImpl) initBinary() error {
	// Try to get cache directory from user's home first
	var cacheDir string
	homeDir, err := os.UserHomeDir()
	if err == nil {
		cacheDir = filepath.Join(homeDir, ".cache", "stockfish-binary")
	} else {
		// Fall back to temp directory if home directory is not available
		tempDir := os.TempDir()
		cacheDir = filepath.Join(tempDir, "stockfish-binary")
	}

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

func (s *stockfishImpl) GetFenPosition(depth int, fen string) (Position, error) {
	s.initOnce.Do(func() {
		s.initErr = s.initBinary()
	})

	if s.initErr != nil {
		return Position{}, fmt.Errorf("failed to initialize stockfish binary: %w", s.initErr)
	}

	cmd := exec.Command(s.binaryPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return Position{}, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Position{}, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return Position{}, fmt.Errorf("failed to start stockfish: %w", err)
	}
	defer cmd.Process.Kill()

	// Initialize UCI and wait for ready
	if _, err := io.WriteString(stdin, "uci\nisready\n"); err != nil {
		return Position{}, fmt.Errorf("failed to write UCI init commands: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if scanner.Text() == "readyok" {
			break
		}
	}

	// Send position and analysis commands
	input := fmt.Sprintf("position fen %s\ngo depth %d\n", fen, depth)
	if _, err := io.WriteString(stdin, input); err != nil {
		return Position{}, fmt.Errorf("failed to write position commands: %w", err)
	}

	var lastInfoLine string
	var seenInfoDepth bool
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "info depth") {
			lastInfoLine = line
			seenInfoDepth = true
		} else if seenInfoDepth {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return Position{}, fmt.Errorf("error reading output: %w", err)
	}
	if lastInfoLine == "" {
		return Position{}, errors.New("no evaluation found in stockfish output")
	}

	// Parse the score from the last info line
	cpRegex := regexp.MustCompile(`score cp (-?\d+)`)
	mateRegex := regexp.MustCompile(`score mate (-?\d+)`)

	if cpMatches := cpRegex.FindStringSubmatch(lastInfoLine); cpMatches != nil {
		score, err := strconv.ParseInt(cpMatches[1], 10, 32)
		if err != nil {
			return Position{}, fmt.Errorf("failed to parse centipawn score: %w", err)
		}
		return Position{
			IsCentipawnScore: true,
			CentipawnScore:   int(score),
		}, nil
	}

	if mateMatches := mateRegex.FindStringSubmatch(lastInfoLine); mateMatches != nil {
		score, err := strconv.ParseInt(mateMatches[1], 10, 32)
		if err != nil {
			return Position{}, fmt.Errorf("failed to parse mate score: %w", err)
		}
		return Position{
			IsMateScore: true,
			MateScore:   int(score),
		}, nil
	}

	return Position{}, errors.New("could not parse score from stockfish output")
}
