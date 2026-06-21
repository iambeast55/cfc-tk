package main

import (
	"strings"
	"testing"
)

func TestBuildRemoteTaskCommandRedactsPassword(t *testing.T) {
	target := Target{
		ID:       1,
		TeamName: "blue",
		Hostname: "host01",
		IP:       "10.0.0.5",
		OS:       "windows",
	}
	credential := Credential{
		ID:         2,
		TeamName:   "blue",
		Username:   "administrator",
		Domain:     "CORP",
		SecretType: "password",
		Secret:     "SuperSecret!",
	}

	command, err := buildRemoteTaskCommand(target, credential, RunRemoteTaskRequest{
		Method:      "wmiexec",
		ToolCommand: "wmiexec.py",
		Command:     "whoami",
	})
	if err != nil {
		t.Fatalf("build command: %v", err)
	}

	if command.Tool != "wmiexec.py" {
		t.Fatalf("unexpected tool: %s", command.Tool)
	}
	if !strings.Contains(strings.Join(command.Args, " "), "SuperSecret!") {
		t.Fatalf("expected raw args to include password for execution: %+v", command.Args)
	}
	if strings.Contains(command.Preview, "SuperSecret!") {
		t.Fatalf("preview leaked password: %s", command.Preview)
	}
	if !strings.Contains(command.Preview, "wmiexec.py CORP/administrator:<redacted>@10.0.0.5") {
		t.Fatalf("unexpected preview: %s", command.Preview)
	}
}

func TestBuildRemoteTaskCommandUsesHashOption(t *testing.T) {
	target := Target{ID: 1, TeamName: "blue", IP: "10.0.0.5", OS: "windows"}
	credential := Credential{
		ID:         2,
		TeamName:   "blue",
		Username:   "administrator",
		SecretType: "ntlm",
		Secret:     "aad3b435b51404eeaad3b435b51404ee:8846f7eaee8fb117ad06bdd830b7586c",
	}

	command, err := buildRemoteTaskCommand(target, credential, RunRemoteTaskRequest{
		Method:  "psexec",
		Command: "hostname",
	})
	if err != nil {
		t.Fatalf("build command: %v", err)
	}
	if command.Tool != "impacket-psexec" {
		t.Fatalf("unexpected tool: %s", command.Tool)
	}
	if len(command.Args) < 4 || command.Args[0] != "-hashes" {
		t.Fatalf("expected -hashes args, got %+v", command.Args)
	}
	if strings.Contains(command.Preview, "8846f7eaee8fb117ad06bdd830b7586c") {
		t.Fatalf("preview leaked NT hash: %s", command.Preview)
	}
}

func TestBuildRemoteTaskCommandUsesKerberosRules(t *testing.T) {
	target := Target{
		ID:         1,
		TeamName:   "blue",
		Hostname:   "host01",
		DomainName: "corp.local",
		IP:         "10.0.0.5",
		OS:         "windows",
	}
	credential := Credential{
		ID:         2,
		TeamName:   "blue",
		Username:   "administrator",
		Domain:     "CORP.LOCAL",
		SecretType: "kerberos-aes256",
		Secret:     "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
	}

	command, err := buildRemoteTaskCommand(target, credential, RunRemoteTaskRequest{
		Method:  "wmiexec",
		Command: "whoami",
		KDCHost: "10.0.0.10",
	})
	if err != nil {
		t.Fatalf("build command: %v", err)
	}

	joined := strings.Join(command.Args, " ")
	for _, snippet := range []string{
		"-k",
		"-dc-ip 10.0.0.10",
		"-aesKey " + credential.Secret,
		"CORP.LOCAL/administrator@host01.corp.local",
	} {
		if !strings.Contains(joined, snippet) {
			t.Fatalf("missing %q in args: %s", snippet, joined)
		}
	}
	if !strings.Contains(command.Preview, "-k -dc-ip 10.0.0.10 -aesKey <redacted> CORP.LOCAL/administrator@host01.corp.local whoami") {
		t.Fatalf("unexpected preview: %s", command.Preview)
	}
	if command.Address != "host01.corp.local" {
		t.Fatalf("unexpected address: %s", command.Address)
	}
}

func TestBuildRemoteTaskCommandRequiresKDCForKerberos(t *testing.T) {
	target := Target{ID: 1, TeamName: "blue", Hostname: "host01", DomainName: "corp.local", IP: "10.0.0.5", OS: "windows"}
	credential := Credential{ID: 2, TeamName: "blue", Username: "administrator", SecretType: "kerberos-aes256", Secret: "abc"}

	_, err := buildRemoteTaskCommand(target, credential, RunRemoteTaskRequest{
		Method:  "wmiexec",
		Command: "whoami",
	})
	if err == nil {
		t.Fatal("expected missing KDC host error")
	}
}

func TestRemoteTaskRunHistoryPersistence(t *testing.T) {
	setupTestDB(t)

	if _, err := CreateNewTeam(CreateTeamRequest{Name: "blue", SubnetId: 1}); err != nil {
		t.Fatalf("create team: %v", err)
	}
	target, err := CreateTarget("blue", CreateTargetRequest{
		Hostname: "host01",
		IP:       "10.0.0.5",
		OS:       "windows",
	})
	if err != nil {
		t.Fatalf("create target: %v", err)
	}
	credential, err := CreateCredential("blue", CreateCredentialRequest{
		OS:         "windows",
		Username:   "administrator",
		SecretType: "password",
		Secret:     "SuperSecret!",
	})
	if err != nil {
		t.Fatalf("create credential: %v", err)
	}

	created, err := CreateRemoteTaskRun(RemoteTaskRun{
		TeamName:       "blue",
		TargetID:       target.ID,
		TargetLabel:    "host01",
		TargetAddress:  "10.0.0.5",
		CredentialID:   credential.ID,
		CredentialName: "administrator",
		Method:         "wmiexec",
		Command:        "whoami",
		CommandPreview: "impacket-wmiexec administrator:<redacted>@10.0.0.5 whoami",
		Status:         "succeeded",
		Output:         "corp\\administrator",
		FinishedAt:     "2026-06-15T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("create remote task run: %v", err)
	}

	runs, err := GetRemoteTaskRunsByTeamName("blue")
	if err != nil {
		t.Fatalf("get runs: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected one run, got %+v", runs)
	}
	if runs[0].ID != created.ID || runs[0].Output != "corp\\administrator" {
		t.Fatalf("unexpected run: %+v", runs[0])
	}
}
