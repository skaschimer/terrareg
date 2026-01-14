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

// TestAuditHistory_Authentication tests audit history endpoint with RequireAdmin middleware
func TestAuditHistory_Authentication(t *testing.T) {
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
			name: "unauthenticated request returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				return testutils.BuildUnauthenticatedRequest(t, "GET", "/v1/terrareg/audit-history"), nil
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "regular authenticated user returns 403",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				return testutils.BuildAuthenticatedRequest(t, db, "GET", "/v1/terrareg/audit-history", "regular-user", false)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "admin user can access audit history",
			setupAuth: func(t *testing.T, db *sqldb.Database) (*http.Request, *model.AuthContext) {
				return testutils.BuildAdminRequest(t, db, "GET", "/v1/terrareg/audit-history")
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
