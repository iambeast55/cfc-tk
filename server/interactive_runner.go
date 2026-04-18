package main

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
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
	shellCommand := terminalShellCommand(title, env, fullCommand)

	terminal, terminalArgs, err := terminalLaunchCommand(title, shellCommand)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(terminal, terminalArgs...)
	cmd.Dir = serverWorkingDir()
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to launch terminal: %w", err)
	}

	return &LaunchInteractiveCommandResponse{
		Command:  fullCommand,
		Terminal: terminal,
		Title:    title,
	}, nil
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

func terminalShellCommand(title string, env []string, command []string) string {
	lines := []string{"printf '\\033]0;%s\\007' " + shQuote(title)}
	for _, item := range env {
		name, value, ok := strings.Cut(item, "=")
		if !ok || name == "" {
			continue
		}
		lines = append(lines, "export "+name+"="+shQuote(value))
	}

	quoted := make([]string, 0, len(command))
	for _, part := range command {
		quoted = append(quoted, shQuote(part))
	}
	lines = append(lines, strings.Join(quoted, " "))
	lines = append(lines, "status=$?")
	lines = append(lines, "echo")
	lines = append(lines, "printf "+shQuote(fmt.Sprintf("cfc-tk: %s exited with status %%s\\n", commandName(command)))+" \"$status\"")
	lines = append(lines, "read -r -p 'Press Enter to close this terminal...' _")
	lines = append(lines, "exit \"$status\"")
	return strings.Join(lines, "; ")
}

func terminalLaunchCommand(title string, shellCommand string) (string, []string, error) {
	terminals := []terminalSpec{
		{"gnome-terminal", func(title string, shellCommand string) []string {
			return []string{"--title", title, "--", "bash", "-lc", shellCommand}
		}},
		{"xfce4-terminal", func(title string, shellCommand string) []string {
			return []string{"--title", title, "--command", "bash -lc " + shQuote(shellCommand)}
		}},
		{"mate-terminal", func(title string, shellCommand string) []string {
			return []string{"--title", title, "--", "bash", "-lc", shellCommand}
		}},
		{"konsole", func(title string, shellCommand string) []string {
			return []string{"--new-tab", "-p", "tabtitle=" + title, "-e", "bash", "-lc", shellCommand}
		}},
		{"x-terminal-emulator", func(title string, shellCommand string) []string {
			return []string{"-T", title, "-e", "bash", "-lc", shellCommand}
		}},
		{"xterm", func(title string, shellCommand string) []string {
			return []string{"-T", title, "-e", "bash", "-lc", shellCommand}
		}},
	}

	for _, terminal := range terminals {
		path, err := exec.LookPath(terminal.name)
		if err == nil {
			return path, terminal.args(title, shellCommand), nil
		}
	}

	return "", nil, errors.New("no supported terminal found; install gnome-terminal, xfce4-terminal, konsole, mate-terminal, x-terminal-emulator, or xterm")
}

func shQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
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
