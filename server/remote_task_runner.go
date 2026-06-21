package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var ErrUnsupportedRemoteTaskMethod = errors.New("unsupported remote task method")
var ErrUnsupportedRemoteTaskCredential = errors.New("unsupported credential type for remote task")

type remoteTaskCommand struct {
	Tool    string
	Args    []string
	Preview string
	Env     []string
	WorkDir string
	Cleanup func()
	Address string
}

func RunRemoteTask(teamName string, req RunRemoteTaskRequest) (*RemoteTaskRun, error) {
	if _, err := GetTeamByName(teamName); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Command) == "" {
		return nil, errors.New("command is required")
	}

	target, err := GetTargetByID(req.TargetID)
	if err != nil {
		return nil, err
	}
	if target.TeamName != teamName {
		return nil, sql.ErrNoRows
	}

	credential, err := GetCredentialByID(req.CredentialID)
	if err != nil {
		return nil, err
	}
	if credential.TeamName != teamName {
		return nil, sql.ErrNoRows
	}

	command, err := buildRemoteTaskCommand(*target, *credential, req)
	if err != nil {
		return nil, err
	}
	if command.Cleanup == nil {
		command.Cleanup = func() {}
	}
	defer command.Cleanup()

	timeout := time.Duration(req.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	if timeout > 300*time.Second {
		timeout = 300 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	started := time.Now().UTC()
	cmd := exec.CommandContext(ctx, command.Tool, command.Args...)
	cmd.Dir = command.WorkDir
	cmd.Env = append(os.Environ(), command.Env...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err = cmd.Run()
	finished := time.Now().UTC()

	status := "succeeded"
	errorText := ""
	if ctx.Err() == context.DeadlineExceeded {
		status = "failed"
		errorText = "remote task timed out"
	} else if err != nil {
		status = "failed"
		errorText = err.Error()
	}

	run := RemoteTaskRun{
		TeamName:       teamName,
		TargetID:       target.ID,
		TargetLabel:    targetLabel(*target),
		TargetAddress:  command.Address,
		CredentialID:   credential.ID,
		CredentialName: credentialLabel(*credential),
		Method:         req.Method,
		Command:        strings.TrimSpace(req.Command),
		CommandPreview: command.Preview,
		Status:         status,
		Output:         output.String(),
		Error:          errorText,
		FinishedAt:     finished.Format(time.RFC3339),
	}

	created, createErr := CreateRemoteTaskRun(run)
	if createErr != nil {
		return nil, createErr
	}
	created.StartedAt = started.Format(time.RFC3339)
	return created, err
}

func buildRemoteTaskCommand(target Target, credential Credential, req RunRemoteTaskRequest) (*remoteTaskCommand, error) {
	method := strings.ToLower(strings.TrimSpace(req.Method))
	if method != "wmiexec" && method != "psexec" {
		return nil, ErrUnsupportedRemoteTaskMethod
	}

	tool := strings.TrimSpace(req.ToolCommand)
	if tool == "" {
		tool = "impacket-" + method
	}
	args := []string{}
	previewArgs := []string{}
	targetSpec, previewSpec, targetAddress, err := remoteTaskTargetSpec(credential, target)
	if err != nil {
		return nil, err
	}
	env := []string{}
	workDir := serverWorkingDir()
	cleanup := func() {}

	isKerberos := credential.SecretType == "kerberos-aes256" || credential.SecretType == "kerberos-aes"
	if isKerberos {
		if strings.TrimSpace(req.KDCHost) == "" {
			return nil, errors.New("KDC host is required for Kerberos remote tasks")
		}
		args = append(args, "-k", "-dc-ip", strings.TrimSpace(req.KDCHost))
		previewArgs = append(previewArgs, "-k", "-dc-ip", strings.TrimSpace(req.KDCHost))

		if shouldUseNSSWrapperForKerberos() {
			spec, err := buildNSSWrapperSpec(target.TeamName, env)
			if err != nil {
				return nil, err
			}
			env = spec.Env
			workDir = spec.WorkDir
			cleanup = spec.Cleanup
			if err := validateKerberosTargetLookup(targetAddress, append(os.Environ(), env...)); err != nil {
				cleanup()
				return nil, err
			}
		}
	}

	switch credential.SecretType {
	case "password":
		args = append(args, targetSpec)
		previewArgs = append(previewArgs, previewSpec)
	case "ntlm", "kerberos-ntlm":
		lmHash, ntHash := splitHashSecret(credential.Secret)
		args = append(args, "-hashes", lmHash+":"+ntHash, targetSpec)
		previewArgs = append(previewArgs, "-hashes", lmHash+":<redacted>", previewSpec)
	case "kerberos-aes256", "kerberos-aes":
		args = append(args, "-aesKey", credential.Secret, targetSpec)
		previewArgs = append(previewArgs, "-aesKey", "<redacted>", previewSpec)
	default:
		return nil, ErrUnsupportedRemoteTaskCredential
	}

	commandText := strings.TrimSpace(req.Command)
	args = append(args, commandText)
	previewArgs = append(previewArgs, commandText)

	return &remoteTaskCommand{
		Tool:    tool,
		Args:    args,
		Preview: tool + " " + strings.Join(previewArgs, " "),
		Env:     env,
		WorkDir: workDir,
		Cleanup: cleanup,
		Address: targetAddress,
	}, nil
}

func remoteTaskTargetSpec(credential Credential, target Target) (string, string, string, error) {
	user := strings.TrimSpace(credential.Username)
	if user == "" {
		return "", "", "", errors.New("credential username is required")
	}

	identity := user
	if domain := strings.TrimSpace(credential.Domain); domain != "" {
		identity = domain + "/" + user
	}

	address := targetAddressForCredential(target, credential)
	if address == "" {
		return "", "", "", errors.New("target address is required")
	}

	if credential.SecretType == "password" {
		return fmt.Sprintf("%s:%s@%s", identity, credential.Secret, address),
			fmt.Sprintf("%s:<redacted>@%s", identity, address),
			address,
			nil
	}

	return fmt.Sprintf("%s@%s", identity, address), fmt.Sprintf("%s@%s", identity, address), address, nil
}

func splitHashSecret(secret string) (string, string) {
	parts := strings.Split(secret, ":")
	if len(parts) >= 2 {
		return parts[0], parts[len(parts)-1]
	}
	return "", secret
}

func targetAddress(target Target) string {
	if strings.TrimSpace(target.IP) != "" {
		return strings.TrimSpace(target.IP)
	}
	return strings.TrimSpace(target.Hostname)
}

func targetAddressForCredential(target Target, credential Credential) string {
	if credential.SecretType == "kerberos-aes256" || credential.SecretType == "kerberos-aes" {
		return targetFQDN(target)
	}
	return targetAddress(target)
}

func targetFQDN(target Target) string {
	hostname := strings.TrimSpace(target.Hostname)
	domainName := strings.TrimSpace(target.DomainName)
	if hostname == "" {
		return strings.TrimSpace(target.IP)
	}
	if domainName != "" && !strings.Contains(hostname, ".") {
		return hostname + "." + domainName
	}
	return hostname
}

func targetLabel(target Target) string {
	if strings.TrimSpace(target.Hostname) != "" {
		return strings.TrimSpace(target.Hostname)
	}
	return targetAddress(target)
}

func credentialLabel(credential Credential) string {
	if strings.TrimSpace(credential.Domain) != "" {
		return credential.Domain + "\\" + credential.Username
	}
	return credential.Username
}
