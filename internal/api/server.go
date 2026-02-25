package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/hbf-agent/internal/config"
	"github.com/yourusername/hbf-agent/internal/firewall"
	"github.com/yourusername/hbf-agent/internal/servicemesh"
)

// Server represents the API server
type Server struct {
	config      *config.Config
	log         *logrus.Logger
	firewall    *firewall.Manager
	serviceMesh *servicemesh.Manager
	server      *http.Server
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, fw *firewall.Manager, sm *servicemesh.Manager, log *logrus.Logger) (*Server, error) {
	return &Server{
		config:      cfg,
		log:         log,
		firewall:    fw,
		serviceMesh: sm,
	}, nil
}

// Start starts the API server
func (s *Server) Start() error {
	mux := http.NewServeMux()
	
	// Health endpoint
	mux.HandleFunc("/api/v1/health", s.handleHealth)
	
	// Service endpoints
	mux.HandleFunc("/api/v1/services", s.handleServices)
	mux.HandleFunc("/api/v1/services/", s.handleServiceByID)
	
	// Firewall endpoints
	mux.HandleFunc("/api/v1/firewall/rules", s.handleFirewallRules)
	mux.HandleFunc("/api/v1/firewall/rules/", s.handleFirewallRuleByID)
	
	// Metrics endpoint (redirects to metrics server)
	mux.HandleFunc("/api/v1/metrics", s.handleMetrics)
	
	s.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.config.Agent.BindAddr, s.config.Agent.APIPort),
		Handler: s.loggingMiddleware(mux),
	}
	
	s.log.Infof("API server listening on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop stops the API server
func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// Middleware

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.log.Infof("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// Handlers

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	response := map[string]interface{}{
		"status": "healthy",
		"agent": map[string]string{
			"node_id":    s.config.Agent.NodeID,
			"datacenter": s.config.Agent.Datacenter,
		},
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listServices(w, r)
	case http.MethodPost:
		s.registerService(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listServices(w http.ResponseWriter, r *http.Request) {
	if s.serviceMesh == nil {
		http.Error(w, "Service mesh not enabled", http.StatusServiceUnavailable)
		return
	}
	
	services := s.serviceMesh.ListServices()
	s.writeJSON(w, http.StatusOK, services)
}

func (s *Server) registerService(w http.ResponseWriter, r *http.Request) {
	if s.serviceMesh == nil {
		http.Error(w, "Service mesh not enabled", http.StatusServiceUnavailable)
		return
	}
	
	var service servicemesh.Service
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}
	
	if err := s.serviceMesh.RegisterService(&service); err != nil {
		http.Error(w, fmt.Sprintf("Failed to register service: %v", err), http.StatusInternalServerError)
		return
	}
	
	s.writeJSON(w, http.StatusCreated, service)
}

func (s *Server) handleServiceByID(w http.ResponseWriter, r *http.Request) {
	if s.serviceMesh == nil {
		http.Error(w, "Service mesh not enabled", http.StatusServiceUnavailable)
		return
	}
	
	// Extract service ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}
	serviceID := parts[4]
	
	switch r.Method {
	case http.MethodGet:
		service, err := s.serviceMesh.GetService(serviceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		s.writeJSON(w, http.StatusOK, service)
		
	case http.MethodDelete:
		if err := s.serviceMesh.DeregisterService(serviceID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleFirewallRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listFirewallRules(w, r)
	case http.MethodPost:
		s.addFirewallRule(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listFirewallRules(w http.ResponseWriter, r *http.Request) {
	rules := s.firewall.ListRules()
	s.writeJSON(w, http.StatusOK, rules)
}

func (s *Server) addFirewallRule(w http.ResponseWriter, r *http.Request) {
	var rule firewall.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}
	
	if err := s.firewall.AddRule(&rule); err != nil {
		http.Error(w, fmt.Sprintf("Failed to add rule: %v", err), http.StatusInternalServerError)
		return
	}
	
	s.writeJSON(w, http.StatusCreated, rule)
}

func (s *Server) handleFirewallRuleByID(w http.ResponseWriter, r *http.Request) {
	// Extract rule ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid rule ID", http.StatusBadRequest)
		return
	}
	ruleID := parts[5]
	
	switch r.Method {
	case http.MethodGet:
		rule, err := s.firewall.GetRule(ruleID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		s.writeJSON(w, http.StatusOK, rule)
		
	case http.MethodDelete:
		if err := s.firewall.DeleteRule(ruleID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Metrics available at /metrics endpoint",
		"port":    s.config.Monitoring.MetricsPort,
		"path":    s.config.Monitoring.MetricsPath,
	}
	s.writeJSON(w, http.StatusOK, response)
}

// Helper methods

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.log.Errorf("Failed to encode JSON response: %v", err)
	}
}
