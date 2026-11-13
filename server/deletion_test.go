package main

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
)

func TestDomainDeletionPreservesTargetsAndIsTeamScoped(t *testing.T) {
	setupTestDB(t)
	for i, name := range []string{"blue", "red"} {
		if _, err := CreateNewTeam(CreateTeamRequest{Name: name, SubnetId: i + 1}); err != nil {
			t.Fatal(err)
		}
	}
	domain, err := CreateDomain("blue", CreateDomainRequest{Name: "corp.local"})
	if err != nil {
		t.Fatal(err)
	}
	target, err := CreateTarget("blue", CreateTargetRequest{Hostname: "dc01", IP: "10.0.0.1", OS: "windows", DomainID: &domain.ID})
	if err != nil {
		t.Fatal(err)
	}
	if err := DeleteDomainByID("red", domain.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("cross-team deletion returned %v", err)
	}
	request := httptest.NewRequest(http.MethodDelete, "/", nil)
	request = mux.SetURLVars(request, map[string]string{"name": "blue", "id": strconv.Itoa(domain.ID)})
	recorder := httptest.NewRecorder()
	deleteDomain(recorder, request)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("delete: %d %s", recorder.Code, recorder.Body.String())
	}
	remaining, err := GetTargetByID(target.ID)
	if err != nil {
		t.Fatal(err)
	}
	if remaining.DomainID != nil || remaining.DomainName != "" {
		t.Fatalf("target still attached: %+v", remaining)
	}
	if err := DeleteDomainByID("blue", domain.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("repeat delete: %v", err)
	}
}

func TestTeamDeletionCascadesOnAdditionalConnection(t *testing.T) {
	setupTestDB(t)
	if _, err := CreateNewTeam(CreateTeamRequest{Name: "blue", SubnetId: 1}); err != nil {
		t.Fatal(err)
	}
	domain, err := CreateDomain("blue", CreateDomainRequest{Name: "corp.local"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CreateTarget("blue", CreateTargetRequest{Hostname: "dc01", IP: "10.0.0.1", OS: "windows", DomainID: &domain.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateCredential("blue", CreateCredentialRequest{OS: "windows", Username: "test", SecretType: "password", Secret: "fixture"}); err != nil {
		t.Fatal(err)
	}
	// Reserve the original connection so deletion must use another pooled connection.
	conn, err := db.Conn(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := DeleteTeamByName("blue"); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"teams", "domains", "targets", "credentials"} {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Errorf("%s still has %d rows", table, count)
		}
	}
	if err := DeleteTeamByName("blue"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("missing team: %v", err)
	}
}
