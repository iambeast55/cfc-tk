package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

var db *sql.DB
var ErrDuplicateTeamName = errors.New("team name already exists")
var ErrDuplicateSubnetID = errors.New("subnet ID already exists")
var ErrDuplicateDomainName = errors.New("domain already exists")
var ErrDuplicateTargetIP = errors.New("target IP already exists")
var ErrDomainTeamMismatch = errors.New("domain does not belong to team")
var ErrDuplicateKerberosCache = errors.New("kerberos cache already exists")

func initDB() error {
	var err error
	db, err = sql.Open("sqlite", "file:teams.db?cache=shared")
	if err != nil {
		return err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return err
	}

	createTeamsTableSQL := `
	CREATE TABLE IF NOT EXISTS teams (
		subnetId INTEGER NOT NULL UNIQUE,
		name TEXT NOT NULL PRIMARY KEY,
		targets TEXT DEFAULT '[]'
	);
	`

	if _, err = db.Exec(createTeamsTableSQL); err != nil {
		return err
	}

	createCredentialsTableSQL := `
	CREATE TABLE IF NOT EXISTS credentials (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		team_name TEXT NOT NULL,
		os TEXT NOT NULL,
		username TEXT NOT NULL,
		secret_type TEXT NOT NULL,
		secret TEXT NOT NULL,
		rid TEXT DEFAULT '',
		domain TEXT DEFAULT '',
		host TEXT DEFAULT '',
		ip TEXT DEFAULT '',
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(team_name) REFERENCES teams(name) ON DELETE CASCADE
	);
	`

	if _, err = db.Exec(createCredentialsTableSQL); err != nil {
		return err
	}

	if err := ensureCredentialColumn("ip", "TEXT DEFAULT ''"); err != nil {
		return err
	}
	if err := ensureCredentialColumn("rid", "TEXT DEFAULT ''"); err != nil {
		return err
	}

	createDomainsTableSQL := `
	CREATE TABLE IF NOT EXISTS domains (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		team_name TEXT NOT NULL,
		name TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(team_name, name),
		FOREIGN KEY(team_name) REFERENCES teams(name) ON DELETE CASCADE
	);
	`

	if _, err = db.Exec(createDomainsTableSQL); err != nil {
		return err
	}

	createTargetsTableSQL := `
	CREATE TABLE IF NOT EXISTS targets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		team_name TEXT NOT NULL,
		domain_id INTEGER NULL,
		hostname TEXT NOT NULL,
		ip TEXT NOT NULL,
		os TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(team_name, ip),
		FOREIGN KEY(team_name) REFERENCES teams(name) ON DELETE CASCADE,
		FOREIGN KEY(domain_id) REFERENCES domains(id) ON DELETE SET NULL
	);
	`

	if _, err = db.Exec(createTargetsTableSQL); err != nil {
		return err
	}

	createKerberosCachesTableSQL := `
	CREATE TABLE IF NOT EXISTS kerberos_caches (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		team_name TEXT NOT NULL,
		domain TEXT NOT NULL,
		username TEXT NOT NULL,
		method TEXT NOT NULL,
		cache_path TEXT NOT NULL,
		kdc_host TEXT DEFAULT '',
		domain_sid TEXT DEFAULT '',
		user_id TEXT DEFAULT '',
		groups TEXT DEFAULT '',
		extra_sid TEXT DEFAULT '',
		duration TEXT DEFAULT '',
		expires_at TEXT DEFAULT '',
		notes TEXT DEFAULT '',
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(team_name, domain, username, method, cache_path),
		FOREIGN KEY(team_name) REFERENCES teams(name) ON DELETE CASCADE
	);
	`

	if _, err = db.Exec(createKerberosCachesTableSQL); err != nil {
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}

func ensureCredentialColumn(name string, definition string) error {
	rows, err := db.Query("PRAGMA table_info(credentials)")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var columnName string
		var columnType string
		var notNull int
		var defaultValue sql.NullString
		var primaryKey int
		if err := rows.Scan(&cid, &columnName, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return err
		}
		if columnName == name {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = db.Exec("ALTER TABLE credentials ADD COLUMN " + name + " " + definition)
	return err
}

func closeDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// GetAllTeams retrieves all teams from the database
func GetAllTeams() ([]Team, error) {
	rows, err := db.Query("SELECT subnetId, name, targets FROM teams ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []Team
	for rows.Next() {
		var team Team
		var targetsJSON string
		if err := rows.Scan(&team.SubnetId, &team.Name, &targetsJSON); err != nil {
			return nil, err
		}
		// Parse JSON targets
		if targetsJSON != "" && targetsJSON != "[]" {
			json.Unmarshal([]byte(targetsJSON), &team.Targets)
		} else {
			team.Targets = []string{}
		}
		teams = append(teams, team)
	}

	return teams, nil
}

// GetTeamByID retrieves a single team by ID
func GetTeamByName(name string) (*Team, error) {
	var team Team
	var targetsJSON string
	err := db.QueryRow("SELECT subnetId, name, targets FROM teams WHERE name = ?", name).
		Scan(&team.SubnetId, &team.Name, &targetsJSON)
	if err != nil {
		return nil, err
	}
	// Parse JSON targets
	if targetsJSON != "" && targetsJSON != "[]" {
		json.Unmarshal([]byte(targetsJSON), &team.Targets)
	} else {
		team.Targets = []string{}
	}
	return &team, nil
}

// CreateNewTeam inserts a new team into the database
func CreateNewTeam(req CreateTeamRequest) (*Team, error) {
	var exists int
	if err := db.QueryRow("SELECT COUNT(*) FROM teams WHERE name = ?", req.Name).Scan(&exists); err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, ErrDuplicateTeamName
	}

	if err := db.QueryRow("SELECT COUNT(*) FROM teams WHERE subnetId = ?", req.SubnetId).Scan(&exists); err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, ErrDuplicateSubnetID
	}

	_, err := db.Exec(
		"INSERT INTO teams (subnetId, name, targets) VALUES (?, ?, ?)",
		req.SubnetId, req.Name, "[]",
	)
	if err != nil {
		return nil, err
	}

	return &Team{
		SubnetId: req.SubnetId,
		Name:     req.Name,
		Targets:  []string{},
	}, nil
}

// UpdateTeamByID updates an existing team
func UpdateTeamByName(name string, req UpdateTeamRequest) (*Team, error) {
	_, err := db.Exec(
		"UPDATE teams SET subnetId = ? WHERE name = ?",
		req.SubnetId, name,
	)
	if err != nil {
		return nil, err
	}

	return GetTeamByName(name)
}

// DeleteTeamByName deletes a team by name
func DeleteTeamByName(name string) error {
	_, err := db.Exec("DELETE FROM teams WHERE name = ?", name)
	return err
}

func GetCredentialsByTeamName(teamName string) ([]Credential, error) {
	rows, err := db.Query(`
		SELECT id, team_name, os, username, secret_type, secret, rid, domain, host, ip, created_at
		FROM credentials
		WHERE team_name = ?
		ORDER BY created_at DESC, id DESC
	`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credentials []Credential
	for rows.Next() {
		var credential Credential
		if err := rows.Scan(
			&credential.ID,
			&credential.TeamName,
			&credential.OS,
			&credential.Username,
			&credential.SecretType,
			&credential.Secret,
			&credential.RID,
			&credential.Domain,
			&credential.Host,
			&credential.IP,
			&credential.CreatedAt,
		); err != nil {
			return nil, err
		}
		credentials = append(credentials, credential)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return credentials, nil
}

func CreateCredential(teamName string, req CreateCredentialRequest) (*Credential, error) {
	if _, err := GetTeamByName(teamName); err != nil {
		return nil, err
	}

	result, err := db.Exec(`
		INSERT INTO credentials (team_name, os, username, secret_type, secret, rid, domain, host, ip)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, teamName, req.OS, req.Username, req.SecretType, req.Secret, req.RID, req.Domain, req.Host, req.IP)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	var credential Credential
	err = db.QueryRow(`
		SELECT id, team_name, os, username, secret_type, secret, rid, domain, host, ip, created_at
		FROM credentials
		WHERE id = ?
	`, id).Scan(
		&credential.ID,
		&credential.TeamName,
		&credential.OS,
		&credential.Username,
		&credential.SecretType,
		&credential.Secret,
		&credential.RID,
		&credential.Domain,
		&credential.Host,
		&credential.IP,
		&credential.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &credential, nil
}

func CreateCredentialIfMissing(teamName string, req CreateCredentialRequest) (*Credential, error) {
	var id int
	err := db.QueryRow(`
		SELECT id
		FROM credentials
		WHERE team_name = ? AND username = ? AND secret_type = ? AND secret = ?
			AND rid = ? AND domain = ? AND host = ? AND ip = ?
	`, teamName, req.Username, req.SecretType, req.Secret, req.RID, req.Domain, req.Host, req.IP).Scan(&id)
	if err == nil {
		return GetCredentialByID(id)
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	return CreateCredential(teamName, req)
}

func GetCredentialByID(id int) (*Credential, error) {
	var credential Credential
	err := db.QueryRow(`
		SELECT id, team_name, os, username, secret_type, secret, rid, domain, host, ip, created_at
		FROM credentials
		WHERE id = ?
	`, id).Scan(
		&credential.ID,
		&credential.TeamName,
		&credential.OS,
		&credential.Username,
		&credential.SecretType,
		&credential.Secret,
		&credential.RID,
		&credential.Domain,
		&credential.Host,
		&credential.IP,
		&credential.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &credential, nil
}

func GetDomainsByTeamName(teamName string) ([]Domain, error) {
	rows, err := db.Query(`
		SELECT id, team_name, name, created_at
		FROM domains
		WHERE team_name = ?
		ORDER BY name
	`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []Domain
	for rows.Next() {
		var domain Domain
		if err := rows.Scan(&domain.ID, &domain.TeamName, &domain.Name, &domain.CreatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, domain)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return domains, nil
}

func CreateDomain(teamName string, req CreateDomainRequest) (*Domain, error) {
	if _, err := GetTeamByName(teamName); err != nil {
		return nil, err
	}

	var exists int
	if err := db.QueryRow("SELECT COUNT(*) FROM domains WHERE team_name = ? AND name = ?", teamName, req.Name).Scan(&exists); err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, ErrDuplicateDomainName
	}

	result, err := db.Exec("INSERT INTO domains (team_name, name) VALUES (?, ?)", teamName, req.Name)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetDomainByID(int(id))
}

func GetDomainByID(id int) (*Domain, error) {
	var domain Domain
	err := db.QueryRow(`
		SELECT id, team_name, name, created_at
		FROM domains
		WHERE id = ?
	`, id).Scan(&domain.ID, &domain.TeamName, &domain.Name, &domain.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &domain, nil
}

func GetOrCreateDomainByName(teamName string, name string) (*Domain, error) {
	var domain Domain
	err := db.QueryRow(`
		SELECT id, team_name, name, created_at
		FROM domains
		WHERE team_name = ? AND name = ?
	`, teamName, name).Scan(&domain.ID, &domain.TeamName, &domain.Name, &domain.CreatedAt)
	if err == nil {
		return &domain, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	return CreateDomain(teamName, CreateDomainRequest{Name: name})
}

func GetTargetsByTeamName(teamName string) ([]Target, error) {
	rows, err := db.Query(`
		SELECT targets.id, targets.team_name, targets.domain_id, COALESCE(domains.name, ''),
			targets.hostname, targets.ip, targets.os, targets.created_at
		FROM targets
		LEFT JOIN domains ON domains.id = targets.domain_id
		WHERE targets.team_name = ?
		ORDER BY COALESCE(domains.name, ''), targets.hostname
	`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []Target
	for rows.Next() {
		var target Target
		var domainID sql.NullInt64
		if err := rows.Scan(
			&target.ID,
			&target.TeamName,
			&domainID,
			&target.DomainName,
			&target.Hostname,
			&target.IP,
			&target.OS,
			&target.CreatedAt,
		); err != nil {
			return nil, err
		}
		if domainID.Valid {
			id := int(domainID.Int64)
			target.DomainID = &id
		}
		targets = append(targets, target)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return targets, nil
}

func CreateTarget(teamName string, req CreateTargetRequest) (*Target, error) {
	if _, err := GetTeamByName(teamName); err != nil {
		return nil, err
	}

	var exists int
	if err := db.QueryRow("SELECT COUNT(*) FROM targets WHERE team_name = ? AND ip = ?", teamName, req.IP).Scan(&exists); err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, ErrDuplicateTargetIP
	}

	var domainID any
	if req.DomainID != nil {
		domain, err := GetDomainByID(*req.DomainID)
		if err != nil {
			return nil, err
		}
		if domain.TeamName != teamName {
			return nil, ErrDomainTeamMismatch
		}
		domainID = domain.ID
	} else if req.DomainName != "" {
		domain, err := GetOrCreateDomainByName(teamName, req.DomainName)
		if err != nil {
			return nil, err
		}
		domainID = domain.ID
	}

	result, err := db.Exec(`
		INSERT INTO targets (team_name, domain_id, hostname, ip, os)
		VALUES (?, ?, ?, ?, ?)
	`, teamName, domainID, req.Hostname, req.IP, req.OS)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetTargetByID(int(id))
}

func GetTargetByID(id int) (*Target, error) {
	var target Target
	var domainID sql.NullInt64
	err := db.QueryRow(`
		SELECT targets.id, targets.team_name, targets.domain_id, COALESCE(domains.name, ''),
			targets.hostname, targets.ip, targets.os, targets.created_at
		FROM targets
		LEFT JOIN domains ON domains.id = targets.domain_id
		WHERE targets.id = ?
	`, id).Scan(
		&target.ID,
		&target.TeamName,
		&domainID,
		&target.DomainName,
		&target.Hostname,
		&target.IP,
		&target.OS,
		&target.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if domainID.Valid {
		id := int(domainID.Int64)
		target.DomainID = &id
	}

	return &target, nil
}

func GetKerberosCachesByTeamName(teamName string) ([]KerberosCache, error) {
	rows, err := db.Query(`
		SELECT id, team_name, domain, username, method, cache_path, kdc_host, domain_sid,
			user_id, groups, extra_sid, duration, expires_at, notes, created_at
		FROM kerberos_caches
		WHERE team_name = ?
		ORDER BY created_at DESC, id DESC
	`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var caches []KerberosCache
	for rows.Next() {
		var cache KerberosCache
		if err := rows.Scan(
			&cache.ID,
			&cache.TeamName,
			&cache.Domain,
			&cache.Username,
			&cache.Method,
			&cache.CachePath,
			&cache.KDCHost,
			&cache.DomainSID,
			&cache.UserID,
			&cache.Groups,
			&cache.ExtraSID,
			&cache.Duration,
			&cache.ExpiresAt,
			&cache.Notes,
			&cache.CreatedAt,
		); err != nil {
			return nil, err
		}
		cache.Status = kerberosCacheStatus(cache.CachePath, cache.ExpiresAt)
		caches = append(caches, cache)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return caches, nil
}

func CreateKerberosCache(teamName string, req CreateKerberosCacheRequest) (*KerberosCache, error) {
	if _, err := GetTeamByName(teamName); err != nil {
		return nil, err
	}

	var exists int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM kerberos_caches
		WHERE team_name = ? AND domain = ? AND username = ? AND method = ? AND cache_path = ?
	`, teamName, req.Domain, req.Username, req.Method, req.CachePath).Scan(&exists); err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, ErrDuplicateKerberosCache
	}

	result, err := db.Exec(`
		INSERT INTO kerberos_caches (
			team_name, domain, username, method, cache_path, kdc_host, domain_sid,
			user_id, groups, extra_sid, duration, expires_at, notes
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, teamName, req.Domain, req.Username, req.Method, req.CachePath, req.KDCHost, req.DomainSID,
		req.UserID, req.Groups, req.ExtraSID, req.Duration, req.ExpiresAt, req.Notes)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetKerberosCacheByID(int(id))
}

func GetKerberosCacheByID(id int) (*KerberosCache, error) {
	var cache KerberosCache
	err := db.QueryRow(`
		SELECT id, team_name, domain, username, method, cache_path, kdc_host, domain_sid,
			user_id, groups, extra_sid, duration, expires_at, notes, created_at
		FROM kerberos_caches
		WHERE id = ?
	`, id).Scan(
		&cache.ID,
		&cache.TeamName,
		&cache.Domain,
		&cache.Username,
		&cache.Method,
		&cache.CachePath,
		&cache.KDCHost,
		&cache.DomainSID,
		&cache.UserID,
		&cache.Groups,
		&cache.ExtraSID,
		&cache.Duration,
		&cache.ExpiresAt,
		&cache.Notes,
		&cache.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	cache.Status = kerberosCacheStatus(cache.CachePath, cache.ExpiresAt)

	return &cache, nil
}

func GetKerberosCacheByKey(teamName string, domain string, username string, method string, cachePath string) (*KerberosCache, error) {
	var id int
	err := db.QueryRow(`
		SELECT id
		FROM kerberos_caches
		WHERE team_name = ? AND domain = ? AND username = ? AND method = ? AND cache_path = ?
	`, teamName, domain, username, method, cachePath).Scan(&id)
	if err != nil {
		return nil, err
	}

	return GetKerberosCacheByID(id)
}

func kerberosCacheStatus(cachePath string, expiresAt string) string {
	if cachePath == "" {
		return "unknown"
	}
	if expiresAt != "" {
		for _, layout := range []string{time.RFC3339, "2006-01-02 15:04", "2006-01-02"} {
			expires, err := time.Parse(layout, expiresAt)
			if err == nil && time.Now().After(expires) {
				return "expired"
			}
		}
	}
	for _, candidate := range kerberosCachePathCandidates(cachePath) {
		if _, err := os.Stat(candidate); err == nil {
			return "available"
		} else if !os.IsNotExist(err) {
			return "unknown"
		}
	}
	return "missing_file"
}

func kerberosCachePathCandidates(cachePath string) []string {
	if filepath.IsAbs(cachePath) {
		return []string{cachePath}
	}
	return []string{
		cachePath,
		filepath.Join("..", cachePath),
	}
}
