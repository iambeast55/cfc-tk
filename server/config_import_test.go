package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func importFixture(body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/config/import", strings.NewReader(body))
	recorder := httptest.NewRecorder()
	importWorkspaceConfig(recorder, request)
	return recorder
}

func TestImportConfigCreatesAndReimportsWithoutDuplicates(t *testing.T) {
	setupTestDB(t)
	body := `{"version":1,"teams":[{"name":"blue","subnetId":1,"domains":["CORP.EXAMPLE"],"targets":[
		{"hostname":"dc01","ip":"192.0.2.1","os":"windows","domain":"corp.example"},
		{"hostname":"web01","ip":"2001:db8::1","os":"linux"}]}]}`
	for i := 0; i < 2; i++ {
		r := importFixture(body)
		if r.Code != 200 {
			t.Fatalf("import %d: %d %s", i, r.Code, r.Body.String())
		}
	}
	targets, err := GetTargetsByTeamName("blue")
	if err != nil || len(targets) != 2 {
		t.Fatalf("targets: %+v %v", targets, err)
	}
	domains, err := GetDomainsByTeamName("blue")
	if err != nil || len(domains) != 1 {
		t.Fatalf("domains: %+v %v", domains, err)
	}
	for _, target := range targets {
		if target.Hostname == "dc01" && target.DomainName != "corp.example" {
			t.Fatal("domain missing")
		}
		if target.Hostname == "web01" && target.DomainID != nil {
			t.Fatal("standalone assigned a domain")
		}
	}
}

func TestImportConflictRollsBackEarlierTeams(t *testing.T) {
	setupTestDB(t)
	if _, err := CreateNewTeam(CreateTeamRequest{Name: "existing", SubnetId: 2}); err != nil {
		t.Fatal(err)
	}
	r := importFixture(`{"version":1,"teams":[{"name":"new","subnetId":1,"domains":["example.test"]},{"name":"conflict","subnetId":2}]}`)
	if r.Code != 409 {
		t.Fatalf("expected conflict: %d %s", r.Code, r.Body.String())
	}
	teams, err := GetAllTeams()
	if err != nil || len(teams) != 1 || teams[0].Name != "existing" {
		t.Fatalf("partial import: %+v %v", teams, err)
	}
	domains, err := GetDomainsByTeamName("new")
	if err != nil || len(domains) != 0 {
		t.Fatalf("orphan domains: %+v %v", domains, err)
	}
}

func TestImportRejectsInvalidConfigBeforeWriting(t *testing.T) {
	setupTestDB(t)
	for _, body := range []string{
		`{"version":1,"teams":[{"name":"blue","subnetId":1,"targets":[{"hostname":"bad","ip":"999.1.1.1","os":"linux"}]}]}`,
		`{"version":2,"teams":[{"name":"blue","subnetId":1}]}`,
		`{"version":1,"teams":[{"name":"blue","subnetId":1,"typo":true}]}`,
		`{"version":1,"teams":[{"name":"blue","subnetId":1}]} {}`,
	} {
		r := importFixture(body)
		if r.Code != 400 {
			t.Fatalf("expected validation error: %d %s", r.Code, r.Body.String())
		}
	}
	teams, err := GetAllTeams()
	if err != nil || len(teams) != 0 {
		t.Fatalf("invalid input wrote data: %+v %v", teams, err)
	}
}
