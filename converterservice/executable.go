package converterservice

import (
	"io"
	"os/exec"
)

// A wrapper for exec.Cmd
type Executable interface {
	// Starts the command
	// Analogous to cmd.Start()
	Start()  error
	// Waits for the command to finish executing.
	// Analogous to cmd.Wait()
	Wait()   error
	// A getter for stdout
	Stdout() io.Writer
	// Sets the stdout stream
	SetStdout(io.Writer)
	// A getter for stdin
	Stdin()  io.Reader
	// Sets the stdin stream
	SetStdin(io.Reader)
	// A getter for stderr
	Stderr() io.Writer
	// Sets the stderr stream
	SetStderr(io.Writer)
}

type convertExecutable struct {
	cmd *exec.Cmd
}

// A command factory
type ExecutableFactory interface {
	// Creates the appropriate file conversion command
	// using the conversion attributes
	SelectCommand(job *ConversionAttributes) Executable
}

// Returns a new executable with the given cmd
func NewExecutable(command string, args ...string) Executable {
	return &convertExecutable{
		cmd: exec.Command(command, args...),
	}
}

func (e *convertExecutable) Start() error {
	if err := e.cmd.Start(); err != nil {
		return err
	}
	return nil
}

func (e *convertExecutable) Wait() error {
	if err := e.cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func (e *convertExecutable) Stdout() io.Writer {
	return e.cmd.Stdout
}

func (e *convertExecutable) SetStdout(stdout io.Writer) {
	e.cmd.Stdout = stdout
}

func (e *convertExecutable) Stdin() io.Reader {
	return e.cmd.Stdin
}

func (e *convertExecutable) SetStdin(stdin io.Reader) {
	e.cmd.Stdin = stdin
}

func (e *convertExecutable) Stderr() io.Writer {
	return e.cmd.Stderr
}

func (e *convertExecutable) SetStderr(stderr io.Writer) {
	e.cmd.Stderr = stderr
}