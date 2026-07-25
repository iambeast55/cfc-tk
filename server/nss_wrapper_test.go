package main

import (
	"strings"
	"testing"
)

func TestHostAliasesForTargetsAddsShortAndFQDNNames(t *testing.T) {
	aliases, err := hostAliasesForTargets([]Target{
		{Hostname: "dc01", DomainName: "corp.local", IP: "10.10.10.5"},
		{Hostname: "fileserver.corp.local", DomainName: "corp.local", IP: "10.10.10.20"},
	})
	if err != nil {
		t.Fatalf("hostAliasesForTargets: %v", err)
	}

	got := map[string]string{}
	for _, alias := range aliases {
		got[alias.Name] = alias.IP
	}

	want := map[string]string{
		"dc01":                  "10.10.10.5",
		"dc01.corp.local":       "10.10.10.5",
		"fileserver":            "10.10.10.20",
		"fileserver.corp.local": "10.10.10.20",
	}

	if len(got) != len(want) {
		t.Fatalf("unexpected alias count: got %d want %d (%v)", len(got), len(want), got)
	}
	for name, ip := range want {
		if got[name] != ip {
			t.Fatalf("alias %s => %q, want %q", name, got[name], ip)
		}
	}
}

func TestHostAliasesForTargetsRejectsConflictingAliases(t *testing.T) {
	_, err := hostAliasesForTargets([]Target{
		{Hostname: "dc01", DomainName: "corp.local", IP: "10.10.10.5"},
		{Hostname: "dc01", DomainName: "corp.local", IP: "10.10.10.6"},
	})
	if err == nil {
		t.Fatal("expected duplicate alias conflict")
	}
}

func TestTerminalShellCommandIncludesCleanupAndWorkDir(t *testing.T) {
	command := terminalShellCommand(
		"red:dc01",
		[]string{"NSS_WRAPPER_HOSTS=/tmp/cfc.hosts", "KRB5CCNAME=/tmp/admin.ccache"},
		[]string{"impacket-wmiexec", "-k", "corp.local/admin@dc01.corp.local"},
		"/opt/cfc-impui/server",
		"/tmp/cfc-impui-launcher.sh",
	)

	for _, snippet := range []string{
		"cd '/opt/cfc-impui/server'",
		"trap ",
		"/tmp/cfc-impui-launcher.sh",
		"/tmp/cfc.hosts",
		"'env' 'NSS_WRAPPER_HOSTS=/tmp/cfc.hosts' 'KRB5CCNAME=/tmp/admin.ccache'",
		"'impacket-wmiexec' '-k' 'corp.local/admin@dc01.corp.local'",
	} {
		if !strings.Contains(command, snippet) {
			t.Fatalf("terminal command missing %q: %s", snippet, command)
		}
	}
}
