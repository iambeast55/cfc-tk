package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

var defaultNSSWrapperLibraryPaths = []string{
	"/usr/lib/x86_64-linux-gnu/libnss_wrapper.so",
	"/usr/lib64/libnss_wrapper.so",
	"/usr/lib/libnss_wrapper.so",
}

type hostAlias struct {
	Name string
	IP   string
}

type nssWrapperSpec struct {
	Env         []string
	WorkDir     string
	Cleanup     func()
	HostsPath   string
	LibraryPath string
}

func shouldUseNSSWrapperForKerberos() bool {
	return runtime.GOOS == "linux"
}

func buildTeamHostAliases(teamName string) ([]hostAlias, error) {
	targets, err := GetTargetsByTeamName(teamName)
	if err != nil {
		return nil, err
	}
	return hostAliasesForTargets(targets)
}

func hostAliasesForTargets(targets []Target) ([]hostAlias, error) {
	seen := map[string]string{}
	var aliases []hostAlias

	for _, target := range targets {
		for _, name := range targetAliasNames(target) {
			key := strings.ToLower(name)
			if existingIP, ok := seen[key]; ok {
				if existingIP != target.IP {
					return nil, fmt.Errorf("target alias %q resolves to both %s and %s", name, existingIP, target.IP)
				}
				continue
			}
			seen[key] = target.IP
			aliases = append(aliases, hostAlias{Name: key, IP: target.IP})
		}
	}

	slices.SortFunc(aliases, func(a hostAlias, b hostAlias) int {
		return strings.Compare(a.Name, b.Name)
	})

	return aliases, nil
}

func targetAliasNames(target Target) []string {
	hostname := strings.ToLower(strings.TrimSpace(target.Hostname))
	domain := strings.ToLower(strings.Trim(strings.TrimSpace(target.DomainName), "."))
	if hostname == "" || strings.TrimSpace(target.IP) == "" {
		return nil
	}

	names := []string{hostname}
	if strings.Contains(hostname, ".") {
		short, _, _ := strings.Cut(hostname, ".")
		if short != "" {
			names = append(names, short)
		}
	} else if domain != "" {
		names = append(names, hostname+"."+domain)
	}

	seen := map[string]bool{}
	var deduped []string
	for _, name := range names {
		name = strings.Trim(name, ".")
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		deduped = append(deduped, name)
	}

	return deduped
}

func buildNSSWrapperSpec(teamName string, baseEnv []string) (*nssWrapperSpec, error) {
	if !shouldUseNSSWrapperForKerberos() {
		return &nssWrapperSpec{Env: baseEnv, WorkDir: serverWorkingDir(), Cleanup: func() {}}, nil
	}

	libPath, err := findNSSWrapperLibrary()
	if err != nil {
		return nil, err
	}

	aliases, err := buildTeamHostAliases(teamName)
	if err != nil {
		return nil, err
	}

	hostsPath, err := writeNSSWrapperHostsFile(teamName, aliases)
	if err != nil {
		return nil, err
	}

	env := append([]string{}, baseEnv...)
	env = append(env,
		"LD_PRELOAD="+libPath,
		"NSS_WRAPPER_HOSTS="+hostsPath,
		"NSS_WRAPPER_PASSWD=/etc/passwd",
		"NSS_WRAPPER_GROUP=/etc/group",
	)

	return &nssWrapperSpec{
		Env:         env,
		WorkDir:     serverWorkingDir(),
		LibraryPath: libPath,
		HostsPath:   hostsPath,
		Cleanup: func() {
			_ = os.Remove(hostsPath)
		},
	}, nil
}

func findNSSWrapperLibrary() (string, error) {
	if value := strings.TrimSpace(os.Getenv("CFC_TK_NSS_WRAPPER_LIB")); value != "" {
		if _, err := os.Stat(value); err != nil {
			return "", fmt.Errorf("CFC_TK_NSS_WRAPPER_LIB points to an unreadable file: %w", err)
		}
		return value, nil
	}

	for _, candidate := range defaultNSSWrapperLibraryPaths {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	if path, err := exec.LookPath("libnss_wrapper.so"); err == nil {
		return path, nil
	}

	return "", errors.New("libnss_wrapper.so not found; install libnss-wrapper or set CFC_TK_NSS_WRAPPER_LIB")
}

func writeNSSWrapperHostsFile(teamName string, aliases []hostAlias) (string, error) {
	file, err := os.CreateTemp("", safeArtifactPart(teamName)+"-hosts-*.txt")
	if err != nil {
		return "", err
	}
	defer file.Close()

	lines := []string{
		"127.0.0.1 localhost",
		"::1 localhost ip6-localhost ip6-loopback",
	}
	for _, alias := range aliases {
		lines = append(lines, alias.IP+" "+alias.Name)
	}

	if _, err := file.WriteString(strings.Join(lines, "\n") + "\n"); err != nil {
		_ = os.Remove(file.Name())
		return "", err
	}

	return file.Name(), nil
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

func validateKerberosTargetLookup(target string, env []string) error {
	if strings.TrimSpace(target) == "" {
		return errors.New("target is required")
	}

	getentPath, err := exec.LookPath("getent")
	if err != nil {
		return nil
	}

	cmd := exec.Command(getentPath, "ahosts", target)
	cmd.Env = env
	outputBytes, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(outputBytes))
	if err != nil {
		if output == "" {
			output = err.Error()
		}
		return fmt.Errorf("Kerberos target lookup failed for %s under nss_wrapper: %s", target, output)
	}
	if output == "" {
		return fmt.Errorf("Kerberos target lookup returned no addresses for %s under nss_wrapper", target)
	}
	return nil
}
