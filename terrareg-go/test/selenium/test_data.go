package selenium

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	integrationTestUtils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// UpdateModuleProviderVerified updates a module provider's verified status.
// This is equivalent to Python's provider.update_attributes(verified=True).
// Python reference: /app/test/selenium/test_homepage.py - provider.update_attributes(verified=True)
func UpdateModuleProviderVerified(t *testing.T, db *sqldb.Database, moduleProviderID int, verified bool) {
	verifiedPtr := &verified
	err := db.DB.Model(&sqldb.ModuleProviderDB{}).
		Where("id = ?", moduleProviderID).
		Update("verified", verifiedPtr).Error
	require.NoError(t, err, "Failed to update module provider verified status")
}

// UpdateModuleVersionPublishedAt updates a module version's published_at timestamp.
// This is equivalent to Python's module_version.update_attributes(published_at=datetime.now()).
// Python reference: /app/test/selenium/test_homepage.py - module_version.update_attributes(published_at=datetime.now())
func UpdateModuleVersionPublishedAt(t *testing.T, db *sqldb.Database, moduleVersionID int, publishedAt time.Time) {
	err := db.DB.Model(&sqldb.ModuleVersionDB{}).
		Where("id = ?", moduleVersionID).
		Update("published_at", publishedAt).Error
	require.NoError(t, err, "Failed to update module version published_at")
}

// GetNamespaceByName retrieves a namespace by name from the database.
// This is equivalent to Python's Namespace(name='...').
// Python reference: /app/test/selenium/test_homepage.py - Namespace('mostrecent')
func GetNamespaceByName(t *testing.T, db *sqldb.Database, name string) sqldb.NamespaceDB {
	var namespace sqldb.NamespaceDB
	err := db.DB.Where("namespace = ?", name).First(&namespace).Error
	require.NoError(t, err, "Failed to find namespace: %s", name)
	return namespace
}

// GetModuleProvider retrieves a module provider by namespace, module, and provider names.
// This is equivalent to Python's ModuleProvider(module=Module(namespace=Namespace(name='...'), name='...'), name='...').
// Python reference: /app/test/selenium/test_homepage.py - ModuleProvider lookup
func GetModuleProvider(t *testing.T, db *sqldb.Database, namespaceName, moduleName, providerName string) sqldb.ModuleProviderDB {
	var moduleProvider sqldb.ModuleProviderDB
	err := db.DB.Joins("JOIN namespace_db ON namespace_db.id = module_provider_db.namespace_id").
		Where("namespace_db.namespace = ?", namespaceName).
		Where("module_provider_db.module = ?", moduleName).
		Where("module_provider_db.provider = ?", providerName).
		First(&moduleProvider).Error
	require.NoError(t, err, "Failed to find module provider: %s/%s/%s", namespaceName, moduleName, providerName)
	return moduleProvider
}

// GetModuleVersion retrieves a module version by provider and version.
// This is equivalent to Python's ModuleVersion(module_provider=..., version='...').
// Python reference: /app/test/selenium/test_homepage.py - ModuleVersion lookup
func GetModuleVersion(t *testing.T, db *sqldb.Database, moduleProviderID int, version string) sqldb.ModuleVersionDB {
	var moduleVersion sqldb.ModuleVersionDB
	err := db.DB.Where("module_provider_id = ?", moduleProviderID).
		Where("version = ?", version).
		First(&moduleVersion).Error
	require.NoError(t, err, "Failed to find module version: %d/%s", moduleProviderID, version)
	return moduleVersion
}

// SetupHomepageTestData creates test data for homepage tests.
// This creates the modules and versions needed for the homepage to display properly.
// Python reference: /app/test/selenium/test_homepage.py - TestHomePage data setup
func SetupHomepageTestData(t *testing.T, db *sqldb.Database) {
	// Create "mostrecent" namespace and module for latest module version tests
	mostRecentNs := integrationTestUtils.CreateNamespace(t, db, "mostrecent")
	mostRecentMp := integrationTestUtils.CreateModuleProvider(t, db, mostRecentNs.ID, "modulename", "providername")
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, mostRecentMp.ID, "1.2.3")
	_ = integrationTestUtils.CreateModuleDetails(t, db, "# Test Module\n\nThis is a test module for homepage display.")

	// Create "trustednamespace" for trusted module tests
	trustedNs := integrationTestUtils.CreateNamespace(t, db, "trustednamespace")
	trustedMp := integrationTestUtils.CreateModuleProvider(t, db, trustedNs.ID, "secondlatestmodule", "aws")
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, trustedMp.ID, "4.4.1")
	_ = integrationTestUtils.CreateModuleDetails(t, db, "# Trusted Module\n\nThis is a trusted module.")
}

// SetupSearchTestData creates test data for search tests.
// This creates multiple modules with different attributes for search testing.
// Python reference: /app/test/selenium/test_module_search.py - search test data
func SetupSearchTestData(t *testing.T, db *sqldb.Database) {
	// Create namespaces
	ns1 := integrationTestUtils.CreateNamespace(t, db, "modulesearch")
	_ = integrationTestUtils.CreateNamespace(t, db, "mixedsearch")

	// Create module providers for module search
	mp1 := integrationTestUtils.CreateModuleProvider(t, db, ns1.ID, "modulesearch-trusted", "testprovider")
	mp2 := integrationTestUtils.CreateModuleProvider(t, db, ns1.ID, "modulesearch-result", "testprovider")
	mp3 := integrationTestUtils.CreateModuleProvider(t, db, ns1.ID, "othermodule", "testprovider")

	// Create published versions
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, mp1.ID, "1.0.0")
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, mp2.ID, "1.0.0")
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, mp3.ID, "1.0.0")

	// Create module details
	integrationTestUtils.CreateModuleDetails(t, db, "# Trusted Module")
	integrationTestUtils.CreateModuleDetails(t, db, "# Search Result Module")
	integrationTestUtils.CreateModuleDetails(t, db, "# Other Module")
}

// SetupNamespaceTestData creates test data for namespace page tests.
// This creates a namespace with multiple modules and providers.
// Python reference: /app/test/selenium/test_namespace.py - namespace test data
func SetupNamespaceTestData(t *testing.T, db *sqldb.Database) {
	// Create namespace with various module types
	namespace := integrationTestUtils.CreateNamespace(t, db, "testnamespace")

	// Create a standard module provider
	moduleProvider := integrationTestUtils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")

	// Create a published version
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create module details
	_ = integrationTestUtils.CreateModuleDetails(t, db, "# Test Module\n\nModule description here.")
}

// SetupModuleProviderTestData creates test data for module provider page tests.
// This creates a full module with versions, readme, inputs, outputs, etc.
// Python reference: /app/test/selenium/test_module_provider.py - module provider test data
func SetupModuleProviderTestData(t *testing.T, db *sqldb.Database) {
	// Create namespace
	namespace := integrationTestUtils.CreateNamespace(t, db, "moduledetails")

	// Create module provider with different versions
	moduleProvider := integrationTestUtils.CreateModuleProvider(t, db, namespace.ID, "fullypopulated", "testprovider")

	// Create multiple versions
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.5.0")
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.4.0")
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create module details with full content
	readmeContent := `# Fully Populated Module

This module is fully populated with all details.

## Features

- Feature 1
- Feature 2

## Usage

` + "```hcl\n" + `module "example" {
  source = "moduledetails/fullypopulated/testprovider"
  version = "1.5.0"
}
` + "```\n"
	_ = integrationTestUtils.CreateModuleDetails(t, db, readmeContent)
}

// SetupLoginTestData creates minimal test data for login tests.
// Login tests typically don't need much module data.
// Python reference: /app/test/selenium/test_login.py - login test data
func SetupLoginTestData(t *testing.T, db *sqldb.Database) {
	// Login tests typically don't need any module data
	// Just creating a namespace for basic testing
	_ = integrationTestUtils.CreateNamespace(t, db, "login-test")
}
