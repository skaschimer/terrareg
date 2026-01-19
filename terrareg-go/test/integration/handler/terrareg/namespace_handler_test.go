package terrareg_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestNamespaceHandler_HandleNamespaceList_Success tests successful namespace list retrieval
// Python reference: /app/test/unit/terrareg/server/test_api_terrareg_namespace_list.py - test_with_namespaces_present
func TestNamespaceHandler_HandleNamespaceList_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespaces
	testutils.CreateNamespace(t, db, "namespace1", nil)
	testutils.CreateNamespace(t, db, "namespace2", nil)

	// Create handler using test utils
	handler := testutils.CreateNamespaceHandler(t, db)

	req := httptest.NewRequest("GET", "/v1/terrareg/namespaces", nil)
	w := httptest.NewRecorder()

	handler.HandleNamespaceList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// For array responses, unmarshal directly
	var response []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON array")
	assert.Len(t, response, 2)

	// Validate all namespace fields (Python validates complete response)
	// Python reference: assert res.json == [{'name': 'testnamespace', 'view_href': '/modules/testnamespace', 'display_name': None}, ...]
	for _, ns := range response {
		namespace := ns.(map[string]interface{})

		// Validate all required fields exist
		assert.Contains(t, namespace, "name")
		assert.NotEmpty(t, namespace["name"], "Namespace name should not be empty")

		assert.Contains(t, namespace, "view_href")
		assert.NotEmpty(t, namespace["view_href"], "View href should not be empty")
		// Verify view_href format
		viewHref := namespace["view_href"].(string)
		assert.Contains(t, viewHref, "/modules/", "View href should contain /modules/ path")

		// display_name may be nil or a string
		if displayName, ok := namespace["display_name"]; ok && displayName != nil {
			assert.IsType(t, "", displayName, "Display name should be a string if present")
		}
	}
}

// TestNamespaceHandler_HandleNamespaceList_Empty tests namespace list with no data
func TestNamespaceHandler_HandleNamespaceList_Empty(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler with no data
	handler := testutils.CreateNamespaceHandler(t, db)

	req := httptest.NewRequest("GET", "/v1/terrareg/namespaces", nil)
	w := httptest.NewRecorder()

	handler.HandleNamespaceList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// For array responses, unmarshal directly
	var response []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON array")
	assert.Len(t, response, 0)
}

// TestNamespaceHandler_HandleNamespaceList_WithPagination tests pagination support
func TestNamespaceHandler_HandleNamespaceList_WithPagination(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespaces
	testutils.CreateNamespace(t, db, "namespace1", nil)
	testutils.CreateNamespace(t, db, "namespace2", nil)

	// Create handler
	handler := testutils.CreateNamespaceHandler(t, db)

	// Request with pagination
	params := url.Values{}
	params.Add("limit", "10")
	params.Add("offset", "0")

	req := httptest.NewRequest("GET", "/v1/terrareg/namespaces?"+params.Encode(), nil)
	w := httptest.NewRecorder()

	handler.HandleNamespaceList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	// With pagination, should return wrapped object with "namespaces" key
	assert.Contains(t, response, "namespaces")

	namespaces := response["namespaces"].([]interface{})
	assert.Len(t, namespaces, 2)
}

// TestNamespaceHandler_HandleNamespaceList_MultipleNamespaces tests with multiple namespaces
// Python reference: /app/test/unit/terrareg/server/test_api_terrareg_namespace_list.py - test_with_namespaces_present
func TestNamespaceHandler_HandleNamespaceList_MultipleNamespaces(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create multiple namespaces with display names
	displayName1 := "Test Namespace One"
	displayName3 := "Test Namespace Three"
	testutils.CreateNamespace(t, db, "namespace1", &displayName1)
	testutils.CreateNamespace(t, db, "namespace2", nil)
	testutils.CreateNamespace(t, db, "namespace3", &displayName3)
	testutils.CreateNamespace(t, db, "namespace4", nil)
	testutils.CreateNamespace(t, db, "namespace5", nil)

	handler := testutils.CreateNamespaceHandler(t, db)

	req := httptest.NewRequest("GET", "/v1/terrareg/namespaces", nil)
	w := httptest.NewRecorder()

	handler.HandleNamespaceList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// For array responses, unmarshal directly
	var response []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON array")
	assert.Len(t, response, 5)

	// Validate namespace names and fields (Python validates exact field values)
	// Python reference: assert res.json == [{'name': 'testnamespace', 'view_href': '/modules/testnamespace', 'display_name': None}, ...]
	expectedNamespaces := map[string]string{
		"namespace1": "Test Namespace One",
		"namespace2": "",
		"namespace3": "Test Namespace Three",
		"namespace4": "",
		"namespace5": "",
	}

	// Build a map of namespace names found in response
	foundNamespaces := make(map[string]map[string]interface{})
	for _, ns := range response {
		namespace := ns.(map[string]interface{})
		name := namespace["name"].(string)
		foundNamespaces[name] = namespace

		// Validate all required fields exist
		assert.Contains(t, namespace, "name")
		assert.NotEmpty(t, namespace["name"])

		assert.Contains(t, namespace, "view_href")
		assert.NotEmpty(t, namespace["view_href"])
		assert.Contains(t, namespace["view_href"], "/modules/")

		// Validate display_name
		assert.Contains(t, namespace, "display_name")
	}

	// Verify expected namespaces are present with correct display names
	for expectedName, expectedDisplayName := range expectedNamespaces {
		namespace, found := foundNamespaces[expectedName]
		assert.True(t, found, "Expected namespace '%s' not found in response", expectedName)

		if expectedDisplayName != "" {
			assert.Equal(t, expectedDisplayName, namespace["display_name"])
		}
	}
}

// TestNamespaceHandler_HandleNamespaceDetails_Success tests successful namespace details retrieval
func TestNamespaceHandler_HandleNamespaceDetails_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespace
	testutils.CreateNamespace(t, db, "test-namespace", nil)

	// Create handler
	handler := testutils.CreateNamespaceHandler(t, db)

	req := httptest.NewRequest("GET", "/v1/terrareg/namespaces/test-namespace", nil)
	w := httptest.NewRecorder()

	// Add Chi context for path parameter
	req = testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace"})

	handler.HandleNamespaceDetails(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Contains(t, response, "name")
	assert.Equal(t, "test-namespace", response["name"])
}

// TestNamespaceHandler_HandleNamespaceDetails_NotFound tests with non-existent namespace
func TestNamespaceHandler_HandleNamespaceDetails_NotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	handler := testutils.CreateNamespaceHandler(t, db)

	req := httptest.NewRequest("GET", "/v1/terrareg/namespaces/nonexistent", nil)
	w := httptest.NewRecorder()

	req = testutils.AddChiContext(t, req, map[string]string{"namespace": "nonexistent"})

	handler.HandleNamespaceDetails(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Equal(t, map[string]interface{}{}, response)
}

// TestNamespaceHandler_HandleNamespaceDetails_MissingParameter tests with missing namespace parameter
func TestNamespaceHandler_HandleNamespaceDetails_MissingParameter(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	handler := testutils.CreateNamespaceHandler(t, db)

	req := httptest.NewRequest("GET", "/v1/terrareg/namespaces/", nil)
	w := httptest.NewRecorder()

	req = testutils.AddChiContext(t, req, map[string]string{"namespace": ""})

	handler.HandleNamespaceDetails(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "error")
	assert.Contains(t, response["error"].(string), "namespace is required")
}

// TestNamespaceHandler_HandleNamespaceCreate_Success tests successful namespace creation
func TestNamespaceHandler_HandleNamespaceCreate_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	handler := testutils.CreateNamespaceHandler(t, db)

	// Create request body
	displayName := "New Namespace"
	requestBody := dto.NamespaceCreateRequest{
		Name:        "new-namespace",
		DisplayName: &displayName,
		Type:        "NONE",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/v1/terrareg/namespaces", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleNamespaceCreate(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Contains(t, response, "name")
	assert.Equal(t, "new-namespace", response["name"])
	assert.Contains(t, response, "display_name")
}

// TestNamespaceHandler_HandleNamespaceCreate_InvalidJSON tests with invalid JSON
func TestNamespaceHandler_HandleNamespaceCreate_InvalidJSON(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	handler := testutils.CreateNamespaceHandler(t, db)

	req := httptest.NewRequest("POST", "/v1/terrareg/namespaces", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleNamespaceCreate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "error")
}

// TestNamespaceHandler_HandleNamespaceDelete_Success tests successful namespace deletion
func TestNamespaceHandler_HandleNamespaceDelete_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create namespace to delete
	_ = testutils.CreateNamespace(t, db, "delete-me", nil)

	// Create handler
	handler := testutils.CreateNamespaceHandler(t, db)

	req := httptest.NewRequest("DELETE", "/v1/terrareg/namespaces/delete-me", nil)
	w := httptest.NewRecorder()

	req = testutils.AddChiContext(t, req, map[string]string{"namespace": "delete-me"})

	handler.HandleNamespaceDelete(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Equal(t, map[string]interface{}{}, response)

	// Verify namespace was deleted
	repos := testutils.CreateTestRepositories(t, db)
	namespaces, err := repos.Namespace.List(requireContext(t))
	require.NoError(t, err)
	assert.Empty(t, namespaces)
}

// TestNamespaceHandler_HandleNamespaceDelete_NotFound tests deleting non-existent namespace
func TestNamespaceHandler_HandleNamespaceDelete_NotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	handler := testutils.CreateNamespaceHandler(t, db)

	req := httptest.NewRequest("DELETE", "/v1/terrareg/namespaces/nonexistent", nil)
	w := httptest.NewRecorder()

	req = testutils.AddChiContext(t, req, map[string]string{"namespace": "nonexistent"})

	handler.HandleNamespaceDelete(w, req)

	// Should return error for non-existent namespace
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "error")
}

// TestNamespaceHandler_HandleNamespaceUpdate_Success tests successful namespace update
func TestNamespaceHandler_HandleNamespaceUpdate_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create namespace
	testutils.CreateNamespace(t, db, "update-namespace", nil)

	// Create handler
	handler := testutils.CreateNamespaceHandler(t, db)

	// Create request body
	displayName := "Updated Display Name"
	requestBody := dto.NamespaceUpdateRequest{
		DisplayName: &displayName,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/v1/terrareg/namespaces/update-namespace", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	req = testutils.AddChiContext(t, req, map[string]string{"namespace": "update-namespace"})

	handler.HandleNamespaceUpdate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Contains(t, response, "name")
	assert.Equal(t, "update-namespace", response["name"])
	assert.Contains(t, response, "display_name")
}

// TestNamespaceHandler_HandleNamespaceUpdate_MissingParameter tests with missing namespace parameter
func TestNamespaceHandler_HandleNamespaceUpdate_MissingParameter(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	handler := testutils.CreateNamespaceHandler(t, db)

	displayName := "Test"
	requestBody := dto.NamespaceUpdateRequest{
		DisplayName: &displayName,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/v1/terrareg/namespaces/", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	req = testutils.AddChiContext(t, req, map[string]string{"namespace": ""})

	handler.HandleNamespaceUpdate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "error")
}

// TestNamespaceHandler_HandleNamespaceUpdate_InvalidJSON tests with invalid JSON
func TestNamespaceHandler_HandleNamespaceUpdate_InvalidJSON(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	handler := testutils.CreateNamespaceHandler(t, db)

	req := httptest.NewRequest("POST", "/v1/terrareg/namespaces/test", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	req = testutils.AddChiContext(t, req, map[string]string{"namespace": "test"})

	handler.HandleNamespaceUpdate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "error")
}

func requireContext(t *testing.T) context.Context {
	ctx := context.Background()
	return ctx
}
