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

var trackedWindowsPorts = []int{135, 139, 389, 445, 3389, 5985}
var networkScanMu sync.Mutex

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
	parts := make([]string, 0, len(trackedWindowsPorts))
	for _, port := range trackedWindowsPorts {
		parts = append(parts, strconv.Itoa(port))
	}
	return strings.Join(parts, ",")
}

func isTrackedWindowsPort(port int) bool {
	for _, tracked := range trackedWindowsPorts {
		if port == tracked {
			return true
		}
	}
	return false
}

func ScanTeamOpenPorts(teamName string) (*TeamNetworkStatus, error) {
	networkScanMu.Lock()
	defer networkScanMu.Unlock()

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

	ports, err := runNmapOpenPortScan(targets)
	if err != nil {
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

	if err := ReplaceTargetOpenPorts(teamName, targets, ports, ""); err != nil {
		return nil, err
	}

	return GetTeamNetworkStatus(teamName)
}

func runNmapOpenPortScan(targets []Target) ([]OpenPort, error) {
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

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
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
			if err != nil || !isTrackedWindowsPort(portID) {
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
			config, err := GetNetworkPollingConfig()
			if err != nil {
				log.Printf("network status poller config error: %v", err)
				time.Sleep(time.Minute)
				continue
			}

			interval := time.Duration(config.IntervalSeconds) * time.Second
			if interval < 15*time.Second {
				interval = 15 * time.Second
			}

			if config.Enabled {
				pollAllTeamsOnce()
			}

			time.Sleep(interval)
		}
	}()
}

func pollAllTeamsOnce() {
	teams, err := GetAllTeams()
	if err != nil {
		log.Printf("network status poller team load error: %v", err)
		return
	}

	for _, team := range teams {
		if _, err := ScanTeamOpenPorts(team.Name); err != nil {
			log.Printf("network status scan failed for %s: %v", team.Name, err)
		}
	}
}
