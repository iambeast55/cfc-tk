# CFCBench Linux Package

This is a source deployment package. Installation builds the app on your Linux
machine and requires internet access for npm and Go dependencies.

## Requirements

- Linux, Node.js 22+, npm, Go 1.24+, Python 3, and Bash
- Impacket installed with its commands available in PATH
- nmap for network-status port scans
- libnss-wrapper (package name on Debian/Ubuntu/Kali)
- For interactive shells: a graphical desktop and a supported terminal such as
  qterminal, gnome-terminal, xfce4-terminal, konsole, mate-terminal, or xterm

Use your distribution's package manager to install system dependencies. Run the
installer and app as your normal desktop user. If Impacket is in a virtual
environment, activate it before starting CFCBench.

SSH and SCP require openssh-client. To supply a password in the Linux tab,
install sshpass; otherwise leave it blank and type it in the opened terminal.
Private keys and transfer paths refer to files on the machine running CFCBench.
Encrypted private keys prompt for their passphrase in the terminal. Host-key
verification uses the normal OpenSSH known_hosts behavior.

Stop polling in Network disables saved automatic polling for all teams and
cancels the active background scan. Manual Scan now remains available.

## Install and Run

```sh
tar -xzf cfcbench-linux-0.0.1.tar.gz
cd cfcbench-linux-0.0.1
./install.sh
./run.sh
```

Open http://127.0.0.1:5173 in a browser on the same Linux machine. Stop both
services with Ctrl+C. The launcher uses existing builds, so rebuild with
./install.sh after source changes.

Interactive terminals open on the Linux machine running the API, not on a
remote browser's machine. Start ./run.sh from your graphical desktop terminal
as the logged-in user. Starting from a headless SSH session or a service may
not provide the display and D-Bus session needed to open a window. Startup
failures are returned to the UI; success requires the launcher to reach a TTY.

The UI port can be changed with CFC_TK_UI_PORT. The API uses port 8765.
The API currently listens on all network interfaces and has no authentication;
restrict inbound access to port 8765 with the host firewall. This package is
intended for a local desktop session, not a public server.

## Data

No local databases, credentials, Kerberos caches, environment files, or logs
are included. A fresh server/teams.db is created at first launch. Back up that
file while the app is stopped before replacing an existing installation.

## Checks

```sh
npm run check
(cd server && go test ./...)
```
