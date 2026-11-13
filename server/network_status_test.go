package main

import "testing"

func TestParseNmapOpenPortsKeepsOnlyTrackedOpenPorts(t *testing.T) {
	target := Target{
		ID:       42,
		TeamName: "blue",
		IP:       "10.0.0.5",
	}
	output := []byte(`
<nmaprun>
  <host>
    <address addr="10.0.0.5" addrtype="ipv4"/>
    <ports>
      <port protocol="tcp" portid="445"><state state="open"/><service name="microsoft-ds"/></port>
      <port protocol="tcp" portid="3389"><state state="open"/><service name="ms-wbt-server"/></port>
      <port protocol="tcp" portid="5985"><state state="closed"/><service name="wsman"/></port>
      <port protocol="tcp" portid="22"><state state="open"/><service name="ssh"/></port>
      <port protocol="tcp" portid="443"><state state="open"/><service name="https"/></port>
      <port protocol="tcp" portid="5986"><state state="open"/><service name="wsmans"/></port>
      <port protocol="tcp" portid="2049"><state state="filtered"/><service name="nfs"/></port>
      <port protocol="tcp" portid="65000"><state state="open"/></port>
    </ports>
  </host>
</nmaprun>`)

	ports, err := parseNmapOpenPorts(output, map[string]Target{target.IP: target})
	if err != nil {
		t.Fatalf("parse nmap output: %v", err)
	}
	if len(ports) != 5 {
		t.Fatalf("expected 5 tracked open ports, got %d: %+v", len(ports), ports)
	}
	if ports[0].Port != 445 || ports[0].Service != "microsoft-ds" {
		t.Fatalf("unexpected first port: %+v", ports[0])
	}
	if ports[1].Port != 3389 || ports[1].Service != "ms-wbt-server" {
		t.Fatalf("unexpected second port: %+v", ports[1])
	}
	for index, expected := range []int{22, 443, 5986} {
		if ports[index+2].Port != expected {
			t.Fatalf("expected port %d, got %+v", expected, ports[index+2])
		}
	}
}

func TestTrackedPortListIncludesCrossPlatformServices(t *testing.T) {
	for _, port := range []int{22, 80, 443, 445, 2049, 3389, 5432, 5985, 5986, 8080} {
		if !isTrackedPort(port) {
			t.Errorf("missing common TCP port %d", port)
		}
	}
	seen := map[int]bool{}
	for _, port := range trackedPorts {
		if seen[port] || port < 1 || port > 65535 {
			t.Fatalf("invalid or duplicate port %d", port)
		}
		seen[port] = true
	}
}

func TestReplaceTargetOpenPortsShowsOnlyLatestOpenFindings(t *testing.T) {
	setupTestDB(t)

	if _, err := CreateNewTeam(CreateTeamRequest{Name: "blue", SubnetId: 1}); err != nil {
		t.Fatalf("create team: %v", err)
	}
	target, err := CreateTarget("blue", CreateTargetRequest{
		Hostname: "dc01",
		IP:       "10.0.0.5",
		OS:       "windows",
	})
	if err != nil {
		t.Fatalf("create target: %v", err)
	}

	firstPorts := []OpenPort{
		{TeamName: "blue", TargetID: target.ID, TargetIP: target.IP, Port: 445, Service: "microsoft-ds"},
		{TeamName: "blue", TargetID: target.ID, TargetIP: target.IP, Port: 3389, Service: "ms-wbt-server"},
	}
	if err := ReplaceTargetOpenPorts("blue", []Target{*target}, firstPorts, ""); err != nil {
		t.Fatalf("replace first ports: %v", err)
	}

	secondPorts := []OpenPort{
		{TeamName: "blue", TargetID: target.ID, TargetIP: target.IP, Port: 445, Service: "microsoft-ds"},
	}
	if err := ReplaceTargetOpenPorts("blue", []Target{*target}, secondPorts, ""); err != nil {
		t.Fatalf("replace second ports: %v", err)
	}

	status, err := GetTeamNetworkStatus("blue")
	if err != nil {
		t.Fatalf("get network status: %v", err)
	}
	if len(status.Targets) != 1 {
		t.Fatalf("expected one target, got %+v", status.Targets)
	}
	if len(status.Targets[0].OpenPorts) != 1 {
		t.Fatalf("expected one latest open port, got %+v", status.Targets[0].OpenPorts)
	}
	if status.Targets[0].OpenPorts[0].Port != 445 {
		t.Fatalf("expected 445 to remain open, got %+v", status.Targets[0].OpenPorts[0])
	}
}
