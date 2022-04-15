package gbs

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// BuilderFunc
type BuilderFunc func() error

// BuildFuncOpt
type BuildFuncOpt struct {
	FuncName string
	Func     BuilderFunc
}

// Build accept BuildFuncOpt and retuns an error if any of the BuilderFunc has
// failed otherwise nil is returned.
func Build(opts ...BuildFuncOpt) error {
	for _, opt := range opts {
		log.Printf("Executing %s", opt.FuncName)
		if err := opt.Func(); err != nil {
			return fmt.Errorf("Failed to execute %s [%w]", opt.FuncName, err)
		}
	}
	return nil
}

// Sh
type Sh struct {
	cmd *exec.Cmd

	err error
}

// Init
func (s *Sh) Init(command string) *Sh {
	sp := strings.Split(command, " ")
	var (
		root string
		args []string
	)
	for i, n := range sp {
		switch i {
		case 0:
			root = n
		case 1:
			args = sp[i:]
			break
		}
	}
	cmd := exec.Command(root, args...)
	s.cmd = cmd
	return s
}

// Run
func (s *Sh) Run() *Sh {
	if s.cmd == nil {
		s.err = fmt.Errorf("sh is not initialized please use Init method.")
		return s
	}

	s.cmd.Stdout = os.Stdout
	s.cmd.Stdin = os.Stdin
	s.cmd.Stderr = os.Stderr

	if err := s.cmd.Start(); err != nil {
		s.err = err
		return s
	}

	if err := s.cmd.Wait(); err != nil {
		s.err = err
	}

	return s
}

// In
func (s *Sh) In(inCommand string) *Sh {
	writer, err := s.cmd.StdinPipe()
	if err != nil {
		s.err = err
		return s
	}

	isExec := make(chan struct{})
	done := make(chan bool)

	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr

	defer close(isExec)
	defer close(done)

	go func() {
		<-isExec
		if err := s.cmd.Wait(); err != nil {
			s.err = err
		}
	}()

	go func() {
		if err := s.cmd.Start(); err != nil {
			s.err = err
			done <- true
			return
		}
		isExec <- struct{}{}
		if _, err := writer.Write([]byte(inCommand)); err != nil {
			s.err = err
		}
		done <- true
	}()

	<-done
	return s
}

// Error
func (s *Sh) Error() error {
	return s.err
}
