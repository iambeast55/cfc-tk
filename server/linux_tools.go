package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type LinuxConnectionRequest struct {
	TargetID   int    `json:"targetId"`
	Username   string `json:"username"`
	Port       int    `json:"port"`
	AuthMode   string `json:"authMode"`
	Password   string `json:"password"`
	KeyPath    string `json:"keyPath"`
	Action     string `json:"action"`
	LocalPath  string `json:"localPath"`
	RemotePath string `json:"remotePath"`
}

func linuxToolCommand(req LinuxConnectionRequest, target Target) ([]string, error) {
	if target.OS != "linux" {
		return nil, fmt.Errorf("choose a Linux target")
	}
	if req.Username == "" || strings.HasPrefix(req.Username, "-") || strings.ContainsAny(req.Username, "@:/\\ \t\r\n\x00") {
		return nil, fmt.Errorf("enter a valid SSH username")
	}
	if net.ParseIP(target.IP) == nil {
		return nil, fmt.Errorf("target has an invalid IP")
	}
	if req.Port == 0 {
		req.Port = 22
	}
	if req.Port < 1 || req.Port > 65535 {
		return nil, fmt.Errorf("port must be between 1 and 65535")
	}
	if req.AuthMode != "password" && req.AuthMode != "key" {
		return nil, fmt.Errorf("choose password or key authentication")
	}
	tool, portFlag := "ssh", "-p"
	if req.Action != "ssh" {
		if req.Action != "upload" && req.Action != "download" {
			return nil, fmt.Errorf("choose SSH, upload, or download")
		}
		tool, portFlag = "scp", "-P"
	}
	args := []string{tool, portFlag, strconv.Itoa(req.Port)}
	if req.AuthMode == "key" {
		if !filepath.IsAbs(req.KeyPath) {
			return nil, fmt.Errorf("private key path must be absolute on the CFCBench machine")
		}
		info, err := os.Stat(req.KeyPath)
		if err != nil || !info.Mode().IsRegular() {
			return nil, fmt.Errorf("private key file does not exist or is not a regular file")
		}
		args = append(args, "-o", "IdentitiesOnly=yes", "-i", req.KeyPath)
	} else {
		args = append(args, "-o", "PubkeyAuthentication=no", "-o", "PreferredAuthentications=password,keyboard-interactive")
	}
	host := target.IP
	if req.Action == "ssh" {
		return append(args, "--", req.Username+"@"+host), nil
	}
	if !filepath.IsAbs(req.LocalPath) || strings.ContainsAny(req.LocalPath, "\x00\r\n") {
		return nil, fmt.Errorf("local file path must be absolute on the CFCBench machine")
	}
	if !strings.HasPrefix(req.RemotePath, "/") || strings.ContainsAny(req.RemotePath, "\x00\r\n") {
		return nil, fmt.Errorf("remote file path must be absolute")
	}
	if req.Action == "upload" {
		info, err := os.Stat(req.LocalPath)
		if err != nil || !info.Mode().IsRegular() {
			return nil, fmt.Errorf("upload source must be an existing regular file")
		}
	} else if info, err := os.Stat(filepath.Dir(req.LocalPath)); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("download destination directory does not exist")
	}
	if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	remote := req.Username + "@" + host + ":" + req.RemotePath
	if req.Action == "upload" {
		return append(args, "--", req.LocalPath, remote), nil
	}
	return append(args, "--", remote, req.LocalPath), nil
}

func launchLinuxConnection(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
	var req LinuxConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid connection request", 400)
		return
	}
	target, err := GetTargetByID(req.TargetID)
	if err != nil || target.TeamName != mux.Vars(r)["name"] {
		http.Error(w, "Target not found for this team", 404)
		return
	}
	command, err := linuxToolCommand(req, *target)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if _, err := exec.LookPath(command[0]); err != nil {
		http.Error(w, "Install openssh-client on the CFCBench machine", 400)
		return
	}
	passwordPath := ""
	cleanup := func() {
		if passwordPath != "" {
			_ = os.Remove(passwordPath)
		}
	}
	env := []string{}
	if req.AuthMode == "password" && req.Password != "" {
		if strings.ContainsAny(req.Password, "\r\n\x00") {
			http.Error(w, "Password must be a single line", 400)
			return
		}
		if _, err := exec.LookPath("sshpass"); err != nil {
			http.Error(w, "Install sshpass or leave the password blank to enter it in the terminal", 400)
			return
		}
		file, err := os.CreateTemp("", "cfcbench-ssh-password-*")
		if err != nil {
			http.Error(w, "Could not prepare authentication", 500)
			return
		}
		passwordPath = file.Name()
		_, writeErr := file.WriteString(req.Password + "\n")
		closeErr := file.Close()
		if writeErr != nil || closeErr != nil {
			cleanup()
			http.Error(w, "Could not prepare authentication", 500)
			return
		}
		command = append([]string{"sshpass", "-f", passwordPath}, command...)
		env = append(env, "CFCBENCH_PASSWORD_FILE="+passwordPath)
	}
	result, err := launchDesktopTerminal(terminalTitle(target.TeamName, target.Hostname, target.IP), command, env, serverWorkingDir(), cleanup)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
