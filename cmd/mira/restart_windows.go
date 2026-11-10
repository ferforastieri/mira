package main

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// startAgain starts a new Mira, as this one was started, which reads the
// restored database. Windows cannot replace a process with another: this
// one waits for the new Mira and exits as it does, so that whoever
// started Mira, such as the desktop app, still watches its process. The
// new Mira shares this one's console: it stops on Ctrl-C or Ctrl-Break,
// which this one ignores meanwhile, and joins this one's job, which ends it
// if this one is killed (see process.KillChildrenOnExit). It only returns
// on failure.
func startAgain() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(executable, os.Args[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	signal.Ignore(os.Interrupt, syscall.SIGTERM)
	if err := cmd.Start(); err != nil {
		return errors.Join(errors.New("cannot start Mira again"), err)
	}
	_ = cmd.Wait()
	os.Exit(cmd.ProcessState.ExitCode())
	return nil
}
