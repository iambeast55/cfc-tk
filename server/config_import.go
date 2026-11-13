package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/netip"
	"strings"
)

type ImportConfig struct {
	Version int          `json:"version"`
	Teams   []ImportTeam `json:"teams"`
}

type ImportTeam struct {
	Name     string         `json:"name"`
	SubnetID int            `json:"subnetId"`
	Domains  []string       `json:"domains"`
	Targets  []ImportTarget `json:"targets"`
}

type ImportTarget struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	OS       string `json:"os"`
	Domain   string `json:"domain"`
}

type ImportSummary struct {
	TeamsAdded   int `json:"teamsAdded"`
	DomainsAdded int `json:"domainsAdded"`
	TargetsAdded int `json:"targetsAdded"`
	Skipped      int `json:"skipped"`
}

var errImportConflict = errors.New("config conflicts with existing data")

func validateImport(config *ImportConfig) error {
	if config.Version != 1 {
		return errors.New("Config version must be 1")
	}
	if len(config.Teams) == 0 {
		return errors.New("Include at least one team")
	}
	for i := range config.Teams {
		team := &config.Teams[i]
		team.Name = strings.TrimSpace(team.Name)
		if team.Name == "" || team.SubnetID <= 0 {
			return fmt.Errorf("Team %d needs a name and positive subnetId", i+1)
		}
		for j := range team.Domains {
			team.Domains[j] = strings.ToLower(strings.TrimSpace(team.Domains[j]))
			if team.Domains[j] == "" {
				return fmt.Errorf("Team %s has an empty domain", team.Name)
			}
		}
		for j := range team.Targets {
			target := &team.Targets[j]
			target.Hostname = strings.TrimSpace(target.Hostname)
			target.OS = strings.ToLower(strings.TrimSpace(target.OS))
			target.Domain = strings.ToLower(strings.TrimSpace(target.Domain))
			ip, err := netip.ParseAddr(strings.TrimSpace(target.IP))
			if err != nil || ip.Zone() != "" {
				return fmt.Errorf("Team %s target %d: invalid IP address", team.Name, j+1)
			}
			target.IP = ip.Unmap().String()
			if target.Hostname == "" || (target.OS != "windows" && target.OS != "linux") {
				return fmt.Errorf("Team %s target %d needs a hostname and OS windows or linux", team.Name, j+1)
			}
		}
	}
	return nil
}

func importConfig(config ImportConfig) (*ImportSummary, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	summary := &ImportSummary{}
	for _, team := range config.Teams {
		var subnet int
		err = tx.QueryRow("SELECT subnetId FROM teams WHERE name = ?", team.Name).Scan(&subnet)
		switch {
		case err == nil:
			if subnet != team.SubnetID {
				return nil, fmt.Errorf("%w: team %s has subnet %d", errImportConflict, team.Name, subnet)
			}
			summary.Skipped++
		case errors.Is(err, sql.ErrNoRows):
			var owner string
			err = tx.QueryRow("SELECT name FROM teams WHERE subnetId = ?", team.SubnetID).Scan(&owner)
			if err == nil {
				return nil, fmt.Errorf("%w: subnet %d belongs to %s", errImportConflict, team.SubnetID, owner)
			}
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}
			if _, err = tx.Exec("INSERT INTO teams (name, subnetId) VALUES (?, ?)", team.Name, team.SubnetID); err != nil {
				return nil, err
			}
			summary.TeamsAdded++
		default:
			return nil, err
		}
		ensureDomain := func(name string) (int, error) {
			var id int
			err := tx.QueryRow("SELECT id FROM domains WHERE team_name = ? AND name = ? COLLATE NOCASE", team.Name, name).Scan(&id)
			if err == nil {
				summary.Skipped++
				return id, nil
			}
			if !errors.Is(err, sql.ErrNoRows) {
				return 0, err
			}
			result, err := tx.Exec("INSERT INTO domains (team_name, name) VALUES (?, ?)", team.Name, name)
			if err != nil {
				return 0, err
			}
			created, err := result.LastInsertId()
			if err == nil {
				summary.DomainsAdded++
			}
			return int(created), err
		}
		domainIDs := map[string]int{}
		for _, name := range team.Domains {
			if _, ok := domainIDs[name]; ok {
				continue
			}
			id, err := ensureDomain(name)
			if err != nil {
				return nil, err
			}
			domainIDs[name] = id
		}
		for _, target := range team.Targets {
			var domainID any
			if target.Domain != "" {
				id, ok := domainIDs[target.Domain]
				if !ok {
					id, err = ensureDomain(target.Domain)
					if err != nil {
						return nil, err
					}
					domainIDs[target.Domain] = id
				}
				domainID = id
			}
			var hostname, os, domain string
			err := tx.QueryRow(`SELECT t.hostname, t.os, COALESCE(d.name, '') FROM targets t
				LEFT JOIN domains d ON d.id = t.domain_id WHERE t.team_name = ? AND t.ip = ?`, team.Name, target.IP).Scan(&hostname, &os, &domain)
			if err == nil {
				if hostname != target.Hostname || os != target.OS || !strings.EqualFold(domain, target.Domain) {
					return nil, fmt.Errorf("%w: IP %s in team %s already has different target details", errImportConflict, target.IP, team.Name)
				}
				summary.Skipped++
				continue
			}
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}
			if _, err = tx.Exec("INSERT INTO targets (team_name, hostname, ip, os, domain_id) VALUES (?, ?, ?, ?, ?)", team.Name, target.Hostname, target.IP, target.OS, domainID); err != nil {
				return nil, err
			}
			summary.TargetsAdded++
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return summary, nil
}

func importWorkspaceConfig(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 2*1024*1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var config ImportConfig
	if err := decoder.Decode(&config); err != nil {
		http.Error(w, "Invalid config JSON (maximum 2 MB): "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		http.Error(w, "Config must contain one JSON object", http.StatusBadRequest)
		return
	}
	if err := validateImport(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	summary, err := importConfig(config)
	if errors.Is(err, errImportConflict) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	if err != nil {
		log.Printf("Config import failed: %v", err)
		http.Error(w, "Could not import config; no changes saved", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
