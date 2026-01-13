package testutils

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terraform/v1"
)

// CreateNamespaceHandler creates a fully configured NamespaceHandler
func CreateNamespaceHandler(t *testing.T, db *sqldb.Database, opts ...ConfigOption) *terrareg.NamespaceHandler {
	repos := CreateTestRepositories(t, db, opts...)
	services := CreateTestApplicationServices(t, db, repos, opts...)

	handler, err := terrareg.NewNamespaceHandler(
		services.ListNamespaces,
		services.CreateNamespace,
		services.UpdateNamespace,
		services.DeleteNamespace,
		services.NamespaceDetails,
	)
	require.NoError(t, err, "Failed to create NamespaceHandler")
	return handler
}

// CreateModuleListHandler creates a fully configured ModuleListHandler (v1 Terraform)
func CreateModuleListHandler(t *testing.T, db *sqldb.Database, opts ...ConfigOption) *v1.ModuleListHandler {
	repos := CreateTestRepositories(t, db, opts...)
	services := CreateTestApplicationServices(t, db, repos, opts...)

	return v1.NewModuleListHandler(services.ListModules)
}

// CreateAnalyticsHandler creates a fully configured AnalyticsHandler
func CreateAnalyticsHandler(t *testing.T, db *sqldb.Database, opts ...ConfigOption) *terrareg.AnalyticsHandler {
	repos := CreateTestRepositories(t, db, opts...)
	services := CreateTestApplicationServices(t, db, repos, opts...)

	return terrareg.NewAnalyticsHandler(
		services.GlobalStats,
		services.GlobalUsageStats,
		services.GetDownloadSummary,
		services.RecordModuleDownload,
		services.GetMostRecentlyPublished,
		services.GetMostDownloadedThisWeek,
		services.GetTokenVersions,
	)
}

// CreateAuditHandler creates a fully configured AuditHandler
func CreateAuditHandler(t *testing.T, db *sqldb.Database, opts ...ConfigOption) *terrareg.AuditHandler {
	repos := CreateTestRepositories(t, db, opts...)
	services := CreateTestApplicationServices(t, db, repos, opts...)

	handler := terrareg.NewAuditHandler(services.GetAuditHistory)
	return handler
}
