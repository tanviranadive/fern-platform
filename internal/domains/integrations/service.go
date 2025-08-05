package integrations

import (
	"context"
	"fmt"
)

// JiraConnectionService handles JIRA connection operations
type JiraConnectionService struct {
	repo           JiraConnectionRepository
	jiraClient     JiraClient
	encryptionKey  []byte
}

// NewJiraConnectionService creates a new JIRA connection service
func NewJiraConnectionService(repo JiraConnectionRepository, jiraClient JiraClient, encryptionKey []byte) *JiraConnectionService {
	return &JiraConnectionService{
		repo:          repo,
		jiraClient:    jiraClient,
		encryptionKey: encryptionKey,
	}
}

// CreateConnection creates a new JIRA connection
func (s *JiraConnectionService) CreateConnection(ctx context.Context, projectID, name, jiraURL string, authType AuthenticationType, projectKey, username, credential string) (*JiraConnection, error) {
	// Create the connection
	conn, err := NewJiraConnection(projectID, name, jiraURL, authType, projectKey, username, credential)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	// Encrypt the credential before saving
	encrypted, err := conn.GetEncryptedCredential(s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt credential: %w", err)
	}
	conn.encryptedCredential = encrypted

	// Save to repository
	if err := s.repo.Create(ctx, conn); err != nil {
		return nil, fmt.Errorf("failed to save connection: %w", err)
	}

	return conn, nil
}

// UpdateConnection updates an existing JIRA connection
func (s *JiraConnectionService) UpdateConnection(ctx context.Context, connectionID, name, jiraURL, projectKey string) (*JiraConnection, error) {
	// Retrieve the connection
	conn, err := s.repo.FindByID(ctx, connectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find connection: %w", err)
	}

	// Update connection info
	if err := conn.UpdateConnectionInfo(name, jiraURL, projectKey); err != nil {
		return nil, fmt.Errorf("failed to update connection info: %w", err)
	}

	// Save changes
	if err := s.repo.Update(ctx, conn); err != nil {
		return nil, fmt.Errorf("failed to save connection: %w", err)
	}

	return conn, nil
}

// UpdateCredentials updates the credentials for a JIRA connection
func (s *JiraConnectionService) UpdateCredentials(ctx context.Context, connectionID string, authType AuthenticationType, username, credential string) (*JiraConnection, error) {
	// Retrieve the connection
	conn, err := s.repo.FindByID(ctx, connectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find connection: %w", err)
	}

	// Update credentials
	if err := conn.UpdateCredentials(authType, username, credential); err != nil {
		return nil, fmt.Errorf("failed to update credentials: %w", err)
	}

	// Encrypt the new credential
	encrypted, err := conn.GetEncryptedCredential(s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt credential: %w", err)
	}
	conn.encryptedCredential = encrypted

	// Save changes
	if err := s.repo.Update(ctx, conn); err != nil {
		return nil, fmt.Errorf("failed to save connection: %w", err)
	}

	return conn, nil
}

// TestConnection tests a JIRA connection
func (s *JiraConnectionService) TestConnection(ctx context.Context, connectionID string) error {
	// Retrieve the connection
	conn, err := s.repo.FindByID(ctx, connectionID)
	if err != nil {
		return fmt.Errorf("failed to find connection: %w", err)
	}

	// Decrypt the credential
	decrypted, err := DecryptCredential(conn.encryptedCredential, s.encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Create a temporary connection with decrypted credential for testing
	testConn := *conn
	testConn.encryptedCredential = decrypted

	// Test the connection
	if err := testConn.TestConnection(ctx, s.jiraClient); err != nil {
		// Save the failed status
		s.repo.Update(ctx, &testConn)
		return err
	}

	// Save the successful status
	return s.repo.Update(ctx, &testConn)
}

// ActivateConnection activates a JIRA connection
func (s *JiraConnectionService) ActivateConnection(ctx context.Context, connectionID string) error {
	conn, err := s.repo.FindByID(ctx, connectionID)
	if err != nil {
		return fmt.Errorf("failed to find connection: %w", err)
	}

	conn.Activate()
	return s.repo.Update(ctx, conn)
}

// DeactivateConnection deactivates a JIRA connection
func (s *JiraConnectionService) DeactivateConnection(ctx context.Context, connectionID string) error {
	conn, err := s.repo.FindByID(ctx, connectionID)
	if err != nil {
		return fmt.Errorf("failed to find connection: %w", err)
	}

	conn.Deactivate()
	return s.repo.Update(ctx, conn)
}

// DeleteConnection deletes a JIRA connection
func (s *JiraConnectionService) DeleteConnection(ctx context.Context, connectionID string) error {
	return s.repo.Delete(ctx, connectionID)
}

// GetConnection retrieves a JIRA connection by ID
func (s *JiraConnectionService) GetConnection(ctx context.Context, connectionID string) (*JiraConnection, error) {
	return s.repo.FindByID(ctx, connectionID)
}

// GetProjectConnections retrieves all connections for a project
func (s *JiraConnectionService) GetProjectConnections(ctx context.Context, projectID string) ([]*JiraConnection, error) {
	return s.repo.FindByProjectID(ctx, projectID)
}

// GetActiveProjectConnections retrieves all active connections for a project
func (s *JiraConnectionService) GetActiveProjectConnections(ctx context.Context, projectID string) ([]*JiraConnection, error) {
	return s.repo.FindActiveByProjectID(ctx, projectID)
}