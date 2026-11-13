package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestLinuxSSHPasswordNotInArguments(t *testing.T) {
	command, err := linuxToolCommand(LinuxConnectionRequest{Username: "operator", Action: "ssh", AuthMode: "password", Password: "fixture-secret", Port: 2222}, Target{OS: "linux", IP: "192.0.2.10"})
	if err != nil {
		t.Fatal(err)
	}
	if command[0] != "ssh" || command[len(command)-1] != "operator@192.0.2.10" || !slices.Contains(command, "2222") {
		t.Fatalf("unexpected command: %v", command)
	}
	if strings.Contains(strings.Join(command, " "), "fixture-secret") {
		t.Fatal("password exposed in arguments")
	}
}

func TestLinuxSCPKeyPathsAndIPv6(t *testing.T) {
	dir := t.TempDir()
	key, source := filepath.Join(dir, "private key"), filepath.Join(dir, "source file")
	for _, path := range []string{key, source} {
		if err := os.WriteFile(path, []byte("fixture"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, action := range []string{"upload", "download"} {
		command, err := linuxToolCommand(LinuxConnectionRequest{Username: "operator", Action: action, AuthMode: "key", KeyPath: key, LocalPath: source, RemotePath: "/tmp/file with spaces"}, Target{OS: "linux", IP: "2001:db8::10"})
		if err != nil {
			t.Fatal(err)
		}
		if command[0] != "scp" || !slices.Contains(command, "-P") || !slices.Contains(command, key) {
			t.Fatalf("bad SCP key arguments: %v", command)
		}
		operands := command[len(command)-2:]
		remote := "operator@[2001:db8::10]:/tmp/file with spaces"
		if action == "upload" && (operands[0] != source || operands[1] != remote) {
			t.Fatalf("bad upload: %v", operands)
		}
		if action == "download" && (operands[0] != remote || operands[1] != source) {
			t.Fatalf("bad download: %v", operands)
		}
	}
}

func TestLinuxConnectionsRejectInvalidInput(t *testing.T) {
	base := LinuxConnectionRequest{Username: "operator", Action: "ssh", AuthMode: "password", Port: 22}
	for _, change := range []func(*LinuxConnectionRequest){
		func(r *LinuxConnectionRequest) { r.Username = "-oProxyCommand=bad" },
		func(r *LinuxConnectionRequest) { r.Port = 65536 },
		func(r *LinuxConnectionRequest) { r.Action = "other" },
		func(r *LinuxConnectionRequest) { r.AuthMode = "key"; r.KeyPath = "relative-key" },
		func(r *LinuxConnectionRequest) {
			r.Action = "upload"
			r.LocalPath = "relative-file"
			r.RemotePath = "/tmp/file"
		},
	} {
		req := base
		change(&req)
		if _, err := linuxToolCommand(req, Target{OS: "linux", IP: "192.0.2.1"}); err == nil {
			t.Fatalf("accepted invalid input: %+v", req)
		}
	}
	if _, err := linuxToolCommand(base, Target{OS: "windows", IP: "192.0.2.1"}); err == nil {
		t.Fatal("accepted Windows target")
	}
}
