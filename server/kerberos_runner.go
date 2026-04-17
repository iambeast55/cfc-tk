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

var safeArtifactPartPattern = regexp.MustCompile(`[^A-Za-z0-9_.-]+`)

func RunKerberosTicket(teamName string, req RunKerberosTicketRequest) (*RunKerberosTicketResponse, error) {
	artifactDir, displayDir, err := kerberosArtifactDir(teamName, req.Domain)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(artifactDir, 0o700); err != nil {
		return nil, err
	}

	commandParts := splitToolCommand(req.ToolCommand)
	if len(commandParts) == 0 {
		return nil, errors.New("tool command is required")
	}

	args, err := kerberosCommandArgs(req)
	if err != nil {
		return nil, err
	}

	startedAt := time.Now().Add(-1 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, commandParts[0], append(commandParts[1:], args...)...)
	cmd.Dir = artifactDir
	outputBytes, runErr := cmd.CombinedOutput()
	output := string(outputBytes)
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("ticket command timed out: %s", output)
	}
	if runErr != nil {
		return nil, fmt.Errorf("ticket command failed: %w\n%s", runErr, output)
	}

	cacheFile, err := findKerberosCacheFile(artifactDir, req.Username, startedAt)
	if err != nil {
		return nil, err
	}

	cachePath := filepath.ToSlash(filepath.Join(displayDir, filepath.Base(cacheFile)))
	cacheReq := CreateKerberosCacheRequest{
		Domain:    req.Domain,
		Username:  req.Username,
		Method:    req.Method,
		CachePath: cachePath,
		KDCHost:   req.KDCHost,
		DomainSID: req.DomainSID,
		UserID:    req.UserID,
		Groups:    req.Groups,
		ExtraSID:  req.ExtraSID,
		Duration:  req.Duration,
		ExpiresAt: req.ExpiresAt,
		Notes:     req.Notes,
	}

	cache, err := CreateKerberosCache(teamName, cacheReq)
	if errors.Is(err, ErrDuplicateKerberosCache) {
		cache, err = GetKerberosCacheByKey(teamName, req.Domain, req.Username, req.Method, cachePath)
	}
	if err != nil {
		return nil, err
	}

	return &RunKerberosTicketResponse{
		Cache:   cache,
		Command: append(commandParts, args...),
		Output:  output,
	}, nil
}

func kerberosCommandArgs(req RunKerberosTicketRequest) ([]string, error) {
	userPrefix := req.Username
	if req.Domain != "" {
		userPrefix = req.Domain + "/" + req.Username
	}

	switch req.Method {
	case "getTGT":
		args := []string{}
		if req.KDCHost != "" {
			args = append(args, "-dc-ip", req.KDCHost)
		}
		switch req.TicketAuthMode {
		case "password":
			if req.Password == "" {
				return nil, errors.New("password is required")
			}
			args = append(args, userPrefix+":"+req.Password)
		case "hash":
			if req.NTHash == "" {
				return nil, errors.New("NT hash is required")
			}
			args = append(args, "-hashes", req.LMHash+":"+req.NTHash, userPrefix)
		case "aes":
			if req.AESKey == "" {
				return nil, errors.New("user AES key is required")
			}
			args = append(args, "-aesKey", req.AESKey, userPrefix)
		default:
			return nil, errors.New("ticket auth mode must be password, hash, or aes")
		}
		return args, nil
	case "ticketer":
		if req.KrbTGTAESKey == "" {
			return nil, errors.New("krbtgt AES key is required")
		}
		if req.DomainSID == "" {
			return nil, errors.New("domain SID is required")
		}
		args := []string{
			"-aesKey", req.KrbTGTAESKey,
			"-domain-sid", req.DomainSID,
			"-domain", req.Domain,
		}
		if req.UserID != "" {
			args = append(args, "-user-id", req.UserID)
		}
		if req.Groups != "" {
			args = append(args, "-groups", req.Groups)
		}
		if req.ExtraSID != "" {
			args = append(args, "-extra-sid", req.ExtraSID)
		}
		if req.Duration != "" {
			args = append(args, "-duration", req.Duration)
		}
		args = append(args, req.Username)
		return args, nil
	default:
		return nil, errors.New("method must be getTGT or ticketer")
	}
}

func kerberosArtifactDir(teamName string, domain string) (string, string, error) {
	teamPart := safeArtifactPart(teamName)
	domainPart := safeArtifactPart(domain)
	if teamPart == "" || domainPart == "" {
		return "", "", errors.New("team and domain are required")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	if filepath.Base(cwd) == "server" {
		return filepath.Join(cwd, teamPart, domainPart), filepath.ToSlash(filepath.Join("server", teamPart, domainPart)), nil
	}

	return filepath.Join(cwd, "server", teamPart, domainPart), filepath.ToSlash(filepath.Join("server", teamPart, domainPart)), nil
}

func findKerberosCacheFile(dir string, username string, startedAt time.Time) (string, error) {
	preferred := filepath.Join(dir, safeArtifactPart(username)+".ccache")
	if info, err := os.Stat(preferred); err == nil && !info.IsDir() {
		return preferred, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var newestPath string
	var newestTime time.Time
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".ccache") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(startedAt) {
			continue
		}
		if newestPath == "" || info.ModTime().After(newestTime) {
			newestPath = filepath.Join(dir, entry.Name())
			newestTime = info.ModTime()
		}
	}

	if newestPath == "" {
		return "", errors.New("command completed but no new .ccache file was found")
	}
	return newestPath, nil
}

func safeArtifactPart(value string) string {
	return strings.Trim(safeArtifactPartPattern.ReplaceAllString(value, "_"), "._-")
}

func splitToolCommand(command string) []string {
	return strings.Fields(command)
}
