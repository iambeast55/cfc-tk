package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type terminalSpec struct {
	name string
	args func(title string, shellCommand string) []string
}

func LaunchInteractiveCommand(teamName string, req LaunchInteractiveCommandRequest) (*LaunchInteractiveCommandResponse, error) {
	if _, err := GetTeamByName(teamName); err != nil {
		return nil, err
	}
	if runtime.GOOS != "linux" {
		return nil, errors.New("interactive terminal launch is supported on Linux only")
	}
	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		return nil, errors.New("no graphical desktop session detected; start CFCBench from a terminal in your Kali desktop as the logged-in user")
	}
	if !isInteractiveCommandKind(req.CommandKind) {
		return nil, errors.New("interactive command must be wmiexec, smbexec, or dcomexec")
	}

	commandParts := splitToolCommand(req.ToolCommand)
	if len(commandParts) == 0 {
		return nil, errors.New("tool command is required")
	}

	args, env, err := interactiveCommandArgs(req)
	if err != nil {
		return nil, err
	}

	title := terminalTitle(teamName, req.TargetLabel, req.Target)
	fullCommand := append(commandParts, args...)
	execSpec, err := prepareImpacketExecution(teamName, env, serverWorkingDir(), req.Target)
	if err != nil {
		return nil, err
	}
	return launchDesktopTerminal(title, fullCommand, execSpec.Env, execSpec.WorkDir, execSpec.Cleanup)
}

func launchDesktopTerminal(title string, fullCommand, env []string, workDir string, cleanup func()) (*LaunchInteractiveCommandResponse, error) {
	if runtime.GOOS != "linux" || (os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "") {
		cleanup()
		return nil, errors.New("start CFCBench from your Linux graphical desktop terminal to open a terminal window")
	}
	startupDir, err := os.MkdirTemp("", "cfc-impui-terminal-*")
	if err != nil {
		cleanup()
		return nil, err
	}
	defer os.RemoveAll(startupDir)
	readyPath := filepath.Join(startupDir, "ready")
	launcherPath, err := writeTerminalLauncherScript(title, env, fullCommand, workDir, readyPath)
	if err != nil {
		cleanup()
		return nil, err
	}

	terminal, terminalArgs, err := terminalLaunchCommand(title, launcherPath)
	if err != nil {
		_ = os.Remove(launcherPath)
		cleanup()
		return nil, err
	}

	cmd := exec.Command(terminal, terminalArgs...)
	cmd.Dir = workDir
	if err := startTerminalAndWait(cmd, readyPath, filepath.Join(startupDir, "startup.log"), 8*time.Second); err != nil {
		_ = os.Remove(launcherPath)
		cleanup()
		return nil, err
	}

	return &LaunchInteractiveCommandResponse{
		Command:  fullCommand,
		Terminal: terminal,
		Title:    title,
	}, nil
}

// Some terminal clients exit after handing off to a desktop service. The script
// acknowledgment, rather than the client process status, confirms a terminal.
func startTerminalAndWait(cmd *exec.Cmd, readyPath, logPath string, timeout time.Duration) error {
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer logFile.Close()
	cmd.Stdout, cmd.Stderr = logFile, logFile
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start terminal: %w", err)
	}
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	poll := time.NewTicker(50 * time.Millisecond)
	defer poll.Stop()
	diagnostic := func(reason string) error {
		data, _ := os.ReadFile(logPath)
		message := strings.TrimSpace(string(data))
		if len(message) > 2048 {
			message = message[:2048]
		}
		return fmt.Errorf("%s: %s. Start CFCBench inside the Linux desktop session as your normal user; check DISPLAY/Wayland and D-Bus access", reason, message)
	}
	for {
		if _, err := os.Stat(readyPath); err == nil {
			return nil
		}
		select {
		case err := <-exited:
			if err != nil {
				return diagnostic("terminal exited before opening the shell")
			}
			// A successful D-Bus handoff can precede the script starting.
		case <-poll.C:
		case <-timer.C:
			_ = cmd.Process.Kill()
			return diagnostic("terminal did not confirm an interactive session")
		}
	}
}

func interactiveCommandArgs(req LaunchInteractiveCommandRequest) ([]string, []string, error) {
	userPrefix := req.Username
	if req.Domain != "" {
		userPrefix = req.Domain + "/" + req.Username
	}

	args := []string{}
	if req.CommandKind == "dcomexec" {
		dcomObject := req.DCOMObject
		if dcomObject == "" {
			dcomObject = "ShellBrowserWindow"
		}
		if !isSupportedDCOMObject(dcomObject) {
			return nil, nil, errors.New("DCOM object must be ShellBrowserWindow, MMC20, or ShellWindows")
		}
		args = append(args, "-object", dcomObject)
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
		if looksLikeIPAddress(req.Target) {
			return nil, nil, fmt.Errorf("Kerberos %s requires a hostname or FQDN target; put the DC IP in -dc-ip", req.CommandKind)
		}
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

func terminalTitle(teamName string, targetLabel string, target string) string {
	box := strings.TrimSpace(targetLabel)
	if box == "" {
		box = strings.TrimSpace(target)
	}
	if box == "" {
		box = "shell"
	}
	return teamName + ":" + box
}

func terminalShellCommand(title string, env []string, command []string, workDir string, launcherPath string) string {
	lines := []string{"printf '\\033]0;%s\\007' " + shQuote(title)}
	if workDir != "" {
		lines = append(lines, "cd "+shQuote(workDir))
	}

	invocation := []string{"env"}
	for _, item := range env {
		name, value, ok := strings.Cut(item, "=")
		if !ok || name == "" {
			continue
		}
		invocation = append(invocation, name+"="+value)
	}
	invocation = append(invocation, command...)

	cleanupTargets := []string{launcherPath}
	if hostsPath := envValue(env, "NSS_WRAPPER_HOSTS"); hostsPath != "" {
		cleanupTargets = append(cleanupTargets, hostsPath)
	}
	cleanupArgs := make([]string, 0, len(cleanupTargets)+3)
	cleanupArgs = append(cleanupArgs, "rm", "-f", "--")
	for _, target := range cleanupTargets {
		if strings.TrimSpace(target) == "" {
			continue
		}
		cleanupArgs = append(cleanupArgs, target)
	}

	quoted := make([]string, 0, len(invocation))
	for _, part := range invocation {
		quoted = append(quoted, shQuote(part))
	}
	quotedCleanup := make([]string, 0, len(cleanupArgs))
	for _, part := range cleanupArgs {
		quotedCleanup = append(quotedCleanup, shQuote(part))
	}
	lines = append(lines, "trap "+shQuote(strings.Join(quotedCleanup, " ")+" >/dev/null 2>&1 || true")+" EXIT")
	if path := envValue(env, "CFCBENCH_PASSWORD_FILE"); path != "" {
		lines = append(lines, "trap "+shQuote(strings.Join(quotedCleanup, " ")+"; rm -f -- "+shQuote(path))+" EXIT")
	}
	lines = append(lines, strings.Join(quoted, " "))
	lines = append(lines, "status=$?")
	if path := envValue(env, "CFCBENCH_PASSWORD_FILE"); path != "" {
		lines = append(lines, "rm -f -- "+shQuote(path))
	}
	lines = append(lines, "echo")
	lines = append(lines, "printf "+shQuote(fmt.Sprintf("CFCBench: %s exited with status %%s\\n", commandName(command)))+" \"$status\"")
	lines = append(lines, "read -r -p 'Press Enter to close this terminal...' _")
	lines = append(lines, "exit \"$status\"")
	return strings.Join(lines, "; ")
}

func writeTerminalLauncherScript(title string, env []string, command []string, workDir string, readyPaths ...string) (string, error) {
	file, err := os.CreateTemp("", "cfc-impui-launcher-*.sh")
	if err != nil {
		return "", fmt.Errorf("failed to create terminal launcher: %w", err)
	}
	launcherPath := file.Name()
	preamble := ""
	if len(readyPaths) > 0 {
		preamble = "if ! [ -t 0 ] || ! [ -t 1 ]; then echo 'No interactive terminal attached' >&2; exit 1; fi\n" +
			"printf ready > " + shQuote(readyPaths[0]) + " || exit 1\n"
	}
	script := "#!/usr/bin/env bash\n" + preamble + terminalShellCommand(title, env, command, workDir, launcherPath) + "\n"
	if _, err := file.WriteString(script); err != nil {
		file.Close()
		_ = os.Remove(launcherPath)
		return "", fmt.Errorf("failed to write terminal launcher: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(launcherPath)
		return "", fmt.Errorf("failed to finalize terminal launcher: %w", err)
	}
	if err := os.Chmod(launcherPath, 0o700); err != nil {
		_ = os.Remove(launcherPath)
		return "", fmt.Errorf("failed to mark terminal launcher executable: %w", err)
	}
	return launcherPath, nil
}

func terminalLaunchCommand(title string, launcherPath string) (string, []string, error) {
	terminals := []terminalSpec{
		{"qterminal", func(title string, launcherPath string) []string {
			return []string{"-e", "bash", "--noprofile", "--norc", launcherPath}
		}},
		{"gnome-terminal", func(title string, launcherPath string) []string {
			return []string{"--title", title, "--", "bash", "--noprofile", "--norc", launcherPath}
		}},
		{"xfce4-terminal", func(title string, launcherPath string) []string {
			return []string{"--title", title, "--command", "bash --noprofile --norc " + shQuote(launcherPath)}
		}},
		{"mate-terminal", func(title string, launcherPath string) []string {
			return []string{"--title", title, "--", "bash", "--noprofile", "--norc", launcherPath}
		}},
		{"konsole", func(title string, launcherPath string) []string {
			return []string{"--new-tab", "-p", "tabtitle=" + title, "-e", "bash", "--noprofile", "--norc", launcherPath}
		}},
		{"x-terminal-emulator", func(title string, launcherPath string) []string {
			return []string{"-e", "bash", "--noprofile", "--norc", launcherPath}
		}},
		{"xterm", func(title string, launcherPath string) []string {
			return []string{"-T", title, "-e", "bash", "--noprofile", "--norc", launcherPath}
		}},
	}

	for _, terminal := range terminals {
		path, err := exec.LookPath(terminal.name)
		if err == nil {
			return path, terminal.args(title, launcherPath), nil
		}
	}

	return "", nil, errors.New("no supported terminal found; install qterminal, gnome-terminal, xfce4-terminal, konsole, mate-terminal, x-terminal-emulator, or xterm")
}

func shQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func envValue(env []string, key string) string {
	for _, item := range env {
		name, value, ok := strings.Cut(item, "=")
		if ok && name == key {
			return value
		}
	}
	return ""
}

func isInteractiveCommandKind(value string) bool {
	return value == "wmiexec" || value == "smbexec" || value == "dcomexec"
}

func isSupportedDCOMObject(value string) bool {
	return value == "ShellBrowserWindow" || value == "MMC20" || value == "ShellWindows"
}

func commandName(command []string) string {
	if len(command) == 0 {
		return "command"
	}
	return command[0]
}
