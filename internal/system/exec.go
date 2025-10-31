package system

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// CommandResult holds the result of a command execution
type CommandResult struct {
	ExitCode int
	Output   string
	Error    error
}

// RunConfig configures how a command should be executed
type RunConfig struct {
	Command     string
	Args        []string
	StreamToLog bool
	LogPath     string
	WorkingDir  string
	Env         []string
}

// Run executes a command with the given configuration
func Run(cfg RunConfig) CommandResult {
	cmd := exec.Command(cfg.Command, cfg.Args...)

	if cfg.WorkingDir != "" {
		cmd.Dir = cfg.WorkingDir
	}

	if len(cfg.Env) > 0 {
		cmd.Env = append(os.Environ(), cfg.Env...)
	}

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return CommandResult{
			ExitCode: -1,
			Error:    fmt.Errorf("failed to create stdout pipe: %w", err),
		}
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return CommandResult{
			ExitCode: -1,
			Error:    fmt.Errorf("failed to create stderr pipe: %w", err),
		}
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return CommandResult{
			ExitCode: -1,
			Error:    fmt.Errorf("failed to start command: %w", err),
		}
	}

	var wg sync.WaitGroup
	var output strings.Builder
	var logFile *os.File

	// Open log file if streaming is enabled
	if cfg.StreamToLog && cfg.LogPath != "" {
		logFile, err = os.OpenFile(cfg.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// Continue without logging if we can't open the file
			fmt.Fprintf(os.Stderr, "Warning: failed to open log file: %v\n", err)
		} else {
			defer logFile.Close()
		}
	}

	// Process stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			output.WriteString(line + "\n")
			if logFile != nil {
				fmt.Fprintln(logFile, line)
				logFile.Sync() // Flush to disk immediately
			}
		}
	}()

	// Process stderr
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			output.WriteString(line + "\n")
			if logFile != nil {
				fmt.Fprintln(logFile, line)
				logFile.Sync() // Flush to disk immediately
			}
		}
	}()

	// Wait for output processing to complete
	wg.Wait()

	// Wait for command to finish
	err = cmd.Wait()

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return CommandResult{
		ExitCode: exitCode,
		Output:   output.String(),
		Error:    err,
	}
}

// RunSimple is a convenience function for simple command execution
func RunSimple(command string, args ...string) CommandResult {
	return Run(RunConfig{
		Command: command,
		Args:    args,
	})
}

// RunLogged runs a command and streams output to the log file
func RunLogged(logPath, command string, args ...string) CommandResult {
	return Run(RunConfig{
		Command:     command,
		Args:        args,
		StreamToLog: true,
		LogPath:     logPath,
	})
}

// RunWithOutput runs a command and captures output in real-time via a callback
func RunWithOutput(command string, args []string, callback func(line string)) error {
	cmd := exec.Command(command, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	var wg sync.WaitGroup

	// Process stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		processOutput(stdout, callback)
	}()

	// Process stderr
	wg.Add(1)
	go func() {
		defer wg.Done()
		processOutput(stderr, callback)
	}()

	wg.Wait()
	return cmd.Wait()
}

func processOutput(reader io.Reader, callback func(string)) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		callback(scanner.Text())
	}
}
