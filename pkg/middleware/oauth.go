// Package middleware provides OAuth 2.0 authentication middleware
package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guidewire-oss/fern-platform/pkg/config"
	"github.com/guidewire-oss/fern-platform/pkg/database"
	"github.com/guidewire-oss/fern-platform/pkg/logging"
	"gorm.io/gorm"
)

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// debugHTTPClient creates an HTTP client that logs all requests and responses
func (m *OAuthMiddleware) debugHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &debugTransport{
			logger: m.logger,
			base:   http.DefaultTransport,
		},
	}
}

// debugTransport is an HTTP transport that logs requests and responses
type debugTransport struct {
	logger *logging.Logger
	base   http.RoundTripper
}

func (t *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Log request
	reqDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		t.logger.WithError(err).Error("Failed to dump HTTP request")
	} else {
		// Sanitize authorization header for logging
		reqStr := string(reqDump)
		if strings.Contains(reqStr, "Authorization: Bearer") {
			reqStr = strings.ReplaceAll(reqStr, req.Header.Get("Authorization"), "Authorization: Bearer [REDACTED]")
		}
		t.logger.WithField("request", reqStr).Debug("HTTP Request")
	}

	// Make the request
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		t.logger.WithError(err).Error("HTTP request failed")
		return nil, err
	}

	// Log response
	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		t.logger.WithError(err).Error("Failed to dump HTTP response")
	} else {
		t.logger.WithField("response", string(respDump)).Debug("HTTP Response")
	}

	return resp, nil
}

// OAuthMiddleware provides OAuth 2.0 authentication middleware
type OAuthMiddleware struct {
	config *config.AuthConfig
	db     *gorm.DB
	logger *logging.Logger
}

// NewOAuthMiddleware creates a new OAuth authentication middleware
func NewOAuthMiddleware(cfg *config.AuthConfig, db *gorm.DB, logger *logging.Logger) *OAuthMiddleware {
	return &OAuthMiddleware{
		config: cfg,
		db:     db,
		logger: logger,
	}
}

// RequireOAuth middleware validates OAuth tokens and ensures user is authenticated
func (m *OAuthMiddleware) RequireOAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.Enabled || !m.config.OAuth.Enabled {
			c.Next()
			return
		}

		user, session, err := m.validateOAuthSession(c)
		if err != nil {
			m.logger.WithRequest(c.GetString("request_id"), c.Request.Method, c.Request.URL.Path).
				WithError(err).Warn("OAuth authentication failed")
			
			// Redirect to login for browser requests, return 401 for API requests
			if m.isAPIRequest(c) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			} else {
				c.Redirect(http.StatusFound, "/auth/login")
			}
			c.Abort()
			return
		}

		// Set user context
		m.setUserContext(c, user, session)
		c.Next()
	}
}

// RequireAdmin middleware ensures user has admin role
func (m *OAuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		m.RequireOAuth()(c)
		if c.IsAborted() {
			return
		}

		user, exists := m.getUserFromContext(c)
		if !exists || user.Role != string(database.RoleAdmin) {
			m.logger.WithRequest(c.GetString("request_id"), c.Request.Method, c.Request.URL.Path).
				WithField("user_id", user.UserID).
				WithField("user_role", user.Role).
				Warn("Admin access denied - insufficient privileges")
			
			if m.isAPIRequest(c) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
			} else {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied - admin privileges required"})
			}
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireProjectAccess middleware ensures user has access to specific project
func (m *OAuthMiddleware) RequireProjectAccess(minRole database.ProjectRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		m.RequireOAuth()(c)
		if c.IsAborted() {
			return
		}

		projectID := c.Param("projectId")
		if projectID == "" {
			projectID = c.Query("project_id")
		}

		if projectID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID required"})
			c.Abort()
			return
		}

		user, exists := m.getUserFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Admin users have access to all projects
		if user.Role == string(database.RoleAdmin) {
			c.Next()
			return
		}

		// Check project-specific access
		hasAccess, err := m.checkProjectAccess(user.UserID, projectID, minRole)
		if err != nil {
			m.logger.WithError(err).Error("Failed to check project access")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Access check failed"})
			c.Abort()
			return
		}

		if !hasAccess {
			m.logger.WithRequest(c.GetString("request_id"), c.Request.Method, c.Request.URL.Path).
				WithField("user_id", user.UserID).
				WithField("project_id", projectID).
				WithField("required_role", minRole).
				Warn("Project access denied")
			
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient project access"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// StartOAuthFlow initiates the OAuth authentication flow
func (m *OAuthMiddleware) StartOAuthFlow() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.OAuth.Enabled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth not enabled"})
			return
		}

		// Generate state parameter for security
		state, err := m.generateState()
		if err != nil {
			m.logger.WithError(err).Error("Failed to generate OAuth state")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start authentication"})
			return
		}

		// Store state in session/cookie for validation
		isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
		c.SetCookie("oauth_state", state, 600, "/", "", isSecure, true) // 10 minutes

		// Build authorization URL
		authURL := m.buildAuthURL(state)
		
		c.Redirect(http.StatusFound, authURL)
	}
}

// HandleOAuthCallback handles the OAuth callback
func (m *OAuthMiddleware) HandleOAuthCallback() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.OAuth.Enabled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth not enabled"})
			return
		}

		// Validate state parameter
		state := c.Query("state")
		expectedState, err := c.Cookie("oauth_state")
		if err != nil || state != expectedState {
			m.logger.WithField("state", state).Warn("OAuth state validation failed")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
			return
		}

		// Get authorization code
		code := c.Query("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code required"})
			return
		}

		// Exchange code for token
		token, err := m.exchangeCodeForToken(code)
		if err != nil {
			m.logger.WithError(err).Error("Failed to exchange code for token")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Token exchange failed"})
			return
		}

		// Get user info
		userInfo, err := m.getUserInfo(token.AccessToken)
		if err != nil {
			m.logger.WithError(err).Error("Failed to get user info")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get user information"})
			return
		}

		// Create or update user
		user, err := m.createOrUpdateUser(userInfo)
		if err != nil {
			m.logger.WithError(err).Error("Failed to create/update user")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User creation failed"})
			return
		}

		// Create session
		session, err := m.createSession(user, token, c.ClientIP(), c.GetHeader("User-Agent"))
		if err != nil {
			m.logger.WithError(err).Error("Failed to create session")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Session creation failed"})
			return
		}

		// Set session cookie
		m.setSessionCookie(c, session.SessionID)

		// Clear state cookie
		isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
		c.SetCookie("oauth_state", "", -1, "/", "", isSecure, true)

		// Update last login
		m.updateUserLastLogin(user.UserID)

		// Redirect to dashboard or intended page
		redirectURL := c.DefaultQuery("redirect", "/")
		c.Redirect(http.StatusFound, redirectURL)
	}
}

// Logout handles user logout
func (m *OAuthMiddleware) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		var idToken string
		
		sessionID, err := c.Cookie("session_id")
		if err == nil && sessionID != "" {
			// Get ID token from session before invalidating
			idToken = m.getIDTokenFromSession(sessionID)
			
			// Invalidate session in database
			m.invalidateSession(sessionID)
		}

		// Clear session cookie
		isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
		c.SetCookie("session_id", "", -1, "/", "", isSecure, true)

		// Build provider logout URL with ID token hint
		providerLogoutURL := m.buildProviderLogoutURL(idToken)
		
		// For AJAX requests, return JSON response with provider logout URL
		if c.GetHeader("Content-Type") == "application/json" || c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			c.JSON(http.StatusOK, gin.H{
				"message": "Logged out successfully",
				"logout_url": providerLogoutURL,
			})
			return
		}
		
		// For direct requests, redirect to provider logout
		c.Redirect(http.StatusFound, providerLogoutURL)
	}
}

// buildProviderLogoutURL constructs the OAuth provider logout URL with ID token hint
func (m *OAuthMiddleware) buildProviderLogoutURL(idToken string) string {
	if !m.config.OAuth.Enabled {
		return "/auth/login"
	}
	
	// If no ID token, just redirect to local login to avoid the error
	if idToken == "" {
		return "/auth/login"
	}
	
	// If a specific logout URL is configured, use it
	if m.config.OAuth.LogoutURL != "" {
		logoutURL := m.config.OAuth.LogoutURL
		
		// Add ID token hint parameter
		separator := "?"
		if strings.Contains(logoutURL, "?") {
			separator = "&"
		}
		logoutURL += fmt.Sprintf("%sid_token_hint=%s", separator, url.QueryEscape(idToken))
		
		// Add post-logout redirect URI
		redirectURL := m.config.OAuth.RedirectURL
		if redirectURL != "" {
			// Replace callback with login
			postLogoutURL := strings.Replace(redirectURL, "/auth/callback", "/auth/login", 1)
			logoutURL += fmt.Sprintf("&post_logout_redirect_uri=%s", url.QueryEscape(postLogoutURL))
		}
		
		return logoutURL
	}
	
	// Fallback: try to construct from issuer URL (common OIDC pattern)
	if m.config.OAuth.IssuerURL != "" {
		logoutURL := strings.TrimSuffix(m.config.OAuth.IssuerURL, "/") + "/protocol/openid-connect/logout"
		
		// Add ID token hint
		logoutURL += fmt.Sprintf("?id_token_hint=%s", url.QueryEscape(idToken))
		
		// Add post-logout redirect URI
		redirectURL := m.config.OAuth.RedirectURL
		if redirectURL != "" {
			// Replace callback with login
			postLogoutURL := strings.Replace(redirectURL, "/auth/callback", "/auth/login", 1)
			logoutURL += fmt.Sprintf("&post_logout_redirect_uri=%s", url.QueryEscape(postLogoutURL))
		}
		
		return logoutURL
	}
	
	// Final fallback to local login page if no provider logout available
	return "/auth/login"
}

// getIDTokenFromSession retrieves the ID token from the session
func (m *OAuthMiddleware) getIDTokenFromSession(sessionID string) string {
	var session database.UserSession
	err := m.db.Where("session_id = ? AND is_active = ?", sessionID, true).First(&session).Error
	if err != nil {
		return ""
	}
	return session.IDToken
}

// ValidateSession validates the current session and returns user info
func (m *OAuthMiddleware) ValidateSession(c *gin.Context) (*database.User, *database.UserSession, error) {
	return m.validateOAuthSession(c)
}

// Helper methods

func (m *OAuthMiddleware) validateOAuthSession(c *gin.Context) (*database.User, *database.UserSession, error) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		return nil, nil, fmt.Errorf("no session cookie")
	}

	var session database.UserSession
	err = m.db.Where("session_id = ? AND is_active = ? AND expires_at > ?", 
		sessionID, true, time.Now()).First(&session).Error
	if err != nil {
		return nil, nil, fmt.Errorf("invalid or expired session")
	}

	var user database.User
	err = m.db.Preload("UserGroups").Where("user_id = ?", session.UserID).First(&user).Error
	if err != nil {
		return nil, nil, fmt.Errorf("user not found")
	}

	if user.Status != "active" {
		return nil, nil, fmt.Errorf("user account inactive")
	}

	// Update last activity
	m.db.Model(&session).Update("last_activity", time.Now())

	return &user, &session, nil
}

func (m *OAuthMiddleware) isAPIRequest(c *gin.Context) bool {
	return strings.HasPrefix(c.Request.URL.Path, "/api/") ||
		   strings.Contains(c.GetHeader("Accept"), "application/json") ||
		   strings.Contains(c.GetHeader("Content-Type"), "application/json")
}

func (m *OAuthMiddleware) setUserContext(c *gin.Context, user *database.User, session *database.UserSession) {
	c.Set("user", user)
	c.Set("user_id", user.UserID)
	c.Set("user_role", user.Role)
	c.Set("session", session)
}

func (m *OAuthMiddleware) getUserFromContext(c *gin.Context) (*database.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	
	u, ok := user.(*database.User)
	return u, ok
}

func (m *OAuthMiddleware) checkProjectAccess(userID, projectID string, minRole database.ProjectRole) (bool, error) {
	var access database.ProjectAccess
	err := m.db.Where("user_id = ? AND project_id = ? AND (expires_at IS NULL OR expires_at > ?)", 
		userID, projectID, time.Now()).First(&access).Error
	
	if err != nil {
		return false, nil // No access found
	}

	// Check role hierarchy: admin > editor > viewer
	userRole := database.ProjectRole(access.Role)
	
	switch minRole {
	case database.ProjectRoleViewer:
		return userRole == database.ProjectRoleViewer || 
			   userRole == database.ProjectRoleEditor || 
			   userRole == database.ProjectRoleAdmin, nil
	case database.ProjectRoleEditor:
		return userRole == database.ProjectRoleEditor || 
			   userRole == database.ProjectRoleAdmin, nil
	case database.ProjectRoleAdmin:
		return userRole == database.ProjectRoleAdmin, nil
	default:
		return false, nil
	}
}

func (m *OAuthMiddleware) generateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (m *OAuthMiddleware) buildAuthURL(state string) string {
	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", m.config.OAuth.ClientID)
	params.Add("redirect_uri", m.config.OAuth.RedirectURL)
	params.Add("scope", strings.Join(m.config.OAuth.Scopes, " "))
	params.Add("state", state)
	
	return m.config.OAuth.AuthURL + "?" + params.Encode()
}

// OAuth token response structure
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`      // ID token for logout
	Scope        string `json:"scope,omitempty"`
}

func (m *OAuthMiddleware) exchangeCodeForToken(code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", m.config.OAuth.RedirectURL)
	data.Set("client_id", m.config.OAuth.ClientID)
	data.Set("client_secret", m.config.OAuth.ClientSecret)

	req, err := http.NewRequest("POST", m.config.OAuth.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := m.debugHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	m.logger.WithField("token_type", token.TokenType).
		WithField("expires_in", token.ExpiresIn).
		WithField("access_token_length", len(token.AccessToken)).
		WithField("access_token_prefix", token.AccessToken[:min(50, len(token.AccessToken))]).
		Debug("Successfully exchanged code for token")

	return &token, nil
}

// UserInfo represents user information from OAuth provider
type UserInfo struct {
	Sub           string                 `json:"sub"`
	Email         string                 `json:"email"`
	Name          string                 `json:"name"`
	Picture       string                 `json:"picture"`
	Groups        []string               `json:"groups"`
	Roles         []string               `json:"roles"`
	Attributes    map[string]interface{} `json:"-"` // Store all other attributes
}

func (m *OAuthMiddleware) getUserInfo(accessToken string) (*UserInfo, error) {
	m.logger.WithField("userinfo_url", m.config.OAuth.UserInfoURL).
		WithField("token_length", len(accessToken)).
		WithField("token_prefix", accessToken[:min(50, len(accessToken))]).
		Debug("Making userinfo request")

	req, err := http.NewRequest("GET", m.config.OAuth.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := m.debugHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		m.logger.WithError(err).Error("Userinfo request failed")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		m.logger.WithField("status", resp.StatusCode).
			WithField("response_body", string(body)).
			Error("Userinfo request returned non-200 status")
		return nil, fmt.Errorf("userinfo request failed with status: %d", resp.StatusCode)
	}

	var rawInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawInfo); err != nil {
		return nil, err
	}

	// Extract user info using configurable field mappings
	userInfo := &UserInfo{
		Attributes: rawInfo,
	}

	if sub, ok := rawInfo[m.config.OAuth.UserIDField].(string); ok {
		userInfo.Sub = sub
	}
	if email, ok := rawInfo[m.config.OAuth.EmailField].(string); ok {
		userInfo.Email = email
	}
	if name, ok := rawInfo[m.config.OAuth.NameField].(string); ok {
		userInfo.Name = name
	}
	if picture, ok := rawInfo["picture"].(string); ok {
		userInfo.Picture = picture
	}

	// Extract groups and roles
	if groups, ok := rawInfo[m.config.OAuth.GroupsField].([]interface{}); ok {
		for _, group := range groups {
			if groupStr, ok := group.(string); ok {
				userInfo.Groups = append(userInfo.Groups, groupStr)
			}
		}
	}

	if roles, ok := rawInfo[m.config.OAuth.RolesField].([]interface{}); ok {
		for _, role := range roles {
			if roleStr, ok := role.(string); ok {
				userInfo.Roles = append(userInfo.Roles, roleStr)
			}
		}
	}

	return userInfo, nil
}

func (m *OAuthMiddleware) createOrUpdateUser(userInfo *UserInfo) (*database.User, error) {
	var user database.User
	
	// Try to find existing user
	err := m.db.Where("user_id = ? OR email = ?", userInfo.Sub, userInfo.Email).First(&user).Error
	
	if err == gorm.ErrRecordNotFound {
		// Create new user
		user = database.User{
			UserID:     userInfo.Sub,
			Email:      userInfo.Email,
			Name:       userInfo.Name,
			Role:       m.determineUserRole(userInfo),
			Status:     "active",
			ProfileURL: userInfo.Picture,
		}
		
		// Safely extract additional fields from attributes
		if firstName, ok := userInfo.Attributes["given_name"].(string); ok {
			user.FirstName = firstName
		}
		if lastName, ok := userInfo.Attributes["family_name"].(string); ok {
			user.LastName = lastName
		}
		if emailVerified, ok := userInfo.Attributes["email_verified"].(bool); ok {
			user.EmailVerified = emailVerified
		}
		
		if err := m.db.Create(&user).Error; err != nil {
			return nil, err
		}
		
		// Create user group memberships
		for _, group := range userInfo.Groups {
			userGroup := database.UserGroup{
				UserID:    user.UserID,
				GroupName: group,
			}
			if err := m.db.Create(&userGroup).Error; err != nil {
				m.logger.WithError(err).WithField("group", group).Warn("Failed to create user group membership")
			}
		}
	} else if err != nil {
		return nil, err
	} else {
		// Update existing user
		user.Email = userInfo.Email
		user.Name = userInfo.Name
		user.ProfileURL = userInfo.Picture
		
		// Safely extract additional fields from attributes
		if firstName, ok := userInfo.Attributes["given_name"].(string); ok {
			user.FirstName = firstName
		}
		if lastName, ok := userInfo.Attributes["family_name"].(string); ok {
			user.LastName = lastName
		}
		if emailVerified, ok := userInfo.Attributes["email_verified"].(bool); ok {
			user.EmailVerified = emailVerified
		}
		
		// Update role if needed
		newRole := m.determineUserRole(userInfo)
		if newRole != user.Role {
			user.Role = newRole
		}
		
		if err := m.db.Save(&user).Error; err != nil {
			return nil, err
		}
		
		// Update user group memberships - remove old ones and add new ones
		if err := m.db.Where("user_id = ?", user.UserID).Delete(&database.UserGroup{}).Error; err != nil {
			m.logger.WithError(err).Warn("Failed to delete old user groups")
		}
		
		for _, group := range userInfo.Groups {
			userGroup := database.UserGroup{
				UserID:    user.UserID,
				GroupName: group,
			}
			if err := m.db.Create(&userGroup).Error; err != nil {
				m.logger.WithError(err).WithField("group", group).Warn("Failed to create user group membership")
			}
		}
	}

	return &user, nil
}

func (m *OAuthMiddleware) determineUserRole(userInfo *UserInfo) string {
	// Check if user is in admin users list
	for _, adminUser := range m.config.OAuth.AdminUsers {
		if adminUser == userInfo.Email || adminUser == userInfo.Sub {
			return string(database.RoleAdmin)
		}
	}

	// Check if user is in admin groups
	for _, group := range userInfo.Groups {
		for _, adminGroup := range m.config.OAuth.AdminGroups {
			if group == adminGroup {
				return string(database.RoleAdmin)
			}
		}
	}

	// Check user role mapping
	if role, exists := m.config.OAuth.UserRoleMapping[userInfo.Email]; exists {
		return role
	}
	if role, exists := m.config.OAuth.UserRoleMapping[userInfo.Sub]; exists {
		return role
	}

	// Check group role mapping
	for _, group := range userInfo.Groups {
		if role, exists := m.config.OAuth.GroupRoleMapping[group]; exists {
			return role
		}
	}

	// Default to regular user
	return string(database.RoleUser)
}

func (m *OAuthMiddleware) createSession(user *database.User, token *TokenResponse, ipAddress, userAgent string) (*database.UserSession, error) {
	sessionID, err := m.generateState() // Reuse state generation for session ID
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	if token.ExpiresIn == 0 {
		expiresAt = time.Now().Add(24 * time.Hour) // Default 24 hours
	}

	session := database.UserSession{
		UserID:       user.UserID,
		SessionID:    sessionID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		IDToken:      token.IDToken,
		ExpiresAt:    expiresAt,
		IsActive:     true,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		LastActivity: time.Now(),
	}

	if err := m.db.Create(&session).Error; err != nil {
		return nil, err
	}

	return &session, nil
}

func (m *OAuthMiddleware) setSessionCookie(c *gin.Context, sessionID string) {
	// Set secure cookie for 24 hours
	// Use Secure flag in production (HTTPS)
	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetCookie("session_id", sessionID, 86400, "/", "", isSecure, true)
}

func (m *OAuthMiddleware) updateUserLastLogin(userID string) {
	m.db.Model(&database.User{}).Where("user_id = ?", userID).Update("last_login_at", time.Now())
}

func (m *OAuthMiddleware) invalidateSession(sessionID string) {
	m.db.Model(&database.UserSession{}).Where("session_id = ?", sessionID).Update("is_active", false)
}

// Helper functions for Gin context

// GetOAuthUser extracts the OAuth user from Gin context
func GetOAuthUser(c *gin.Context) (*database.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	
	u, ok := user.(*database.User)
	return u, ok
}

// GetOAuthSession extracts the OAuth session from Gin context
func GetOAuthSession(c *gin.Context) (*database.UserSession, bool) {
	session, exists := c.Get("session")
	if !exists {
		return nil, false
	}
	
	s, ok := session.(*database.UserSession)
	return s, ok
}

// IsAdmin checks if the current user is an admin
func IsAdmin(c *gin.Context) bool {
	user, exists := GetOAuthUser(c)
	return exists && user.Role == string(database.RoleAdmin)
}