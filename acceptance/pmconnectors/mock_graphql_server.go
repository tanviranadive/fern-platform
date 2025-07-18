package pmconnectors_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

// MockGraphQLServer simulates the Fern platform GraphQL API for PM connectors
type MockGraphQLServer struct {
	server     *httptest.Server
	connectors []PMConnector
	mappings   map[string][]FieldMapping
}

type PMConnector struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Type             string     `json:"type"`
	BaseURL          string     `json:"baseURL"`
	Status           string     `json:"status"`
	HealthStatus     string     `json:"healthStatus"`
	LastHealthCheck  *time.Time `json:"lastHealthCheck"`
	SyncInterval     string     `json:"syncInterval"`
	LastSyncAt       *time.Time `json:"lastSyncAt"`
	NextSyncAt       *time.Time `json:"nextSyncAt"`
	RequirementCount int        `json:"requirementCount"`
	HasCredentials   bool       `json:"hasCredentials"`
	CanManage        bool       `json:"canManage"`
}

type FieldMapping struct {
	ID              string      `json:"id"`
	SourcePath      string      `json:"sourcePath"`
	TargetField     string      `json:"targetField"`
	TransformType   string      `json:"transformType"`
	TransformConfig interface{} `json:"transformConfig"`
	IsActive        bool        `json:"isActive"`
	Order           int         `json:"order"`
}

// NewMockGraphQLServer creates a new mock GraphQL server
func NewMockGraphQLServer() *MockGraphQLServer {
	m := &MockGraphQLServer{
		connectors: make([]PMConnector, 0),
		mappings:   make(map[string][]FieldMapping),
	}

	m.server = httptest.NewServer(http.HandlerFunc(m.handler))
	return m
}

// URL returns the mock server URL
func (m *MockGraphQLServer) URL() string {
	return m.server.URL
}

// Close shuts down the mock server
func (m *MockGraphQLServer) Close() {
	m.server.Close()
}

// handler processes GraphQL requests
func (m *MockGraphQLServer) handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req GraphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"message": "Invalid request body"},
			},
		})
		return
	}

	// Route based on operation
	if strings.Contains(req.Query, "GetPMConnectors") {
		m.handleGetConnectors(w, req)
	} else if strings.Contains(req.Query, "CreatePMConnector") {
		m.handleCreateConnector(w, req)
	} else if strings.Contains(req.Query, "SetPMConnectorCredentials") {
		m.handleSetCredentials(w, req)
	} else if strings.Contains(req.Query, "TestPMConnection") {
		m.handleTestConnection(w, req)
	} else if strings.Contains(req.Query, "SyncPMConnector") {
		m.handleSyncConnector(w, req)
	} else if strings.Contains(req.Query, "GetConnectorMappings") {
		m.handleGetMappings(w, req)
	} else if strings.Contains(req.Query, "UpdateFieldMappings") {
		m.handleUpdateMappings(w, req)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]string{
				{"message": "Unknown operation"},
			},
		})
	}
}

// handleGetConnectors returns mock PM connectors
func (m *MockGraphQLServer) handleGetConnectors(w http.ResponseWriter, req GraphQLRequest) {
	edges := make([]map[string]interface{}, len(m.connectors))
	for i, conn := range m.connectors {
		edges[i] = map[string]interface{}{
			"node":   conn,
			"cursor": fmt.Sprintf("cursor_%d", i),
		}
	}

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"pmConnectors": map[string]interface{}{
				"edges": edges,
				"pageInfo": map[string]interface{}{
					"hasNextPage": false,
					"endCursor":   fmt.Sprintf("cursor_%d", len(m.connectors)-1),
				},
				"totalCount": len(m.connectors),
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleCreateConnector creates a new mock connector
func (m *MockGraphQLServer) handleCreateConnector(w http.ResponseWriter, req GraphQLRequest) {
	input := req.Variables["input"].(map[string]interface{})

	connector := PMConnector{
		ID:               fmt.Sprintf("conn_%d", len(m.connectors)+1),
		Name:             input["name"].(string),
		Type:             input["type"].(string),
		BaseURL:          input["baseURL"].(string),
		Status:           "INACTIVE",
		HealthStatus:     "UNKNOWN",
		SyncInterval:     "DAILY",
		RequirementCount: 0,
		HasCredentials:   false,
		CanManage:        true,
	}

	m.connectors = append(m.connectors, connector)

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"createPMConnector": connector,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSetCredentials sets credentials for a connector
func (m *MockGraphQLServer) handleSetCredentials(w http.ResponseWriter, req GraphQLRequest) {
	connectorID := req.Variables["connectorID"].(string)

	// Find and update connector
	for i, conn := range m.connectors {
		if conn.ID == connectorID {
			m.connectors[i].HasCredentials = true

			response := map[string]interface{}{
				"data": map[string]interface{}{
					"setPMConnectorCredentials": m.connectors[i],
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Connector not found
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"errors": []map[string]string{
			{"message": "Connector not found"},
		},
	})
}

// handleTestConnection tests a connector connection
func (m *MockGraphQLServer) handleTestConnection(w http.ResponseWriter, req GraphQLRequest) {
	connectorID := req.Variables["id"].(string)

	// Find connector
	for i, conn := range m.connectors {
		if conn.ID == connectorID {
			// Update health status
			now := time.Now()
			m.connectors[i].HealthStatus = "HEALTHY"
			m.connectors[i].LastHealthCheck = &now

			response := map[string]interface{}{
				"data": map[string]interface{}{
					"testPMConnection": "HEALTHY",
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Connector not found
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"errors": []map[string]string{
			{"message": "Connector not found"},
		},
	})
}

// handleSyncConnector syncs a connector
func (m *MockGraphQLServer) handleSyncConnector(w http.ResponseWriter, req GraphQLRequest) {
	connectorID := req.Variables["id"].(string)

	// Find connector
	for i, conn := range m.connectors {
		if conn.ID == connectorID {
			// Update sync status
			now := time.Now()
			m.connectors[i].LastSyncAt = &now
			m.connectors[i].RequirementCount = 42 // Mock requirement count
			m.connectors[i].Status = "ACTIVE"

			response := map[string]interface{}{
				"data": map[string]interface{}{
					"syncPMConnector": map[string]interface{}{
						"id":             connectorID,
						"status":         "COMPLETED",
						"itemsProcessed": 42,
						"itemsFailed":    0,
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Connector not found
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"errors": []map[string]string{
			{"message": "Connector not found"},
		},
	})
}

// handleGetMappings returns field mappings for a connector
func (m *MockGraphQLServer) handleGetMappings(w http.ResponseWriter, req GraphQLRequest) {
	connectorID := req.Variables["id"].(string)

	mappings, exists := m.mappings[connectorID]
	if !exists {
		mappings = []FieldMapping{}
	}

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"pmConnector": map[string]interface{}{
				"id":            connectorID,
				"fieldMappings": mappings,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleUpdateMappings updates field mappings for a connector
func (m *MockGraphQLServer) handleUpdateMappings(w http.ResponseWriter, req GraphQLRequest) {
	connectorID := req.Variables["connectorID"].(string)
	mappingsInput := req.Variables["mappings"].([]interface{})

	// Convert input to field mappings
	mappings := make([]FieldMapping, len(mappingsInput))
	for i, input := range mappingsInput {
		mapping := input.(map[string]interface{})
		mappings[i] = FieldMapping{
			ID:            fmt.Sprintf("mapping_%d", i+1),
			SourcePath:    mapping["sourcePath"].(string),
			TargetField:   mapping["targetField"].(string),
			TransformType: getStringOrDefault(mapping, "transformType", "DIRECT"),
			IsActive:      getBoolOrDefault(mapping, "isActive", true),
			Order:         getIntOrDefault(mapping, "order", i),
		}
	}

	m.mappings[connectorID] = mappings

	// Find and update connector to activate it if it has mappings and credentials
	for i, conn := range m.connectors {
		if conn.ID == connectorID && conn.HasCredentials && len(mappings) > 0 {
			m.connectors[i].Status = "ACTIVE"
			break
		}
	}

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"updateFieldMappings": map[string]interface{}{
				"id":            connectorID,
				"fieldMappings": mappings,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper functions
func getStringOrDefault(m map[string]interface{}, key, defaultValue string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return defaultValue
}

func getBoolOrDefault(m map[string]interface{}, key string, defaultValue bool) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return defaultValue
}

func getIntOrDefault(m map[string]interface{}, key string, defaultValue int) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return defaultValue
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// AddConnector adds a pre-existing connector for testing
func (m *MockGraphQLServer) AddConnector(connector PMConnector) {
	m.connectors = append(m.connectors, connector)
}

// GetConnectors returns all connectors
func (m *MockGraphQLServer) GetConnectors() []PMConnector {
	return m.connectors
}

// Reset clears all data
func (m *MockGraphQLServer) Reset() {
	m.connectors = make([]PMConnector, 0)
	m.mappings = make(map[string][]FieldMapping)
}
