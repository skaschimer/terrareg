package auth

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
)

// AdminSessionAuthMethod implements immutable authentication for admin users via session cookies
type AdminSessionAuthMethod struct {
	sessionRepo   repository.SessionRepository
	userGroupRepo repository.UserGroupRepository
}

// NewAdminSessionAuthMethod creates a new immutable admin session authentication method
func NewAdminSessionAuthMethod(
	sessionRepo repository.SessionRepository,
	userGroupRepo repository.UserGroupRepository,
) *AdminSessionAuthMethod {
	return &AdminSessionAuthMethod{
		sessionRepo:   sessionRepo,
		userGroupRepo: userGroupRepo,
	}
}

// GetProviderType returns the authentication provider type
func (a *AdminSessionAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodAdminSession
}

// IsEnabled returns whether this authentication method is enabled
func (a *AdminSessionAuthMethod) IsEnabled() bool {
	return true
}

// Authenticate authenticates a request and returns an AdminSessionAuthContext
// This implements the SessionAuthMethod interface, which receives sessionData from the auth factory
func (a *AdminSessionAuthMethod) Authenticate(ctx context.Context, sessionData map[string]interface{}) (auth.AuthContext, error) {
	// Get session ID from sessionData map (populated by auth factory from cookies/headers)
	sessionIDInterface, exists := sessionData["session_id"]
	if !exists {
		return nil, nil // No session ID, let other auth methods try
	}

	sessionID, ok := sessionIDInterface.(string)
	if !ok || sessionID == "" || strings.TrimSpace(sessionID) == "" {
		return nil, nil // Invalid session ID, let other auth methods try
	}

	// Find session in database
	session, err := a.sessionRepo.FindByID(ctx, sessionID)
	if err != nil || session == nil {
		return nil, nil // Let other auth methods try
	}

	// Check if session is expired
	if session.IsExpired() {
		return nil, nil // Let other auth methods try
	}

	// Parse provider source auth to get user information
	userInfo, err := a.parseProviderSourceAuth(session.ProviderSourceAuth)
	if err != nil {
		return nil, nil // Let other auth methods try
	}

	// Create AdminSessionAuthContext with session state (convert string UserID to int for compatibility)
	// For now, use 0 as placeholder since we don't have proper user ID conversion
	userID := 0 // TODO: Convert userInfo.UserID from string to int when user ID system is defined
	authContext := auth.NewAdminSessionAuthContext(ctx, userID, userInfo.Username, userInfo.Email, sessionID)

	// Admin sessions are always admin (matching Python's BaseAdminAuthMethod)
	// In Python, AdminSessionAuthMethod inherits from both BaseAdminAuthMethod AND BaseSessionAuthMethod
	// BaseAdminAuthMethod always returns is_admin=True
	authContext.SetAdmin(true)

	// Get user permissions
	permissions, err := a.getUserPermissions(ctx, userInfo.UserID)
	if err == nil && permissions != nil {
		for namespace, permission := range permissions {
			authContext.SetPermission(namespace, permission)
		}
	}

	return authContext, nil
}

// parseProviderSourceAuth parses the provider_source_auth JSON to extract user information
func (a *AdminSessionAuthMethod) parseProviderSourceAuth(providerSourceAuth []byte) (*UserInfo, error) {
	if len(providerSourceAuth) == 0 {
		// Default admin user info if no auth data
		return &UserInfo{
			UserID:   "admin",
			Username: "Admin User",
		}, nil
	}

	var userInfo UserInfo
	err := json.Unmarshal(providerSourceAuth, &userInfo)
	if err != nil {
		// If parsing fails, return default admin info
		return &UserInfo{
			UserID:   "admin",
			Username: "Admin User",
		}, nil
	}

	return &userInfo, nil
}

// getUserPermissions gets the user's permissions across all namespaces
// @TODO Can these be removed? The auth Method shouldn't need these, since everything uses AuthContext
func (a *AdminSessionAuthMethod) getUserPermissions(ctx context.Context, userID string) (map[string]string, error) {
	permissions := make(map[string]string)

	// In a real implementation, query the database for user permissions
	// For now, return empty permissions - actual permissions would be set by user groups

	return permissions, nil
}

// AuthMethod interface implementation for the base AdminSessionAuthMethod
// These return default values since the actual auth state is in the AdminSessionAuthContext
// @TODO Can these be removed? The auth Method shouldn't need these, since everything uses AuthContext
func (a *AdminSessionAuthMethod) IsBuiltInAdmin() bool                     { return false }
func (a *AdminSessionAuthMethod) IsAdmin() bool                            { return false }
func (a *AdminSessionAuthMethod) IsAuthenticated() bool                    { return false }
func (a *AdminSessionAuthMethod) RequiresCSRF() bool                       { return true }
func (a *AdminSessionAuthMethod) CheckAuthState() bool                     { return false }
func (a *AdminSessionAuthMethod) CanPublishModuleVersion(string) bool      { return false }
func (a *AdminSessionAuthMethod) CanUploadModuleVersion(string) bool       { return false }
func (a *AdminSessionAuthMethod) CheckNamespaceAccess(string, string) bool { return false }
func (a *AdminSessionAuthMethod) GetAllNamespacePermissions() map[string]string {
	return make(map[string]string)
}
func (a *AdminSessionAuthMethod) GetUsername() string           { return "" }
func (a *AdminSessionAuthMethod) GetUserGroupNames() []string   { return []string{} }
func (a *AdminSessionAuthMethod) CanAccessReadAPI() bool        { return false }
func (a *AdminSessionAuthMethod) CanAccessTerraformAPI() bool   { return false }
func (a *AdminSessionAuthMethod) GetTerraformAuthToken() string { return "" }
func (a *AdminSessionAuthMethod) GetProviderData() map[string]interface{} {
	return make(map[string]interface{})
}

// UserInfo represents user information extracted from session
type UserInfo struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
