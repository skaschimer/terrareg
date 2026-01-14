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

// TestConfig_AllAuthMethods tests GET /v1/terrareg/config endpoint with OptionalAuth
// All authentication states should return 200
func TestConfig_AllAuthMethods(t *testing.T) {
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
				return testutils.BuildUnauthenticatedRequest(t, "GET", "/v1/terrareg/config"), nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated regular user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				return testutils.BuildAuthenticatedRequest(t, db, "GET", "/v1/terrareg/config", "regular-user", false)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated admin user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				return testutils.BuildAdminRequest(t, db, "GET", "/v1/terrareg/config")
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

// TestHealth_AllAuthMethods tests GET /v1/terrareg/health endpoint with OptionalAuth
func TestHealth_AllAuthMethods(t *testing.T) {
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
				return testutils.BuildUnauthenticatedRequest(t, "GET", "/v1/terrareg/health"), nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated regular user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				return testutils.BuildAuthenticatedRequest(t, db, "GET", "/v1/terrareg/health", "regular-user", false)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated admin user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				return testutils.BuildAdminRequest(t, db, "GET", "/v1/terrareg/health")
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

// TestVersion_AllAuthMethods tests GET /v1/terrareg/version endpoint with OptionalAuth
func TestVersion_AllAuthMethods(t *testing.T) {
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
				return testutils.BuildUnauthenticatedRequest(t, "GET", "/v1/terrareg/version"), nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated regular user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				return testutils.BuildAuthenticatedRequest(t, db, "GET", "/v1/terrareg/version", "regular-user", false)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated admin user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				return testutils.BuildAdminRequest(t, db, "GET", "/v1/terrareg/version")
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

// TestModuleList_AllAuthMethods tests GET /v1/terrareg/modules/{namespace} endpoint with OptionalAuth
func TestModuleList_AllAuthMethods(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "module-list-namespace")

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) (*http.Request, *model.AuthContext)
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				req := testutils.BuildUnauthenticatedRequest(t, "GET", "/v1/terrareg/modules/module-list-namespace")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "module-list-namespace"}), nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated regular user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				req, authCtx := testutils.BuildAuthenticatedRequest(t, db, "GET", "/v1/terrareg/modules/module-list-namespace", "regular-user", false)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "module-list-namespace"}), authCtx
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated admin user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				req, authCtx := testutils.BuildAdminRequest(t, db, "GET", "/v1/terrareg/modules/module-list-namespace")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "module-list-namespace"}), authCtx
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

// TestModuleVersions_AllAuthMethods tests GET /v1/terrareg/modules/{namespace}/{name}/{provider}/versions endpoint with OptionalAuth
func TestModuleVersions_AllAuthMethods(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "version-list-namespace")
	_ = testutils.CreateModuleProvider(t, db, namespace.ID, "testmod", "testprovider")

	tests := []struct {
		name           string
		setupAuth      func(*testing.T, *sqldb.Database) (*http.Request, *model.AuthContext)
		expectedStatus int
	}{
		{
			name: "unauthenticated request returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				req := testutils.BuildUnauthenticatedRequest(t, "GET", "/v1/terrareg/modules/version-list-namespace/testmod/testprovider/versions")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "version-list-namespace", "name": "testmod", "provider": "testprovider"}), nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated regular user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				req, authCtx := testutils.BuildAuthenticatedRequest(t, db, "GET", "/v1/terrareg/modules/version-list-namespace/testmod/testprovider/versions", "regular-user", false)
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "version-list-namespace", "name": "testmod", "provider": "testprovider"}), authCtx
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated admin user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				req, authCtx := testutils.BuildAdminRequest(t, db, "GET", "/v1/terrareg/modules/version-list-namespace/testmod/testprovider/versions")
				return testutils.AddChiContext(t, req, map[string]string{"namespace": "version-list-namespace", "name": "testmod", "provider": "testprovider"}), authCtx
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

// TestAnalytics_AllAuthMethods tests GET /v1/terrareg/analytics/* endpoints with OptionalAuth
func TestAnalytics_AllAuthMethods(t *testing.T) {
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
				return testutils.BuildUnauthenticatedRequest(t, "GET", "/v1/terrareg/analytics/global/stats_summary"), nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated regular user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				return testutils.BuildAuthenticatedRequest(t, db, "GET", "/v1/terrareg/analytics/global/stats_summary", "regular-user", false)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "authenticated admin user returns 200",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				return testutils.BuildAdminRequest(t, db, "GET", "/v1/terrareg/analytics/global/stats_summary")
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
