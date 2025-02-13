// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows

package exec_test

import (
	"internal/testenv"
	"io"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"testing"
)

func TestPipePassing(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	const marker = "arrakis, dune, desert planet"
	childProc := helperCommand(t, "pipehandle", strconv.FormatUint(uint64(w.Fd()), 16), marker)
	childProc.SysProcAttr = &syscall.SysProcAttr{AdditionalInheritedHandles: []syscall.Handle{syscall.Handle(w.Fd())}}
	err = childProc.Start()
	if err != nil {
		t.Error(err)
	}
	w.Close()
	response, err := io.ReadAll(r)
	if err != nil {
		t.Error(err)
	}
	r.Close()
	if string(response) != marker {
		t.Errorf("got %q; want %q", string(response), marker)
	}
	err = childProc.Wait()
	if err != nil {
		t.Error(err)
	}
}

func TestNoInheritHandles(t *testing.T) {
	cmd := exec.Command("cmd", "/c exit 88")
	cmd.SysProcAttr = &syscall.SysProcAttr{NoInheritHandles: true}
	err := cmd.Run()
	exitError, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("got error %v; want ExitError", err)
	}
	if exitError.ExitCode() != 88 {
		t.Fatalf("got exit code %d; want 88", exitError.ExitCode())
	}
}

func TestErrProcessDone(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	// On Windows, ProcAttr cannot be empty
	p, err := os.StartProcess(testenv.GoToolPath(t), []string{""},
		&os.ProcAttr{Dir: "", Env: nil, Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}, Sys: nil})
	if err != nil {
		t.Errorf("starting test process: %v", err)
	}
	_, err = p.Wait()
	if err != nil {
		t.Errorf("Wait: %v", err)
	}
	if got := p.Signal(os.Kill); got != os.ErrProcessDone {
		t.Fatalf("got %v want %v", got, os.ErrProcessDone)
	}
}
