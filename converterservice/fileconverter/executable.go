package fileconverter

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

type defaultExecutable struct {
	cmd *exec.Cmd
}

// Returns a new executable with the given cmd
func newDefaultExecutable(command string, args ...string) Executable {
	return &defaultExecutable{
		cmd: exec.Command(command, args...),
	}
}

func (e *defaultExecutable) Start() error {
	if err := e.cmd.Start(); err != nil {
		return err
	}
	return nil
}

func (e *defaultExecutable) Wait() error {
	if err := e.cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func (e *defaultExecutable) Stdout() io.Writer {
	return e.cmd.Stdout
}

func (e *defaultExecutable) SetStdout(stdout io.Writer) {
	e.cmd.Stdout = stdout
}

func (e *defaultExecutable) Stdin() io.Reader {
	return e.cmd.Stdin
}

func (e *defaultExecutable) SetStdin(stdin io.Reader) {
	e.cmd.Stdin = stdin
}

func (e *defaultExecutable) Stderr() io.Writer {
	return e.cmd.Stderr
}

func (e *defaultExecutable) SetStderr(stderr io.Writer) {
	e.cmd.Stderr = stderr
}