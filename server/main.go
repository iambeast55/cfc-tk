package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var debug bool

func main() {
	// Initialize database
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	if err := initDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer closeDB()

	// Create router
	router := mux.NewRouter()
	router.PathPrefix("/").Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Team endpoints
	router.HandleFunc("/api/teams", getTeams).Methods("GET")
	router.HandleFunc("/api/teams", createTeam).Methods("POST")
	router.HandleFunc("/api/teams/{name}", getTeam).Methods("GET")
	router.HandleFunc("/api/teams/{name}", updateTeam).Methods("PUT")
	router.HandleFunc("/api/teams/{name}", deleteTeam).Methods("DELETE")
	router.HandleFunc("/api/teams/{name}/credentials", getCredentials).Methods("GET")
	router.HandleFunc("/api/teams/{name}/credentials", createCredential).Methods("POST")
	router.HandleFunc("/api/teams/{name}/credentials", deleteCredentials).Methods("DELETE")
	router.HandleFunc("/api/teams/{name}/secretsdump/run", runSecretsdump).Methods("POST")
	router.HandleFunc("/api/teams/{name}/kerberos-caches", getKerberosCaches).Methods("GET")
	router.HandleFunc("/api/teams/{name}/kerberos-caches", createKerberosCache).Methods("POST")
	router.HandleFunc("/api/teams/{name}/kerberos-caches/run", runKerberosTicket).Methods("POST")
	router.HandleFunc("/api/teams/{name}/domains", getDomains).Methods("GET")
	router.HandleFunc("/api/teams/{name}/domains", createDomain).Methods("POST")
	router.HandleFunc("/api/teams/{name}/targets", getTargets).Methods("GET")
	router.HandleFunc("/api/teams/{name}/targets", createTarget).Methods("POST")
	router.HandleFunc("/api/teams/{name}/targets/{id}", deleteTarget).Methods("DELETE")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	// CORS middleware
	router.Use(corsMiddleware)
	router.Use(loggingMiddleware)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

func (lw *loggingResponseWriter) Write(b []byte) (int, error) {
	lw.body.Write(b)
	return lw.ResponseWriter.Write(b)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody string
		if r.Body != nil {
			bodyBytes, _ := io.ReadAll(r.Body)
			reqBody = string(bodyBytes)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		lw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lw, r)
		if debug {
			log.Printf("[%d] %s %s | Request: %s | Response: %s",
				lw.statusCode, r.Method, r.RequestURI, reqBody, lw.body.String())
		}
	})
}
