package terrareg_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/middleware"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/middleware/model"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestGPGKeyCreate_Authentication tests GPG key creation with RequireAdmin middleware (POST /v2/gpg-keys)
func TestGPGKeyCreate_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "gpg-namespace")

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) (*http.Request, *model.AuthContext)
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				return testutils.BuildUnauthenticatedRequest(t, "POST", "/v2/gpg-keys"), nil
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "regular authenticated user returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				return testutils.BuildAuthenticatedRequest(t, db, "POST", "/v2/gpg-keys", "regular-user", false)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "admin user can create GPG keys",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				return testutils.BuildAdminRequest(t, db, "POST", "/v2/gpg-keys")
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, authCtx := tt.setupAuth(t, db)
			if authCtx != nil {
				ctx := middleware.SetAuthContextInContext(req.Context(), authCtx)
				req = req.WithContext(ctx)
			}

			w := testutils.ServeHTTP(router, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestGPGKeyDelete_Authentication tests GPG key deletion with RequireAdmin middleware (DELETE /v2/gpg-keys/{namespace}/{key_id})
func TestGPGKeyDelete_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "gpg-delete-namespace")

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) (*http.Request, *model.AuthContext)
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				req := testutils.BuildUnauthenticatedRequest(t, "DELETE", "/v2/gpg-keys/gpg-delete-namespace/test-key-id")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "gpg-delete-namespace", "key_id": "test-key-id"}), nil
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "regular authenticated user returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				req, authCtx := testutils.BuildAuthenticatedRequest(t, db, "DELETE", "/v2/gpg-keys/gpg-delete-namespace/test-key-id", "regular-user", false)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "gpg-delete-namespace", "key_id": "test-key-id"}), authCtx
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "admin user can delete GPG keys",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				req, authCtx := testutils.BuildAdminRequest(t, db, "DELETE", "/v2/gpg-keys/gpg-delete-namespace/test-key-id")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "gpg-delete-namespace", "key_id": "test-key-id"}), authCtx
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, authCtx := tt.setupAuth(t, db)
			if authCtx != nil {
				ctx := middleware.SetAuthContextInContext(req.Context(), authCtx)
				req = req.WithContext(ctx)
			}

			w := testutils.ServeHTTP(router, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestGPGKeyList_AllAuthMethods tests GET GPG key list endpoint with OptionalAuth
// All authentication states should return 200
func TestGPGKeyList_AllAuthMethods(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) (*http.Request, *model.AuthContext)
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				return testutils.BuildUnauthenticatedRequest(t, "GET", "/v2/gpg-keys"), nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated regular user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				return testutils.BuildAuthenticatedRequest(t, db, "GET", "/v2/gpg-keys", "regular-user", false)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated admin user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				return testutils.BuildAdminRequest(t, db, "GET", "/v2/gpg-keys")
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, authCtx := tt.setupAuth(t, db)
			if authCtx != nil {
				ctx := middleware.SetAuthContextInContext(req.Context(), authCtx)
				req = req.WithContext(ctx)
			}

			w := testutils.ServeHTTP(router, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
