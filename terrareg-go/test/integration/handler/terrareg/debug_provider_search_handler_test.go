package terrareg

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

func TestDebugProviderSearchHandler(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Setup test data
	testutils.SetupComprehensiveProviderSearchTestData(t, db)

	// Create container and handler
	cont := testutils.CreateTestContainer(t, db)
	handler := cont.Server.ProviderHandler()

	// Create a request
	req := httptest.NewRequest("GET", "/v1/providers/search?q=mixed&include_count=true&limit=6", nil)
	w := httptest.NewRecorder()

	// Execute handler
	handler.HandleProviderSearch(w, req)

	// Check status
	require.Equal(t, http.StatusOK, w.Code)

	// Print response
	t.Logf("Response body: %s", w.Body.String())

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check providers
	providers, ok := response["providers"].([]interface{})
	require.True(t, ok, "providers field missing")
	t.Logf("Providers count: %d", len(providers))

	for i, p := range providers {
		providerMap, ok := p.(map[string]interface{})
		require.True(t, ok)
		id := providerMap["id"].(string)
		t.Logf("  %d. ID: %s", i+1, id)
	}
}
