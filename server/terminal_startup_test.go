package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTerminalStartupHelper(t *testing.T) {
	if os.Getenv("IMPUI_TEST_TERMINAL") != "1" {
		return
	}
	mode, ready := os.Args[len(os.Args)-2], os.Args[len(os.Args)-1]
	switch mode {
	case "ready":
		if err := os.WriteFile(ready, []byte("ready"), 0o600); err != nil {
			os.Exit(2)
		}
	case "failure":
		fmt.Fprintln(os.Stderr, "unsupported terminal argument")
		os.Exit(1)
	}
	os.Exit(0)
}

func TestTerminalRequiresAcknowledgment(t *testing.T) {
	for _, mode := range []string{"ready", "failure", "handoff-without-ready"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			ready := filepath.Join(dir, "ready")
			cmd := exec.Command(os.Args[0], "-test.run=^TestTerminalStartupHelper$", "--", mode, ready)
			cmd.Env = append(os.Environ(), "IMPUI_TEST_TERMINAL=1")
			err := startTerminalAndWait(cmd, ready, filepath.Join(dir, "startup.log"), 2*time.Second)
			if mode == "ready" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil {
				t.Fatal("reported success without a ready acknowledgment")
			} else if mode == "failure" && !strings.Contains(err.Error(), "unsupported terminal argument") {
				t.Fatalf("lost startup diagnostic: %v", err)
			}
		})
	}
}

func TestTerminalLauncherChecksTTYBeforeRunningCommand(t *testing.T) {
	path, err := writeTerminalLauncherScript("test", nil, []string{"echo", "fixture"}, "", "/tmp/ready marker")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	script := string(data)
	if !strings.Contains(script, "[ -t 0 ]") || !strings.Contains(script, "[ -t 1 ]") {
		t.Fatal("missing TTY checks")
	}
	if strings.Index(script, "printf ready") > strings.Index(script, "'fixture'") {
		t.Fatal("ready marker must precede tool invocation")
	}
}
