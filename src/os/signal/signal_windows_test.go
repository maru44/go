// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package signal

import (
	"bytes"
	"internal/testenv"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestCtrlBreak(t *testing.T) {
	// create source file
	const source = `
package main

import (
	"log"
	"os"
	"os/signal"
	"time"
)


func main() {
	c := make(chan os.Signal, 10)
	signal.Notify(c)
	select {
	case s := <-c:
		if s != os.Interrupt {
			log.Fatalf("Wrong signal received: got %q, want %q\n", s, os.Interrupt)
		}
	case <-time.After(3 * time.Second):
		log.Fatalf("Timeout waiting for Ctrl+Break\n")
	}
}
`
	tmp := t.TempDir()

	// write ctrlbreak.go
	name := filepath.Join(tmp, "ctlbreak")
	src := name + ".go"
	f, err := os.Create(src)
	if err != nil {
		t.Fatalf("Failed to create %v: %v", src, err)
	}
	defer f.Close()
	f.Write([]byte(source))

	// compile it
	exe := name + ".exe"
	defer os.Remove(exe)
	o, err := exec.Command(testenv.GoToolPath(t), "build", "-o", exe, src).CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to compile: %v\n%v", err, string(o))
	}

	// run it
	cmd := exec.Command(exe)
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
	err = cmd.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	go func() {
		time.Sleep(1 * time.Second)
		cmd.Process.Signal(os.Interrupt)
	}()
	err = cmd.Wait()
	if err != nil {
		t.Fatalf("Program exited with error: %v\n%v", err, string(b.Bytes()))
	}
}
