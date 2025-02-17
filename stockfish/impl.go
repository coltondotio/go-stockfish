// code generated mostly by claude-3.5-sonnet

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
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	scanner    *bufio.Scanner
	mutex      sync.Mutex
	debug      bool
}

func (s *stockfishImpl) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.initBinary(); err != nil {
		return fmt.Errorf("failed to initialize binary: %w", err)
	}

	cmd := exec.Command(s.binaryPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start stockfish: %w", err)
	}

	s.cmd = cmd
	s.stdin = stdin
	s.stdout = stdout
	s.scanner = bufio.NewScanner(stdout)

	// Initialize UCI and wait for ready
	if _, err := io.WriteString(stdin, "uci\nisready\n"); err != nil {
		return fmt.Errorf("failed to write UCI init commands: %w", err)
	}
	for s.scanner.Scan() {
		if s.scanner.Text() == "readyok" {
			break
		}
	}

	// Set number of threads to number of CPUs
	numCPU := runtime.NumCPU()
	if s.debug {
		fmt.Printf("Using %d CPU threads\n", numCPU)
	}
	if _, err := io.WriteString(stdin, fmt.Sprintf("setoption name Threads value %d\n", numCPU)); err != nil {
		return fmt.Errorf("failed to set thread count: %w", err)
	}

	return nil
}

func (s *stockfishImpl) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.cmd != nil && s.cmd.Process != nil {
		if err := s.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill stockfish process: %w", err)
		}
		s.cmd = nil
		s.stdin = nil
		s.stdout = nil
		s.scanner = nil
	}
	return nil
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

func (s *stockfishImpl) parsePosition(lastInfoLine string) (Position, error) {
	// Parse the score from the info line
	cpRegex := regexp.MustCompile(`score cp (-?\d+)`)
	mateRegex := regexp.MustCompile(`score mate (-?\d+)`)

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

	return Position{}, errors.New("could not parse score from stockfish output")
}

func (s *stockfishImpl) flushState() error {
	if _, err := io.WriteString(s.stdin, "ucinewgame\nisready\n"); err != nil {
		return fmt.Errorf("failed to write flush commands: %w", err)
	}

	for s.scanner.Scan() {
		if s.scanner.Text() == "readyok" {
			break
		}
	}
	return nil
}

func (s *stockfishImpl) GetFenPosition(depth int, fen string) (Position, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.cmd == nil {
		return Position{}, errors.New("stockfish engine not started")
	}

	// Get the side to move from FEN
	fenParts := strings.Fields(fen)
	if len(fenParts) < 2 {
		return Position{}, errors.New("invalid FEN string")
	}
	isBlackToMove := fenParts[1] == "b"

	// Flush existing state
	if err := s.flushState(); err != nil {
		return Position{}, err
	}

	// Send position and analysis commands
	input := fmt.Sprintf("position fen %s\ngo depth %d\n", fen, depth)
	if _, err := io.WriteString(s.stdin, input); err != nil {
		return Position{}, fmt.Errorf("failed to write position commands: %w", err)
	}

	var lastInfoLine string
	var seenInfoDepth bool
	var initialLines []string
	for s.scanner.Scan() {
		line := s.scanner.Text()
		initialLines = append(initialLines, line)
		if strings.HasPrefix(line, "info depth") {
			lastInfoLine = line
			seenInfoDepth = true
		} else if seenInfoDepth && strings.HasPrefix(line, "bestmove") {
			break
		}
	}

	if s.debug {
		fmt.Printf("Debug: initial position analysis:\n%s\n", strings.Join(initialLines, "\n"))
	}

	if err := s.scanner.Err(); err != nil {
		return Position{}, fmt.Errorf("error reading output: %w", err)
	}
	if lastInfoLine == "" {
		return Position{}, errors.New("no evaluation found in stockfish output")
	}

	// First get the initial position evaluation
	pos, err := s.parsePosition(lastInfoLine)
	if err != nil {
		return Position{}, err
	}

	// Negate scores if it's Black's turn
	if isBlackToMove {
		if pos.IsMateScore {
			pos.MateScore = -pos.MateScore
		}
		if pos.IsCentipawnScore {
			pos.CentipawnScore = -pos.CentipawnScore
		}
	}

	return pos, nil
}
