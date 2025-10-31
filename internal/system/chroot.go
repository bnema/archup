package system

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ChrootExec executes a command inside a chroot environment
func ChrootExec(logPath, mountPoint, command string, args ...string) error {
	// Construct the command to run in chroot
	fullCommand := command
	if len(args) > 0 {
		fullCommand = command + " " + strings.Join(args, " ")
	}

	result := RunLogged(logPath, "arch-chroot", mountPoint, "bash", "-c", fullCommand)
	if result.Error != nil {
		return fmt.Errorf("failed to execute in chroot: %w", result.Error)
	}
	return nil
}

// ChrootExecWithOutput executes a command inside a chroot environment and captures output
func ChrootExecWithOutput(logPath, mountPoint, command string, args ...string) (string, error) {
	// Construct the command to run in chroot
	fullCommand := command
	if len(args) > 0 {
		fullCommand = command + " " + strings.Join(args, " ")
	}

	result := RunLogged(logPath, "arch-chroot", mountPoint, "bash", "-c", fullCommand)
	if result.Error != nil {
		return "", fmt.Errorf("failed to execute in chroot: %w", result.Error)
	}
	return result.Output, nil
}

// ChrootExecWithOutputAndContext executes a command inside chroot with context and captures output
func ChrootExecWithOutputAndContext(ctx context.Context, logPath, mountPoint, command string, args ...string) (string, error) {
	// Construct the command to run in chroot
	fullCommand := command
	if len(args) > 0 {
		fullCommand = command + " " + strings.Join(args, " ")
	}

	cmd := exec.CommandContext(ctx, "arch-chroot", mountPoint, "bash", "-c", fullCommand)

	// Capture output
	output, err := cmd.CombinedOutput()

	// Also log to file
	if logPath != "" {
		logFile, logErr := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if logErr == nil {
			defer logFile.Close()
			logFile.Write(output)
		}
	}

	if err != nil {
		// Check if it was a context deadline exceeded
		if ctx.Err() == context.DeadlineExceeded {
			return string(output), fmt.Errorf("command timed out: %w", err)
		}
		return string(output), fmt.Errorf("failed to execute in chroot: %w", err)
	}

	return string(output), nil
}

// ChrootExecWithContext executes a command inside chroot with timeout support via context
// The context deadline will terminate the command if it takes too long
func ChrootExecWithContext(ctx context.Context, logPath, mountPoint, command string) error {
	cmd := exec.CommandContext(ctx, "arch-chroot", mountPoint, "bash", "-c", command)

	// Set up logging
	if logPath != "" {
		logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			defer logFile.Close()
			cmd.Stdout = logFile
			cmd.Stderr = logFile
		}
	}

	// Execute the command
	if err := cmd.Run(); err != nil {
		// Check if it was a context deadline exceeded
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("command timed out: %w", err)
		}
		return fmt.Errorf("failed to execute in chroot: %w", err)
	}

	return nil
}

// ChrootExecWithStdin executes a command inside chroot with stdin input
// This is secure for passing sensitive data like passwords without exposing them in process listings
func ChrootExecWithStdin(logPath, mountPoint, command, stdin string) error {
	cmd := exec.Command("arch-chroot", mountPoint, "bash", "-c", command)

	// Set up stdin pipe
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Set up logging
	if logPath != "" {
		logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			defer logFile.Close()
			cmd.Stdout = logFile
			cmd.Stderr = logFile
		}
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Write to stdin and close pipe
	if _, err := stdinPipe.Write([]byte(stdin)); err != nil {
		return fmt.Errorf("failed to write to stdin: %w", err)
	}
	stdinPipe.Close()

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to execute in chroot: %w", err)
	}

	return nil
}

// ChrootPacman runs pacman inside the chroot
func ChrootPacman(logPath, mountPoint, operation string, packages ...string) error {
	args := append([]string{operation, "--noconfirm"}, packages...)
	command := "pacman " + strings.Join(args, " ")

	result := RunLogged(logPath, "arch-chroot", mountPoint, "bash", "-c", command)
	if result.Error != nil {
		return fmt.Errorf("failed to run pacman in chroot: %w", result.Error)
	}
	return nil
}

// ChrootSystemctl runs systemctl inside the chroot
func ChrootSystemctl(logPath, mountPoint, action, service string) error {
	command := fmt.Sprintf("systemctl %s %s", action, service)

	result := RunLogged(logPath, "arch-chroot", mountPoint, "bash", "-c", command)
	if result.Error != nil {
		return fmt.Errorf("failed to run systemctl in chroot: %w", result.Error)
	}
	return nil
}

// Mount mounts a filesystem
func Mount(logPath, device, mountPoint string, options ...string) error {
	args := []string{device, mountPoint}
	if len(options) > 0 {
		args = append([]string{"-o", strings.Join(options, ",")}, args...)
	}

	result := RunLogged(logPath, "mount", args...)
	if result.Error != nil {
		return fmt.Errorf("failed to mount %s: %w", device, result.Error)
	}
	return nil
}

// Unmount unmounts a filesystem
func Unmount(logPath, mountPoint string) error {
	result := RunLogged(logPath, "umount", "-R", mountPoint)
	if result.Error != nil {
		return fmt.Errorf("failed to unmount %s: %w", mountPoint, result.Error)
	}
	return nil
}

// Genfstab generates fstab for the installed system
func Genfstab(logPath, mountPoint, fstabPath string) error {
	result := RunLogged(logPath, "genfstab", "-U", mountPoint)
	if result.Error != nil {
		return fmt.Errorf("failed to generate fstab: %w", result.Error)
	}

	// TODO: Write result.Output to fstabPath
	// This should be done using file operations
	return nil
}

// DefaultChrootSession represents a persistent chroot session
type DefaultChrootSession struct {
	cmd         *exec.Cmd
	stdin       *os.File
	stdoutPipe  *os.File
	stderrPipe  *os.File
	logFile     *os.File
	logPath     string
	mountPoint  string
	ctx         context.Context
	cancel      context.CancelFunc
	outputChan  chan string
	errorChan   chan error
	closed      bool
}

// BeginSession starts a persistent chroot session
func BeginSession(logPath, mountPoint string) (*DefaultChrootSession, error) {
	return BeginSessionWithContext(context.Background(), logPath, mountPoint)
}

// BeginSessionWithContext starts a persistent chroot session with context support
func BeginSessionWithContext(ctx context.Context, logPath, mountPoint string) (*DefaultChrootSession, error) {
	// Create a cancellable context
	sessionCtx, cancel := context.WithCancel(ctx)

	// Start arch-chroot with bash
	cmd := exec.CommandContext(sessionCtx, "arch-chroot", mountPoint, "bash", "--norc", "--noprofile")

	// Create pipes for stdin, stdout, stderr
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Open log file
	var logFile *os.File
	if logPath != "" {
		logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		if logFile != nil {
			logFile.Close()
		}
		cancel()
		return nil, fmt.Errorf("failed to start chroot session: %w", err)
	}

	session := &DefaultChrootSession{
		cmd:        cmd,
		stdin:      stdinPipe.(*os.File),
		stdoutPipe: stdoutPipe.(*os.File),
		stderrPipe: stderrPipe.(*os.File),
		logFile:    logFile,
		logPath:    logPath,
		mountPoint: mountPoint,
		ctx:        sessionCtx,
		cancel:     cancel,
		outputChan: make(chan string, 100),
		errorChan:  make(chan error, 1),
		closed:     false,
	}

	// Start goroutines to handle output
	go session.handleOutput()
	go session.handleError()

	return session, nil
}

// handleOutput continuously reads from stdout and logs it
func (s *DefaultChrootSession) handleOutput() {
	buf := make([]byte, 4096)
	for {
		n, err := s.stdoutPipe.Read(buf)
		if n > 0 {
			data := string(buf[:n])
			if s.logFile != nil {
				s.logFile.WriteString(data)
				s.logFile.Sync()
			}
			s.outputChan <- data
		}
		if err != nil {
			close(s.outputChan)
			return
		}
	}
}

// handleError continuously reads from stderr and logs it
func (s *DefaultChrootSession) handleError() {
	buf := make([]byte, 4096)
	for {
		n, err := s.stderrPipe.Read(buf)
		if n > 0 {
			data := string(buf[:n])
			if s.logFile != nil {
				s.logFile.WriteString(data)
				s.logFile.Sync()
			}
		}
		if err != nil {
			return
		}
	}
}

// Exec executes a command in the chroot session
func (s *DefaultChrootSession) Exec(command string) error {
	return s.ExecWithContext(s.ctx, command)
}

// ExecWithOutput executes a command and returns its output
func (s *DefaultChrootSession) ExecWithOutput(command string) (string, error) {
	return s.ExecWithOutputAndContext(s.ctx, command)
}

// ExecWithContext executes a command with context support
func (s *DefaultChrootSession) ExecWithContext(ctx context.Context, command string) error {
	_, err := s.ExecWithOutputAndContext(ctx, command)
	return err
}

// ExecWithOutputAndContext executes a command with context and returns output
func (s *DefaultChrootSession) ExecWithOutputAndContext(ctx context.Context, command string) (string, error) {
	if s.closed {
		return "", fmt.Errorf("session is closed")
	}

	// Clear the output channel
	for len(s.outputChan) > 0 {
		<-s.outputChan
	}

	// Write the command followed by a marker
	marker := fmt.Sprintf("ARCHUP_CMD_DONE_%d", os.Getpid())
	fullCommand := fmt.Sprintf("%s\necho \"%s_$?\"\n", command, marker)

	if _, err := s.stdin.WriteString(fullCommand); err != nil {
		return "", fmt.Errorf("failed to write command: %w", err)
	}

	// Collect output until we see the marker
	var output strings.Builder
	var exitCode string

	for {
		select {
		case <-ctx.Done():
			return output.String(), fmt.Errorf("command context cancelled: %w", ctx.Err())
		case data, ok := <-s.outputChan:
			if !ok {
				return output.String(), fmt.Errorf("session stdout closed unexpectedly")
			}

			// Check if output contains the marker
			if strings.Contains(data, marker) {
				// Extract exit code from marker line
				lines := strings.Split(data, "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, marker+"_") {
						exitCode = strings.TrimPrefix(line, marker+"_")
						exitCode = strings.TrimSpace(exitCode)

						// Add any output before the marker
						beforeMarker := strings.Split(data, marker)[0]
						output.WriteString(beforeMarker)

						// Check exit code
						if exitCode != "0" {
							return output.String(), fmt.Errorf("command exited with code %s", exitCode)
						}
						return output.String(), nil
					}
				}
			} else {
				output.WriteString(data)
			}
		}
	}
}

// Close terminates the session and cleans up resources
func (s *DefaultChrootSession) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true

	// Send exit command
	if s.stdin != nil {
		s.stdin.WriteString("exit\n")
		s.stdin.Close()
	}

	// Wait for the command to finish with timeout
	done := make(chan error, 1)
	go func() {
		done <- s.cmd.Wait()
	}()

	select {
	case <-time.After(5 * time.Second):
		// Force kill if it doesn't exit gracefully
		if s.cmd.Process != nil {
			s.cmd.Process.Kill()
		}
	case <-done:
		// Command exited normally
	}

	// Close pipes and log file
	if s.stdoutPipe != nil {
		s.stdoutPipe.Close()
	}
	if s.stderrPipe != nil {
		s.stderrPipe.Close()
	}
	if s.logFile != nil {
		s.logFile.Close()
	}

	// Cancel context
	s.cancel()

	return nil
}
