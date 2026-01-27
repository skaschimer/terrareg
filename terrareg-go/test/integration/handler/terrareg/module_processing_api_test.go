package terrareg_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	moduleCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestModuleUpload_APIParity tests ZIP upload endpoint matches expected API behavior
func TestModuleUpload_APIParity(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespace and module provider
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")

	// Create test ZIP file
	zipPath := createTestModuleArchive(t)
	defer os.Remove(zipPath)

	// Setup request with authentication
	req := buildUploadRequest(t, namespace.Name, "test-module", "aws", "1.0.0", zipPath)
	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": namespace.Name,
		"name":      "test-module",
		"provider":  "aws",
		"version":   "1.0.0",
	})

	// Setup mock command
	mockCmd := &mockProcessModuleCommand{}
	handler := terrareg.NewModuleHandlerForTesting(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		mockCmd, nil, nil,
	)

	w := httptest.NewRecorder()
	handler.HandleModuleVersionUpload(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "message")
}

// TestModuleUpload_InvalidFile tests upload with invalid file
func TestModuleUpload_InvalidFile(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespace and module provider
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	_ = testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")

	// Setup request with invalid content
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "invalid.txt")
	part.Write([]byte("not a zip file"))
	writer.Close()

	req := httptest.NewRequest("POST",
		"/v1/terrareg/modules/test-namespace/test-module/aws/1.0.0/upload",
		body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "test-namespace",
		"name":      "test-module",
		"provider":  "aws",
		"version":   "1.0.0",
	})

	// Setup handler
	domainConfig := testutils.CreateTestDomainConfig(t)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	handler := terrareg.NewModuleUploadHandlerForTesting(domainConfig, nil, nil, nil, moduleProviderRepository)

	w := httptest.NewRecorder()
	handler.HandleModuleVersionUpload(w, req)

	// Should return error (400 or 500 depending on implementation)
	assert.NotEqual(t, http.StatusOK, w.Code)
}

// TestModuleUpload_AuthenticationRequired tests upload requires authentication
func TestModuleUpload_AuthenticationRequired(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespace
	_ = testutils.CreateNamespace(t, db, "test-namespace", nil)

	// Setup unauthenticated request
	zipPath := createTestModuleArchive(t)
	defer os.Remove(zipPath)

	req := buildUploadRequest(t, "test-namespace", "test-module", "aws", "1.0.0", zipPath)
	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "test-namespace",
		"name":      "test-module",
		"provider":  "aws",
		"version":   "1.0.0",
	})

	// Setup handler
	domainConfig := testutils.CreateTestDomainConfig(t)
	handler := terrareg.NewModuleUploadHandlerForTesting(domainConfig, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	handler.HandleModuleVersionUpload(w, req)

	// Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestModuleImport_APIParity tests tag-based import endpoint matches expected API behavior
func TestModuleImport_APIParity(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespace and module provider with git config
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	moduleProvider := testutils.CreateModuleProviderWithGit(t, db, namespace.ID, "test-module", "aws")

	// Setup request body
	reqBody := map[string]interface{}{
		"version": "1.0.0",
		"git_tag": "v1.0.0",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST",
		"/v1/terrareg/modules/test-namespace/test-module/aws/import",
		bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "test-namespace",
		"name":      "test-module",
		"provider":  "aws",
	})

	// Setup mock command
	mockCmd := &mockProcessModuleCommand{}
	handler := terrareg.NewModuleHandlerForTesting(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		mockCmd, nil, nil,
	)

	w := httptest.NewRecorder()
	handler.HandleModuleVersionImport(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Success", response["status"])
}

// TestModuleImport_InvalidVersion tests import with invalid version
func TestModuleImport_InvalidVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespace and module provider
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	_ = testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")

	// Setup request body with invalid version
	reqBody := map[string]interface{}{
		"version": "invalid",
		"git_tag": "v1.0.0",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST",
		"/v1/terrareg/modules/test-namespace/test-module/aws/import",
		bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "test-namespace",
		"name":      "test-module",
		"provider":  "aws",
	})

	// Setup handler
	domainConfig := testutils.CreateTestDomainConfig(t)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, domainConfig)
	require.NoError(t, err)
	handler := terrareg.NewModuleImportHandlerForTesting(domainConfig, nil, nil, nil, moduleProviderRepository)

	w := httptest.NewRecorder()
	handler.HandleModuleVersionImport(w, req)

	// Should return error
	var response map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "errors")
}

// TestModuleProcessing_CompletePipeline tests the complete processing pipeline
func TestModuleProcessing_CompletePipeline(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespace and module provider with git config
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	moduleProvider := testutils.CreateModuleProviderWithGit(t, db, namespace.ID, "test-module", "aws")

	// Execute import
	reqBody := map[string]interface{}{
		"version": "1.0.0",
		"git_tag": "v1.0.0",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST",
		"/v1/terrareg/modules/test-namespace/test-module/aws/import",
		bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "test-namespace",
		"name":      "test-module",
		"provider":  "aws",
	})

	// Setup mock command with successful result
	mockCmd := &mockProcessModuleCommand{
		executeFunc: func(ctx context.Context, req moduleCmd.ProcessModuleRequest) error {
			// Simulate successful processing
			return nil
		},
	}

	handler := terrareg.NewModuleHandlerForTesting(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		mockCmd, nil, nil,
	)

	w := httptest.NewRecorder()
	handler.HandleModuleVersionImport(w, req)

	// Verify successful processing
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify module version was created in database
	var moduleVersionCount int64
	db.DB.Table("module_version").
		Joins("JOIN module_provider ON module_version.module_provider_id = module_provider.id").
		Joins("JOIN module ON module_provider.module_id = module.id").
		Joins("JOIN namespace ON module.namespace_id = namespace.id").
		Where("namespace.name = ? AND module.name = ? AND module_provider.provider = ? AND module_version.version = ?",
			"test-namespace", "test-module", "aws", "1.0.0").
		Count(&moduleVersionCount)

	assert.Greater(t, moduleVersionCount, int64(0), "Module version should be created")
}

// Helper functions

// mockProcessModuleCommand is a mock for ProcessModuleCommand
type mockProcessModuleCommand struct {
	executeFunc func(ctx context.Context, req moduleCmd.ProcessModuleRequest) error
}

func (m *mockProcessModuleCommand) Execute(ctx context.Context, req moduleCmd.ProcessModuleRequest) error {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, req)
	}
	return nil
}

// buildUploadRequest builds a multipart upload request for testing
func buildUploadRequest(t *testing.T, namespace, module, provider, version, zipPath string) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(zipPath)
	require.NoError(t, err)
	defer file.Close()

	part, err := writer.CreateFormFile("file", "module.zip")
	require.NoError(t, err)

	_, err = io.Copy(part, file)
	require.NoError(t, err)

	writer.Close()

	req := httptest.NewRequest("POST",
		strings.Join([]string{"/v1/terrareg/modules", namespace, module, provider, version, "upload"}, "/"),
		body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

// createTestModuleArchive creates a simple test module ZIP archive
func createTestModuleArchive(t *testing.T) string {
	tempDir := t.TempDir()
	zipPath := tempDir + "/module.zip"

	// Create a simple ZIP file with test terraform files
	// This is a minimal implementation - in production, use archive/zip
	body := &bytes.Buffer{}
	body.WriteString("PK\x03\x04") // ZIP file header
	body.WriteString(strings.Repeat("\x00", 100)) // Minimal ZIP content

	err := os.WriteFile(zipPath, body.Bytes(), 0644)
	require.NoError(t, err)

	return zipPath
}

// CreateModuleProviderWithGit creates a module provider with git configuration for testing
func CreateModuleProviderWithGit(t *testing.T, db *testutils.TestDatabase, namespaceID int, moduleName, provider string) *testutils.ModuleProvider {
	module := testutils.CreateModule(t, db, namespaceID, moduleName)

	cloneURL := "https://github.com/test/{namespace}/{module}.git"
	tagFormat := "v{version}"

	moduleProvider := &testutils.ModuleProvider{
		ModuleID:             module.ID,
		Provider:             provider,
		RepoCloneURLTemplate: &cloneURL,
		GitTagFormat:         &tagFormat,
	}

	err := db.DB.Create(moduleProvider).Error
	require.NoError(t, err)

	return moduleProvider
}
