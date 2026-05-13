package main

import (
	"database/sql"
	"os"
	"testing"
)

func setupTestDB(t *testing.T) {
	t.Helper()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("change to temp dir: %v", err)
	}
	if err := initDB(); err != nil {
		t.Fatalf("init test db: %v", err)
	}

	t.Cleanup(func() {
		if err := closeDB(); err != nil {
			t.Fatalf("close test db: %v", err)
		}
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
}

func TestTeamNotesPersistAndStayIsolated(t *testing.T) {
	setupTestDB(t)

	if _, err := CreateNewTeam(CreateTeamRequest{Name: "blue-one", SubnetId: 1}); err != nil {
		t.Fatalf("create first team: %v", err)
	}
	if _, err := CreateNewTeam(CreateTeamRequest{Name: "blue-two", SubnetId: 2}); err != nil {
		t.Fatalf("create second team: %v", err)
	}

	empty, err := GetTeamNoteByTeamName("blue-one")
	if err != nil {
		t.Fatalf("get empty note: %v", err)
	}
	if empty.TeamName != "blue-one" || empty.Content != "" || empty.UpdatedAt != "" {
		t.Fatalf("unexpected empty note: %+v", empty)
	}

	if _, err := SaveTeamNote("blue-one", SaveTeamNoteRequest{Content: "Initial foothold on 10.0.1.5"}); err != nil {
		t.Fatalf("save first note: %v", err)
	}
	if _, err := SaveTeamNote("blue-two", SaveTeamNoteRequest{Content: "Watch domain admin sessions"}); err != nil {
		t.Fatalf("save second note: %v", err)
	}

	first, err := GetTeamNoteByTeamName("blue-one")
	if err != nil {
		t.Fatalf("get first note: %v", err)
	}
	second, err := GetTeamNoteByTeamName("blue-two")
	if err != nil {
		t.Fatalf("get second note: %v", err)
	}
	if first.Content != "Initial foothold on 10.0.1.5" {
		t.Fatalf("unexpected first note content: %+v", first)
	}
	if second.Content != "Watch domain admin sessions" {
		t.Fatalf("unexpected second note content: %+v", second)
	}

	updated, err := SaveTeamNote("blue-one", SaveTeamNoteRequest{Content: "Updated operator notes"})
	if err != nil {
		t.Fatalf("update first note: %v", err)
	}
	if updated.Content != "Updated operator notes" || updated.UpdatedAt == "" {
		t.Fatalf("unexpected updated note: %+v", updated)
	}
}

func TestTeamNotesRejectMissingTeam(t *testing.T) {
	setupTestDB(t)

	if _, err := GetTeamNoteByTeamName("missing"); err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows loading missing team note, got %v", err)
	}
	if _, err := SaveTeamNote("missing", SaveTeamNoteRequest{Content: "nope"}); err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows saving missing team note, got %v", err)
	}
}

func TestTeamNotesCascadeWhenTeamDeleted(t *testing.T) {
	setupTestDB(t)

	if _, err := CreateNewTeam(CreateTeamRequest{Name: "blue-one", SubnetId: 1}); err != nil {
		t.Fatalf("create team: %v", err)
	}
	if _, err := SaveTeamNote("blue-one", SaveTeamNoteRequest{Content: "temporary notes"}); err != nil {
		t.Fatalf("save note: %v", err)
	}
	if err := DeleteTeamByName("blue-one"); err != nil {
		t.Fatalf("delete team: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM team_notes WHERE team_name = ?", "blue-one").Scan(&count); err != nil {
		t.Fatalf("count note rows: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected deleted team notes to cascade, found %d rows", count)
	}
	if _, err := GetTeamNoteByTeamName("blue-one"); err != sql.ErrNoRows {
		t.Fatalf("expected missing team after delete, got %v", err)
	}
}
