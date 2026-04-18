package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

// getTeams retrieves all teams
func getTeams(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	teams, err := GetAllTeams()
	if err != nil {
		log.Printf("✗ Error fetching teams: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if teams == nil {
		teams = []Team{}
	}

	log.Printf("✓ Fetched %d teams", len(teams))
	json.NewEncoder(w).Encode(teams)
}

// getTeam retrieves a single team by ID
func getTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	name := vars["name"]
	if name == "" {
		log.Printf("✗ Invalid team name: %v", name)
		http.Error(w, "Invalid team name", http.StatusBadRequest)
		return
	}

	log.Printf("→ Fetching team name: %s", name)
	team, err := GetTeamByName(name)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("✗ Team not found: Name=%s", name)
			http.Error(w, "Team not found", http.StatusNotFound)
		} else {
			log.Printf("✗ Error fetching team: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	teamJSON, _ := json.Marshal(team)
	log.Printf("✓ Team fetched: %s", string(teamJSON))
	json.NewEncoder(w).Encode(team)
}

// createTeam creates a new team
func createTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("✗ Invalid request payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Printf("→ Creating team: Name=%s, subnetId=%d", req.Name, req.SubnetId)

	if req.Name == "" {
		log.Printf("✗ Team name is required")
		http.Error(w, "Team name is required", http.StatusBadRequest)
		return
	}

	if req.SubnetId <= 0 {
		log.Printf("✗ Invalid subnetId: %d", req.SubnetId)
		http.Error(w, "subnetId must be a positive integer", http.StatusBadRequest)
		return
	}

	team, err := CreateNewTeam(req)
	if err != nil {
		if errors.Is(err, ErrDuplicateTeamName) {
			http.Error(w, "Team name already exists", http.StatusConflict)
			return
		}
		if errors.Is(err, ErrDuplicateSubnetID) {
			http.Error(w, "Subnet ID already exists", http.StatusConflict)
			return
		}
		log.Printf("✗ Error creating team: %v", err)
		http.Error(w, "Failed to create team", http.StatusInternalServerError)
		return
	}

	teamJSON, _ := json.Marshal(team)
	log.Printf("✓ Team created: %s", string(teamJSON))
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(team)
}

// updateTeam updates an existing team
func updateTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req UpdateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("✗ Invalid request payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Printf("→ Updating team Name=%s: subnetId=%d", req.Name, req.SubnetId)
	team, err := UpdateTeamByName(req.Name, req)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("✗ Team not found: Name=%s", req.Name)
			http.Error(w, "Team not found", http.StatusNotFound)
		} else {
			log.Printf("✗ Error updating team: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	teamJSON, _ := json.Marshal(team)
	log.Printf("✓ Team updated: %s", string(teamJSON))
	json.NewEncoder(w).Encode(team)
}

// deleteTeam deletes a team
func deleteTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	name := vars["name"]
	if name == "" {
		log.Printf("✗ Invalid team name: %v", name)
		http.Error(w, "Invalid team name", http.StatusBadRequest)
		return
	}

	log.Printf("→ Deleting team name: %s", name)
	err := DeleteTeamByName(name)
	if err != nil {
		log.Printf("✗ Error deleting team: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✓ Team deleted: Name=%s", name)
	w.WriteHeader(http.StatusNoContent)
}

func getCredentials(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	teamName := mux.Vars(r)["name"]
	if teamName == "" {
		http.Error(w, "Team name is required", http.StatusBadRequest)
		return
	}

	credentials, err := GetCredentialsByTeamName(teamName)
	if err != nil {
		log.Printf("Error fetching credentials: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if credentials == nil {
		credentials = []Credential{}
	}

	json.NewEncoder(w).Encode(credentials)
}

func createCredential(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	teamName := mux.Vars(r)["name"]
	if teamName == "" {
		http.Error(w, "Team name is required", http.StatusBadRequest)
		return
	}

	var req CreateCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.OS != "windows" && req.OS != "linux" {
		http.Error(w, "OS must be windows or linux", http.StatusBadRequest)
		return
	}
	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}
	if req.SecretType == "" {
		http.Error(w, "Credential type is required", http.StatusBadRequest)
		return
	}
	if req.Secret == "" {
		http.Error(w, "Credential value is required", http.StatusBadRequest)
		return
	}
	if req.IP != "" && net.ParseIP(req.IP) == nil {
		http.Error(w, "IP must be a valid IPv4 or IPv6 address", http.StatusBadRequest)
		return
	}

	credential, err := CreateCredentialIfMissing(teamName, req)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Team not found", http.StatusNotFound)
			return
		}
		log.Printf("Error creating credential: %v", err)
		http.Error(w, "Failed to create credential", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(credential)
}

func runSecretsdump(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	teamName := mux.Vars(r)["name"]
	if teamName == "" {
		http.Error(w, "Team name is required", http.StatusBadRequest)
		return
	}

	var req RunSecretsdumpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if req.Target == "" {
		http.Error(w, "Target is required", http.StatusBadRequest)
		return
	}
	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	result, err := RunSecretsdump(teamName, req)
	if err != nil {
		log.Printf("Error running secretsdump: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func getKerberosCaches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	teamName := mux.Vars(r)["name"]
	if teamName == "" {
		http.Error(w, "Team name is required", http.StatusBadRequest)
		return
	}

	caches, err := GetKerberosCachesByTeamName(teamName)
	if err != nil {
		log.Printf("Error fetching Kerberos caches: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if caches == nil {
		caches = []KerberosCache{}
	}

	json.NewEncoder(w).Encode(caches)
}

func createKerberosCache(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	teamName := mux.Vars(r)["name"]
	if teamName == "" {
		http.Error(w, "Team name is required", http.StatusBadRequest)
		return
	}

	var req CreateKerberosCacheRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Domain == "" {
		http.Error(w, "Domain is required", http.StatusBadRequest)
		return
	}
	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}
	if req.Method != "getTGT" && req.Method != "ticketer" {
		http.Error(w, "Method must be getTGT or ticketer", http.StatusBadRequest)
		return
	}
	if req.CachePath == "" {
		http.Error(w, "Cache path is required", http.StatusBadRequest)
		return
	}
	if req.Method == "ticketer" && req.DomainSID == "" {
		http.Error(w, "Domain SID is required for ticketer caches", http.StatusBadRequest)
		return
	}

	cache, err := CreateKerberosCache(teamName, req)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Team not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, ErrDuplicateKerberosCache) {
			http.Error(w, "Kerberos cache already exists", http.StatusConflict)
			return
		}
		log.Printf("Error creating Kerberos cache: %v", err)
		http.Error(w, "Failed to create Kerberos cache", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cache)
}

func runKerberosTicket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	teamName := mux.Vars(r)["name"]
	if teamName == "" {
		http.Error(w, "Team name is required", http.StatusBadRequest)
		return
	}

	var req RunKerberosTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Domain == "" {
		http.Error(w, "Domain is required", http.StatusBadRequest)
		return
	}
	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}
	if req.Method != "getTGT" && req.Method != "ticketer" {
		http.Error(w, "Method must be getTGT or ticketer", http.StatusBadRequest)
		return
	}

	result, err := RunKerberosTicket(teamName, req)
	if err != nil {
		log.Printf("Error running Kerberos ticket command: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func getDomains(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	teamName := mux.Vars(r)["name"]
	if teamName == "" {
		http.Error(w, "Team name is required", http.StatusBadRequest)
		return
	}

	domains, err := GetDomainsByTeamName(teamName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if domains == nil {
		domains = []Domain{}
	}

	json.NewEncoder(w).Encode(domains)
}

func createDomain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	teamName := mux.Vars(r)["name"]
	if teamName == "" {
		http.Error(w, "Team name is required", http.StatusBadRequest)
		return
	}

	var req CreateDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "Domain name is required", http.StatusBadRequest)
		return
	}

	domain, err := CreateDomain(teamName, req)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Team not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, ErrDuplicateDomainName) {
			http.Error(w, "Domain already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create domain", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(domain)
}

func getTargets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	teamName := mux.Vars(r)["name"]
	if teamName == "" {
		http.Error(w, "Team name is required", http.StatusBadRequest)
		return
	}

	targets, err := GetTargetsByTeamName(teamName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if targets == nil {
		targets = []Target{}
	}

	json.NewEncoder(w).Encode(targets)
}

func createTarget(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	teamName := mux.Vars(r)["name"]
	if teamName == "" {
		http.Error(w, "Team name is required", http.StatusBadRequest)
		return
	}

	var req CreateTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Hostname == "" {
		http.Error(w, "Hostname is required", http.StatusBadRequest)
		return
	}
	if req.IP == "" || net.ParseIP(req.IP) == nil {
		http.Error(w, "IP must be a valid IPv4 or IPv6 address", http.StatusBadRequest)
		return
	}
	if req.OS != "windows" && req.OS != "linux" {
		http.Error(w, "OS must be windows or linux", http.StatusBadRequest)
		return
	}
	if req.DomainID != nil && req.DomainName != "" {
		http.Error(w, "Use domainId or domainName, not both", http.StatusBadRequest)
		return
	}

	target, err := CreateTarget(teamName, req)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Team or domain not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, ErrDuplicateTargetIP) {
			http.Error(w, "Target IP already exists", http.StatusConflict)
			return
		}
		if errors.Is(err, ErrDomainTeamMismatch) {
			http.Error(w, "Domain does not belong to team", http.StatusBadRequest)
			return
		}
		if errors.Is(err, ErrDuplicateDomainName) {
			http.Error(w, "Domain already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create target", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(target)
}
