// Package terrareg_test provides integration tests for the terrareg HTTP handlers
package terrareg_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	moduleQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	namespaceService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	analyticsRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/analytics"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestModuleHandler_HandleModuleList_Success tests the module list endpoint
func TestModuleHandler_HandleModuleList_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "testnamespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Update version to published
	published := true
	moduleVersion.Published = &published
	db.DB.Save(&moduleVersion)

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	listModulesQuery := moduleQuery.NewListModulesQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		listModulesQuery,
		nil, // search not used
		nil, // get provider not used
		nil, // list providers not used
		analyticsRepository,
	)

	// Create request
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleList(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "modules")
	modules := response["modules"].([]interface{})
	assert.GreaterOrEqual(t, len(modules), 1)

	// Check that our test module is in the list
	found := false
	for _, m := range modules {
		module := m.(map[string]interface{})
		if module["namespace"] == "testnamespace" && module["name"] == "testmodule" {
			found = true
			assert.Equal(t, "aws", module["provider"])
			break
		}
	}
	assert.True(t, found, "test module not found in response")
}

// TestModuleHandler_HandleModuleList_Empty tests module list with empty results
func TestModuleHandler_HandleModuleList_Empty(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Don't create any test data

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	listModulesQuery := moduleQuery.NewListModulesQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		listModulesQuery,
		nil,
		nil,
		nil,
		analyticsRepository,
	)

	// Create request
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleList(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 0)
}

// TestModuleHandler_HandleNamespaceModules_Success tests the namespace modules endpoint
func TestModuleHandler_HandleNamespaceModules_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "mynamespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "mymodule", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Update version to published
	published := true
	moduleVersion.Published = &published
	db.DB.Save(&moduleVersion)

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	listModulesQuery := moduleQuery.NewListModulesQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		listModulesQuery,
		nil,
		nil,
		nil,
		analyticsRepository,
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/modules/mynamespace", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "mynamespace")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleNamespaceModules(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "modules")
	modules := response["modules"].([]interface{})
	assert.GreaterOrEqual(t, len(modules), 1)

	// Verify all modules are from the requested namespace
	for _, m := range modules {
		module := m.(map[string]interface{})
		assert.Equal(t, "mynamespace", module["namespace"])
	}
}

// TestModuleHandler_HandleNamespaceModules_NotFound tests namespace modules with non-existent namespace
func TestModuleHandler_HandleNamespaceModules_NotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler (no test data)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	listModulesQuery := moduleQuery.NewListModulesQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		listModulesQuery,
		nil,
		nil,
		nil,
		analyticsRepository,
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/modules/nonexistent", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleNamespaceModules(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 0)
}

// TestModuleHandler_HandleModuleDetails_Success tests the module details endpoint
// Python reference: /app/test/unit/terrareg/server/test_api_module_details.py - test_existing_module
func TestModuleHandler_HandleModuleDetails_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "testns", nil)
	moduleProvider1 := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "aws")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider1.ID, "1.0.0")
	moduleProvider2 := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "azure")
	_ = testutils.CreatePublishedModuleVersion(t, db, moduleProvider2.ID, "2.0.0")

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	listModuleProvidersQuery := moduleQuery.NewListModuleProvidersQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil, // list not used
		nil, // search not used
		nil, // get provider not used
		listModuleProvidersQuery,
		analyticsRepository,
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/modules/testns/testmodule", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "testns")
	rctx.URLParams.Add("name", "testmodule")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleDetails(w, req)

	// Assert - Comprehensive validation matching Python pattern
	// Python reference: assert res.json == {'meta': {'limit': 10, 'current_offset': 0}, 'modules': [...]}
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Validate meta structure (Python validates pagination metadata)
	assert.Contains(t, response, "meta")
	meta := response["meta"].(map[string]interface{})
	assert.Contains(t, meta, "limit")
	assert.Contains(t, meta, "current_offset")

	// Validate modules array exists and has expected providers
	assert.Contains(t, response, "modules")
	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 2, "Should return exactly two module providers (aws and azure)")

	// Validate all module fields (Python validates complete response)
	// Python reference: {'id': 'testnamespace/lonelymodule/testprovider/1.0.0', 'owner': 'Mock Owner', ...}
	expectedProviders := []string{"aws", "azure"}
	for i, m := range modules {
		module := m.(map[string]interface{})

		// Validate all required fields exist (matching Python's complete JSON structure)
		assert.Contains(t, module, "id")
		assert.NotEmpty(t, module["id"], "Module ID should not be empty")

		assert.Equal(t, "testns", module["namespace"])
		assert.Equal(t, "testmodule", module["name"])
		assert.Equal(t, expectedProviders[i], module["provider"])

		assert.Contains(t, module, "verified")
		assert.IsType(t, false, module["verified"], "Verified should be a boolean")

		assert.Contains(t, module, "trusted")
		assert.IsType(t, false, module["trusted"], "Trusted should be a boolean")

		// Optional fields - validate if present
		if owner, ok := module["owner"]; ok && owner != nil {
			assert.NotEmpty(t, owner, "Owner should not be empty if present")
		}

		if description, ok := module["description"]; ok && description != nil {
			assert.NotEmpty(t, description, "Description should not be empty if present")
		}

		if source, ok := module["source"]; ok && source != nil {
			assert.NotEmpty(t, source, "Source should not be empty if present")
		}

		assert.Contains(t, module, "published_at")
		assert.NotNil(t, module["published_at"], "Published at should be present for published version")

		assert.Contains(t, module, "downloads")
		assert.IsType(t, float64(0), module["downloads"], "Downloads should be a number")
	}
}

// TestModuleHandler_HandleModuleProviderDetails_Success tests the module provider details endpoint
func TestModuleHandler_HandleModuleProviderDetails_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "hashicorp", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "consul", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Update version to published
	published := true
	moduleVersion.Published = &published
	db.DB.Save(&moduleVersion)

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	getModuleProviderQuery := moduleQuery.NewGetModuleProviderQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		nil,
		getModuleProviderQuery,
		nil,
		analyticsRepository,
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/modules/hashicorp/consul/aws", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "hashicorp")
	rctx.URLParams.Add("name", "consul")
	rctx.URLParams.Add("provider", "aws")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleProviderDetails(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "hashicorp/consul/aws", response["id"])
	assert.Equal(t, "hashicorp", response["namespace"])
	assert.Equal(t, "consul", response["name"])
	assert.Equal(t, "aws", response["provider"])
	assert.Contains(t, response, "downloads")
}

// TestModuleHandler_HandleModuleProviderDetails_NotFound tests provider not found
func TestModuleHandler_HandleModuleProviderDetails_NotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler (no test data)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	getModuleProviderQuery := moduleQuery.NewGetModuleProviderQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		nil,
		getModuleProviderQuery,
		nil,
		analyticsRepository,
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/modules/unknown/module/provider", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "unknown")
	rctx.URLParams.Add("name", "module")
	rctx.URLParams.Add("provider", "provider")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleProviderDetails(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Contains(t, response["error"], "not found")
}

// TestModuleHandler_HandleModuleSearch_Success tests the module search endpoint
func TestModuleHandler_HandleModuleSearch_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "searchns", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "networking-module", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Update version to published
	published := true
	moduleVersion.Published = &published
	db.DB.Save(&moduleVersion)

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	searchModulesQuery, err := moduleQuery.NewSearchModulesQuery(moduleProviderRepository)
	require.NoError(t, err)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		searchModulesQuery,
		nil,
		nil,
		analyticsRepository,
	)

	// Create request with query parameters
	req := httptest.NewRequest("GET", "/v1/modules/search?q=networking", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleSearch(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "modules")
	assert.Contains(t, response, "meta")

	meta := response["meta"].(map[string]interface{})

	assert.Equal(t, float64(20), meta["limit"])
	assert.Equal(t, float64(0), meta["current_offset"])
}

// TestModuleHandler_HandleModuleSearch_WithFilters tests search with filters
func TestModuleHandler_HandleModuleSearch_WithFilters(t *testing.T) {
	tests := []struct {
		name        string
		queryString string
	}{
		{
			name:        "with namespace filter",
			queryString: "?q=test&namespace=searchns",
		},
		{
			name:        "with provider filter",
			queryString: "?q=test&provider=aws",
		},
		{
			name:        "with verified filter",
			queryString: "?q=test&verified=true",
		},
		{
			name:        "with custom pagination",
			queryString: "?q=test&limit=10&offset=5",
		},
		{
			name:        "with multiple filters",
			queryString: "?q=test&namespace=testns&provider=aws&verified=true&limit=10&offset=0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testutils.SetupTestDatabase(t)
			defer testutils.CleanupTestDatabase(t, db)

			// Create handler
			namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
			domainConfig := testutils.CreateTestDomainConfig(t)
			moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
			require.NoError(t, err)
			searchModulesQuery, err := moduleQuery.NewSearchModulesQuery(moduleProviderRepository)
			require.NoError(t, err)
			namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
			analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
			require.NoError(t, err)

			handler := terrareg.NewModuleReadHandlerForTesting(
				nil,
				searchModulesQuery,
				nil,
				nil,
				analyticsRepository,
			)

			// Create request with query parameters
			req := httptest.NewRequest("GET", "/v1/modules/search"+tt.queryString, nil)
			w := httptest.NewRecorder()

			// Act
			handler.HandleModuleSearch(w, req)

			// Assert
			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Contains(t, response, "modules")
			assert.Contains(t, response, "meta")
		})
	}
}

// TestModuleHandler_HandleModuleSearch_EmptyResults tests search with no matching results
func TestModuleHandler_HandleModuleSearch_EmptyResults(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler (no test data)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	searchModulesQuery, err := moduleQuery.NewSearchModulesQuery(moduleProviderRepository)
	require.NoError(t, err)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		searchModulesQuery,
		nil,
		nil,
		analyticsRepository,
	)

	// Create request
	req := httptest.NewRequest("GET", "/v1/modules/search?q=nonexistentxyz123", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleSearch(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	modules := response["modules"].([]interface{})
	assert.Len(t, modules, 0)
}

// TestModuleHandler_HandleModuleProviderDetails_WithVersion tests provider details with versions
func TestModuleHandler_HandleModuleProviderDetails_WithVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data with version
	namespace := testutils.CreateNamespace(t, db, "versionns", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "versionmodule", "aws")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	// Update version to published
	published := true
	moduleVersion.Published = &published
	db.DB.Save(&moduleVersion)

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	getModuleProviderQuery := moduleQuery.NewGetModuleProviderQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		nil,
		getModuleProviderQuery,
		nil,
		analyticsRepository,
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v1/modules/versionns/versionmodule/aws", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "versionns")
	rctx.URLParams.Add("name", "versionmodule")
	rctx.URLParams.Add("provider", "aws")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleProviderDetails(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "versionns/versionmodule/aws", response["id"])
	// Should contain version information
	assert.Contains(t, response, "published_at")
}

// TestModuleHandler_MultipleProviders tests handler with multiple providers for the same module
func TestModuleHandler_MultipleProviders(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data with multiple providers
	namespace := testutils.CreateNamespace(t, db, "multins", nil)
	testutils.CreateModuleProvider(t, db, namespace.ID, "multimodule", "aws")
	testutils.CreateModuleProvider(t, db, namespace.ID, "multimodule", "azure")
	testutils.CreateModuleProvider(t, db, namespace.ID, "multimodule", "gcp")

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	listModuleProvidersQuery := moduleQuery.NewListModuleProvidersQuery(moduleProviderRepository)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)
	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	require.NoError(t, err)

	handler := terrareg.NewModuleReadHandlerForTesting(
		nil,
		nil,
		nil,
		listModuleProvidersQuery,
		analyticsRepository,
	)

	// Create request
	req := httptest.NewRequest("GET", "/v1/modules/multins/multimodule", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "multins")
	rctx.URLParams.Add("name", "multimodule")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleModuleDetails(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	modules := response["modules"].([]interface{})
	assert.GreaterOrEqual(t, len(modules), 3)

	providers := make([]string, 0)
	for _, m := range modules {
		module := m.(map[string]interface{})
		providers = append(providers, module["provider"].(string))
	}
	assert.Contains(t, providers, "aws")
	assert.Contains(t, providers, "azure")
	assert.Contains(t, providers, "gcp")
}
