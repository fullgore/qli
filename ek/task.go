package main

import "os/exec"

type Task struct {
	ID      string
	Command string
	Cmd     *exec.Cmd
	State   string
}

func (t Task) stop() error {
	if t.Cmd != nil && t.Cmd.Process != nil {
		err := t.Cmd.Process.Kill()
		if err != nil {
			return err
		}
	}
	return nil
}
