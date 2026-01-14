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

// TestNamespaceCreate_Authentication tests namespace creation with RequireAuth middleware
func TestNamespaceCreate_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 401",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				return testutils.BuildUnauthenticatedRequest(t, "POST", "/v1/terrareg/namespaces")
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "authenticated regular user can create namespace",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithSession(t, db, "POST", "/v1/terrareg/namespaces", "regular-user", false)
				return req
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated admin user can create namespace",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "POST", "/v1/terrareg/namespaces")
				return req
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupAuth(t, db)
			w := testutils.ServeHTTP(router, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestNamespaceUpdate_Authentication tests namespace update with RequireNamespacePermission(FULL) middleware
func TestNamespaceUpdate_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "test-namespace")

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) *http.Request
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 401",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req := testutils.BuildUnauthenticatedRequest(t, "POST", "/v1/terrareg/namespaces/test-namespace")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace"})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "user with READ permission returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "POST", "/v1/terrareg/namespaces/test-namespace",
					"readonly-user", "test-namespace", sqldb.PermissionTypeRead,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "user with MODIFY permission returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "POST", "/v1/terrareg/namespaces/test-namespace",
					"modify-user", "test-namespace", sqldb.PermissionTypeModify,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace"})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "user with FULL permission can update namespace",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "POST", "/v1/terrareg/namespaces/test-namespace",
					"full-user", "test-namespace", sqldb.PermissionTypeFull,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace"})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "admin user can update any namespace",
			setupAuth: func(t *testing.T, db *sqldb.Database) *http.Request {
				req, _ := testutils.BuildAdminRequest(t, db, "POST", "/v1/terrareg/namespaces/test-namespace")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace"})
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupAuth(t, db)
			w := testutils.ServeHTTP(router, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestNamespaceDelete_Authentication tests namespace deletion with RequireNamespacePermission(FULL) middleware
func TestNamespaceDelete_Authentication(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "delete-test-namespace")

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) (*http.Request, *model.AuthContext)
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 401",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				req := testutils.BuildUnauthenticatedRequest(t, "DELETE", "/v1/terrareg/namespaces/delete-test-namespace")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "delete-test-namespace"}), nil
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "user with READ permission returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				req, authCtx := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "DELETE", "/v1/terrareg/namespaces/delete-test-namespace",
					"readonly-user", "delete-test-namespace", sqldb.PermissionTypeRead,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "delete-test-namespace"}), authCtx
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "user with MODIFY permission returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				req, authCtx := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "DELETE", "/v1/terrareg/namespaces/delete-test-namespace",
					"modify-user", "delete-test-namespace", sqldb.PermissionTypeModify,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "delete-test-namespace"}), authCtx
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "user with FULL permission can delete namespace",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				req, authCtx := testutils.BuildAuthenticatedRequestWithNamespacePermission(
					t, db, "DELETE", "/v1/terrareg/namespaces/delete-test-namespace",
					"full-user", "delete-test-namespace", sqldb.PermissionTypeFull,
				)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "delete-test-namespace"}), authCtx
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "admin user can delete any namespace",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				req, authCtx := testutils.BuildAdminRequest(t, db, "DELETE", "/v1/terrareg/namespaces/delete-test-namespace")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "delete-test-namespace"}), authCtx
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

// TestNamespaceGet_AllAuthMethods tests GET namespace endpoint with OptionalAuth
// All authentication states should return 200
func TestNamespaceGet_AllAuthMethods(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "get-test-namespace")

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) (*http.Request, *model.AuthContext)
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				req := testutils.BuildUnauthenticatedRequest(t, "GET", "/v1/terrareg/namespaces/get-test-namespace")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "get-test-namespace"}), nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated regular user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				req, authCtx := testutils.BuildAuthenticatedRequest(t, db, "GET", "/v1/terrareg/namespaces/get-test-namespace", "regular-user", false)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "get-test-namespace"}), authCtx
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated admin user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				req, authCtx := testutils.BuildAdminRequest(t, db, "GET", "/v1/terrareg/namespaces/get-test-namespace")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "get-test-namespace"}), authCtx
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

// TestNamespaceList_AllAuthMethods tests GET namespace list endpoint with OptionalAuth
// All authentication states should return 200
func TestNamespaceList_AllAuthMethods(t *testing.T) {
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
				return testutils.BuildUnauthenticatedRequest(t, "GET", "/v1/terrareg/namespaces"), nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated regular user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				return testutils.BuildAuthenticatedRequest(t, db, "GET", "/v1/terrareg/namespaces", "regular-user", false)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated admin user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				return testutils.BuildAdminRequest(t, db, "GET", "/v1/terrareg/namespaces")
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
