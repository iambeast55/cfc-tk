package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Scan the same service set for either OS; services are not OS-exclusive.
var trackedPorts = []int{
	21, 22, 25, 53, 80, 88, 110, 111, 135, 139, 143, 389, 443, 445,
	464, 465, 587, 636, 993, 995, 1433, 2049, 3268, 3269, 3306,
	3389, 5432, 5900, 5985, 5986, 6379, 8000, 8080, 8443, 9200, 27017,
}
var networkScanMu sync.Mutex
var pollingMu sync.Mutex
var pollingCancel context.CancelFunc
var pollingWake = make(chan struct{}, 1)

func resetNetworkPolling() {
	pollingMu.Lock()
	if pollingCancel != nil {
		pollingCancel()
	}
	pollingMu.Unlock()
	select {
	case pollingWake <- struct{}{}:
	default:
	}
}

type nmapRun struct {
	Hosts []nmapHost `xml:"host"`
}

type nmapHost struct {
	Addresses []nmapAddress `xml:"address"`
	Ports     []nmapPort    `xml:"ports>port"`
}

type nmapAddress struct {
	Addr string `xml:"addr,attr"`
}

type nmapPort struct {
	PortID  string      `xml:"portid,attr"`
	State   nmapState   `xml:"state"`
	Service nmapService `xml:"service"`
}

type nmapState struct {
	State string `xml:"state,attr"`
}

type nmapService struct {
	Name string `xml:"name,attr"`
}

func trackedPortList() string {
	parts := make([]string, 0, len(trackedPorts))
	for _, port := range trackedPorts {
		parts = append(parts, strconv.Itoa(port))
	}
	return strings.Join(parts, ",")
}

func isTrackedPort(port int) bool {
	for _, tracked := range trackedPorts {
		if port == tracked {
			return true
		}
	}
	return false
}

func ScanTeamOpenPorts(teamName string) (*TeamNetworkStatus, error) {
	return scanTeamOpenPorts(context.Background(), teamName)
}

func scanTeamOpenPorts(ctx context.Context, teamName string) (*TeamNetworkStatus, error) {
	networkScanMu.Lock()
	defer networkScanMu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if _, err := GetTeamByName(teamName); err != nil {
		return nil, err
	}

	targets, err := GetTargetsByTeamName(teamName)
	if err != nil {
		return nil, err
	}

	if len(targets) == 0 {
		if err := ReplaceTargetOpenPorts(teamName, targets, nil, ""); err != nil {
			return nil, err
		}
		return GetTeamNetworkStatus(teamName)
	}

	ports, err := runNmapOpenPortScan(ctx, targets)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		scanErr := err.Error()
		if updateErr := UpdateNetworkScanError(teamName, scanErr); updateErr != nil {
			return nil, updateErr
		}
		status, statusErr := GetTeamNetworkStatus(teamName)
		if statusErr != nil {
			return nil, statusErr
		}
		status.LastError = scanErr
		return status, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := ReplaceTargetOpenPorts(teamName, targets, ports, ""); err != nil {
		return nil, err
	}

	return GetTeamNetworkStatus(teamName)
}

func runNmapOpenPortScan(parent context.Context, targets []Target) ([]OpenPort, error) {
	args := []string{"-Pn", "-p", trackedPortList(), "--open", "-oX", "-"}
	targetsByIP := map[string]Target{}
	for _, target := range targets {
		ip := strings.TrimSpace(target.IP)
		if ip == "" {
			continue
		}
		targetsByIP[ip] = target
		args = append(args, ip)
	}

	if len(targetsByIP) == 0 {
		return nil, ErrNoNetworkTargets
	}

	ctx, cancel := context.WithTimeout(parent, 90*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "nmap", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("nmap scan timed out")
	}
	if err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = err.Error()
		}
		return nil, fmt.Errorf("nmap scan failed: %s", detail)
	}

	return parseNmapOpenPorts(output, targetsByIP)
}

func parseNmapOpenPorts(output []byte, targetsByIP map[string]Target) ([]OpenPort, error) {
	var parsed nmapRun
	if err := xml.Unmarshal(output, &parsed); err != nil {
		return nil, err
	}

	var ports []OpenPort
	for _, host := range parsed.Hosts {
		var target *Target
		for _, address := range host.Addresses {
			if candidate, ok := targetsByIP[address.Addr]; ok {
				target = &candidate
				break
			}
		}
		if target == nil {
			continue
		}

		for _, port := range host.Ports {
			if port.State.State != "open" {
				continue
			}
			portID, err := strconv.Atoi(port.PortID)
			if err != nil || !isTrackedPort(portID) {
				continue
			}
			ports = append(ports, OpenPort{
				TeamName: target.TeamName,
				TargetID: target.ID,
				TargetIP: target.IP,
				Port:     portID,
				Service:  port.Service.Name,
			})
		}
	}

	return ports, nil
}

func StartNetworkStatusPoller() {
	go func() {
		for {
			ctx, cancel := context.WithCancel(context.Background())
			pollingMu.Lock()
			pollingCancel = cancel
			pollingMu.Unlock()
			config, err := GetNetworkPollingConfig()
			if err != nil {
				cancel()
				log.Printf("network status poller config error: %v", err)
				time.Sleep(time.Minute)
				continue
			}

			interval := time.Duration(config.IntervalSeconds) * time.Second
			if interval < 15*time.Second {
				interval = 15 * time.Second
			}

			if config.Enabled {
				pollAllTeamsOnce(ctx)
			}
			cancel()
			timer := time.NewTimer(interval)
			select {
			case <-timer.C:
			case <-pollingWake:
			}
			timer.Stop()
		}
	}()
}

func pollAllTeamsOnce(ctx context.Context) {
	teams, err := GetAllTeams()
	if err != nil {
		log.Printf("network status poller team load error: %v", err)
		return
	}

	for _, team := range teams {
		if ctx.Err() != nil {
			return
		}
		if _, err := scanTeamOpenPorts(ctx, team.Name); err != nil && ctx.Err() == nil {
			log.Printf("network status scan failed for %s: %v", team.Name, err)
		}
	}
}
