package testutils

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// SetupComprehensiveModuleSearchTestData creates comprehensive module search test data
// matching Python's integration_test_data.py.
// This creates 60+ module versions across multiple namespaces.
func SetupComprehensiveModuleSearchTestData(t *testing.T, db *sqldb.Database) {
	t.Helper()

	published := true
	beta := true
	internal := true

	// ===== modulesearch namespace (untrusted, not in TRUSTED_NAMESPACES) =====
	modulesearchNs := CreateNamespace(t, db, "modulesearch", nil)

	// contributedmodule-oneversion
	provider1 := CreateModuleProvider(t, db, modulesearchNs.ID, "contributedmodule-oneversion", "aws")
	createVersion(t, db, provider1.ID, "1.0.0", &published, nil, "TestOwner1", "DESCRIPTION-Search-PUBLISHED")

	// contributedmodule-multiversion
	provider2 := CreateModuleProvider(t, db, modulesearchNs.ID, "contributedmodule-multiversion", "aws")
	createVersion(t, db, provider2.ID, "1.2.3", &published, nil, "TestOwner2", "DESCRIPTION-Search-OLDVERSION")
	createVersion(t, db, provider2.ID, "2.0.0", &published, nil, "", "")

	// contributedmodule-withbetaversion
	provider3 := CreateModuleProvider(t, db, modulesearchNs.ID, "contributedmodule-withbetaversion", "aws")
	createVersion(t, db, provider3.ID, "1.2.3", &published, nil, "", "")
	createVersion(t, db, provider3.ID, "2.0.0-beta", &published, &beta, "", "DESCRIPTION-Search-BETAVERSION")

	// contributedmodule-onlybeta
	provider4 := CreateModuleProvider(t, db, modulesearchNs.ID, "contributedmodule-onlybeta", "aws")
	createVersion(t, db, provider4.ID, "2.5.0-beta", &published, &beta, "", "")

	// contributedmodule-differentprovider
	provider5 := CreateModuleProvider(t, db, modulesearchNs.ID, "contributedmodule-differentprovider", "gcp")
	createVersion(t, db, provider5.ID, "1.2.3", &published, nil, "", "")

	// contributedmodule-unpublished
	provider6 := CreateModuleProvider(t, db, modulesearchNs.ID, "contributedmodule-unpublished", "aws")
	createVersion(t, db, provider6.ID, "1.0.0", nil, nil, "TestOwner1", "DESCRIPTION-Search-UNPUBLISHED")

	// verifiedmodule-oneversion
	provider7 := CreateModuleProviderWithVerified(t, db, modulesearchNs.ID, "verifiedmodule-oneversion", "aws", true)
	createVersion(t, db, provider7.ID, "1.0.0", &published, nil, "", "")

	// verifiedmodule-multiversion
	provider8 := CreateModuleProviderWithVerified(t, db, modulesearchNs.ID, "verifiedmodule-multiversion", "aws", true)
	createVersion(t, db, provider8.ID, "1.2.3", &published, nil, "", "")
	createVersion(t, db, provider8.ID, "2.0.0", &published, nil, "", "")

	// verifiedmodule-withbetaversion
	provider9 := CreateModuleProviderWithVerified(t, db, modulesearchNs.ID, "verifiedmodule-withbetaversion", "aws", true)
	createVersion(t, db, provider9.ID, "1.2.3", &published, nil, "", "")
	createVersion(t, db, provider9.ID, "2.0.0-beta", &published, &beta, "", "")

	// verifiedmodule-onybeta
	provider10 := CreateModuleProviderWithVerified(t, db, modulesearchNs.ID, "verifiedmodule-onybeta", "aws", true)
	createVersion(t, db, provider10.ID, "2.0.0-beta", &published, &beta, "", "")

	// verifiedmodule-differentprovider
	provider11 := CreateModuleProviderWithVerified(t, db, modulesearchNs.ID, "verifiedmodule-differentprovider", "gcp", true)
	createVersion(t, db, provider11.ID, "1.2.3", &published, nil, "", "")

	// verifiedmodule-unpublished
	provider12 := CreateModuleProviderWithVerified(t, db, modulesearchNs.ID, "verifiedmodule-unpublished", "aws", true)
	createVersion(t, db, provider12.ID, "1.0.0", nil, nil, "", "")

	// ===== searchbynamespace namespace =====
	searchbyNs := CreateNamespace(t, db, "searchbynamespace", nil)

	// searchbymodulename1/searchbyprovideraws (verified)
	provider13 := CreateModuleProviderWithVerified(t, db, searchbyNs.ID, "searchbymodulename1", "searchbyprovideraws", true)
	createVersion(t, db, provider13.ID, "1.2.3", &published, nil, "", "")

	// searchbymodulename1/searchbyprovidergcp
	provider14 := CreateModuleProvider(t, db, searchbyNs.ID, "searchbymodulename1", "searchbyprovidergcp")
	createVersion(t, db, provider14.ID, "2.0.0", &published, nil, "", "")

	// searchbymodulename2/notpublished
	provider15 := CreateModuleProvider(t, db, searchbyNs.ID, "searchbymodulename2", "notpublished")
	createVersion(t, db, provider15.ID, "1.2.3", nil, nil, "", "")

	// searchbymodulename2/published
	provider16 := CreateModuleProvider(t, db, searchbyNs.ID, "searchbymodulename2", "published")
	createVersion(t, db, provider16.ID, "3.1.6", &published, nil, "", "")

	// ===== searchbynamesp-similar namespace =====
	searchbySimilarNs := CreateNamespace(t, db, "searchbynamesp-similar", nil)

	// searchbymodulename3/searchbyprovideraws (verified)
	provider17 := CreateModuleProviderWithVerified(t, db, searchbySimilarNs.ID, "searchbymodulename3", "searchbyprovideraws", true)
	createVersion(t, db, provider17.ID, "4.4.1", &published, nil, "", "")

	// searchbymodulename4/aws
	provider18 := CreateModuleProvider(t, db, searchbySimilarNs.ID, "searchbymodulename4", "aws")
	createVersion(t, db, provider18.ID, "5.5.5", &published, nil, "", "")

	// ===== testnamespace (from unit tests) =====
	testNs := CreateNamespace(t, db, "testnamespace", nil)

	// testmodulename/testprovider
	provider19 := CreateModuleProvider(t, db, testNs.ID, "testmodulename", "testprovider")
	createVersion(t, db, provider19.ID, "2.4.1", &published, nil, "", "")
	createVersion(t, db, provider19.ID, "1.0.0", &published, nil, "", "")

	// lonelymodule/testprovider
	provider20 := CreateModuleProviderWithVerified(t, db, testNs.ID, "lonelymodule", "testprovider", true)
	createVersion(t, db, provider20.ID, "1.0.0", &published, nil, "", "")

	// mock-module/testprovider
	provider21 := CreateModuleProviderWithVerified(t, db, testNs.ID, "mock-module", "testprovider", true)
	createVersion(t, db, provider21.ID, "1.2.3", &published, nil, "", "")

	// unverifiedmodule/testprovider
	provider22 := CreateModuleProvider(t, db, testNs.ID, "unverifiedmodule", "testprovider")
	createVersion(t, db, provider22.ID, "1.2.3", &published, nil, "", "")

	// internalmodule/testprovider
	provider23 := CreateModuleProvider(t, db, testNs.ID, "internalmodule", "testprovider")
	createVersion(t, db, provider23.ID, "5.2.0", &published, nil, "", "")

	// modulenorepourl/testprovider
	provider24 := CreateModuleProvider(t, db, testNs.ID, "modulenorepourl", "testprovider")
	createVersion(t, db, provider24.ID, "2.2.4", &published, nil, "", "")

	// onlybeta/testprovider
	provider25 := CreateModuleProvider(t, db, testNs.ID, "onlybeta", "testprovider")
	createVersion(t, db, provider25.ID, "2.2.4-beta", &published, &beta, "", "")

	// modulewithrepourl/testprovider
	provider26 := CreateModuleProvider(t, db, testNs.ID, "modulewithrepourl", "testprovider")
	createVersion(t, db, provider26.ID, "2.1.0", nil, nil, "", "")

	// modulenotpublished/testprovider
	provider27 := CreateModuleProvider(t, db, testNs.ID, "modulenotpublished", "testprovider")
	createVersion(t, db, provider27.ID, "10.2.1", nil, nil, "", "")

	// ===== real_providers namespace =====
	realNs := CreateNamespace(t, db, "real_providers", nil)

	// test-module/aws
	provider28 := CreateModuleProvider(t, db, realNs.ID, "test-module", "aws")
	createVersion(t, db, provider28.ID, "1.0.0", nil, nil, "", "")

	// test-module/gcp
	provider29 := CreateModuleProvider(t, db, realNs.ID, "test-module", "gcp")
	createVersion(t, db, provider29.ID, "1.0.0", nil, nil, "", "")

	// test-module/null
	provider30 := CreateModuleProvider(t, db, realNs.ID, "test-module", "null")
	createVersion(t, db, provider30.ID, "1.0.0", nil, nil, "", "")

	// test-module/doesnotexist
	provider31 := CreateModuleProvider(t, db, realNs.ID, "test-module", "doesnotexist")
	createVersion(t, db, provider31.ID, "1.0.0", nil, nil, "", "")

	// ===== genericmodules namespace =====
	genericNs := CreateNamespace(t, db, "genericmodules", nil)

	// modulename/providername
	provider32 := CreateModuleProvider(t, db, genericNs.ID, "modulename", "providername")
	createVersion(t, db, provider32.ID, "1.2.3", &published, nil, "", "")

	// ===== Additional search test modules for comprehensive testing =====
	searchNs1 := CreateNamespace(t, db, "searchnamespace1", nil)

	// Various modules for multi-term and description search testing
	provider33 := CreateModuleProvider(t, db, searchNs1.ID, "aws-vpc-module", "aws")
	createVersion(t, db, provider33.ID, "1.0.0", &published, nil, "terraform-aws-modules", "VPC module for AWS infrastructure")

	provider34 := CreateModuleProvider(t, db, searchNs1.ID, "vpc-module", "gcp")
	createVersion(t, db, provider34.ID, "1.0.0", &published, nil, "", "")

	provider35 := CreateModuleProvider(t, db, searchNs1.ID, "aws-module", "azure")
	createVersion(t, db, provider35.ID, "1.0.0", &published, nil, "", "")

	provider36 := CreateModuleProvider(t, db, searchNs1.ID, "networking-firewall", "aws")
	createVersion(t, db, provider36.ID, "1.2.0", &published, nil, "", "Firewall module for VPC networking")

	provider37 := CreateModuleProvider(t, db, searchNs1.ID, "compute-instance", "aws")
	createVersion(t, db, provider37.ID, "2.0.0", &published, nil, "", "EC2 instance management module")

	// Module with internal versions
	provider38 := CreateModuleProvider(t, db, searchNs1.ID, "internal-test", "aws")
	createVersion(t, db, provider38.ID, "1.0.0", &published, nil, "", "")
	createVersion(t, db, provider38.ID, "2.0.0-internal", nil, &internal, "", "Internal development version")

	// Module for description and owner search
	provider39 := CreateModuleProvider(t, db, searchNs1.ID, "custom-auth-module", "aws")
	createVersion(t, db, provider39.ID, "3.0.0", &published, nil, "CustomAuthTeam", "Custom authentication provider module")

	// ===== Additional namespace for pagination testing =====
	largeNs := CreateNamespace(t, db, "large-search-ns", nil)

	for i := 1; i <= 15; i++ {
		provider := CreateModuleProvider(t, db, largeNs.ID, fmt.Sprintf("search-module-%d", i), fmt.Sprintf("provider-%d", i))
		createVersion(t, db, provider.ID, fmt.Sprintf("1.%d.0", i), &published, nil, "", "")
	}

	// ===== modulesearch-contributed namespace (from Python test_data.py) =====
	// This namespace tests "contributed" search results (not verified, not in trusted namespaces)
	modulesearchContributedNs := CreateNamespace(t, db, "modulesearch-contributed", nil)

	// mixedsearch-result (published, single version)
	provider40 := CreateModuleProvider(t, db, modulesearchContributedNs.ID, "mixedsearch-result", "aws")
	createVersion(t, db, provider40.ID, "1.0.0", &published, nil, "", "")

	// mixedsearch-result-multiversion (published, multiple versions - IMPORTANT for duplicate bug testing)
	provider41 := CreateModuleProvider(t, db, modulesearchContributedNs.ID, "mixedsearch-result-multiversion", "aws")
	createVersion(t, db, provider41.ID, "1.2.3", &published, nil, "", "")
	createVersion(t, db, provider41.ID, "2.0.0", &published, nil, "", "")

	// mixedsearch-result-unpublished (unpublished)
	provider42 := CreateModuleProvider(t, db, modulesearchContributedNs.ID, "mixedsearch-result-unpublished", "aws")
	createVersion(t, db, provider42.ID, "1.2.3", nil, nil, "", "")
	createVersion(t, db, provider42.ID, "2.0.0", nil, nil, "", "")

	// ===== modulesearch-trusted namespace (from Python test_data.py) =====
	// This namespace tests "trusted" search results (namespaces configured as trusted)
	// Note: The actual "trusted" status is configured via TRUSTED_NAMESPACES in config
	modulesearchTrustedNs := CreateNamespace(t, db, "modulesearch-trusted", nil)

	// mixedsearch-trusted-result (published, single version)
	provider43 := CreateModuleProvider(t, db, modulesearchTrustedNs.ID, "mixedsearch-trusted-result", "aws")
	createVersion(t, db, provider43.ID, "1.0.0", &published, nil, "", "")

	// mixedsearch-trusted-second-result (published, single version)
	provider44 := CreateModuleProvider(t, db, modulesearchTrustedNs.ID, "mixedsearch-trusted-second-result", "datadog")
	createVersion(t, db, provider44.ID, "5.2.1", &published, nil, "", "")

	// mixedsearch-trusted-result-multiversion (published, multiple versions - IMPORTANT for duplicate bug testing)
	provider45 := CreateModuleProvider(t, db, modulesearchTrustedNs.ID, "mixedsearch-trusted-result-multiversion", "null")
	createVersion(t, db, provider45.ID, "1.2.3", &published, nil, "", "")
	createVersion(t, db, provider45.ID, "2.0.0", &published, nil, "", "")

	// mixedsearch-trusted-result-unpublished (unpublished)
	provider46 := CreateModuleProvider(t, db, modulesearchTrustedNs.ID, "mixedsearch-trusted-result-unpublished", "aws")
	createVersion(t, db, provider46.ID, "1.2.3", nil, nil, "", "")
	createVersion(t, db, provider46.ID, "2.0.0", nil, nil, "", "")

	// mixedsearch-trusted-result-verified (published, verified)
	// Python reference: /app/test/selenium/test_data.py - mixedsearch-trusted-result-verified
	provider47 := CreateModuleProviderWithVerified(t, db, modulesearchTrustedNs.ID, "mixedsearch-trusted-result-verified", "gcp", true)
	createVersion(t, db, provider47.ID, "2.0.0", &published, nil, "", "")

	// ===== Additional testnamespace modules from Python unit tests =====
	// These are important for testing edge cases like wrong version order, no versions, etc.

	// wrongversionorder/testprovider - tests version sorting with various version formats
	provider48 := CreateModuleProvider(t, db, testNs.ID, "wrongversionorder", "testprovider")
	createVersion(t, db, provider48.ID, "1.5.4", &published, nil, "", "")
	createVersion(t, db, provider48.ID, "2.1.0", &published, nil, "", "")
	createVersion(t, db, provider48.ID, "0.1.1", &published, nil, "", "")
	createVersion(t, db, provider48.ID, "10.23.0", &published, nil, "", "")
	createVersion(t, db, provider48.ID, "0.1.10", &published, nil, "", "")
	createVersion(t, db, provider48.ID, "0.0.9", &published, nil, "", "")
	createVersion(t, db, provider48.ID, "0.1.09", &published, nil, "", "")
	createVersion(t, db, provider48.ID, "0.1.8", &published, nil, "", "")
	createVersion(t, db, provider48.ID, "23.2.3-beta", &published, &beta, "", "")
	// Unpublished version
	createVersion(t, db, provider48.ID, "5.21.2", nil, nil, "", "")

	// noversions/testprovider - module with no versions
	_ = CreateModuleProvider(t, db, testNs.ID, "noversions", "testprovider")
	// No versions created - intentionally unused to test modules without versions

	// onlyunpublished/testprovider - module with only unpublished versions
	provider49 := CreateModuleProvider(t, db, testNs.ID, "onlyunpublished", "testprovider")
	createVersion(t, db, provider49.ID, "0.1.8", nil, nil, "", "")

	// onlybeta/testprovider - module with only beta versions
	provider50 := CreateModuleProvider(t, db, testNs.ID, "onlybeta", "testprovider")
	createVersion(t, db, provider50.ID, "2.5.0-beta", &published, &beta, "", "")
}

// createVersion is a helper to create a module version with common attributes
// It automatically sets the version as the latest version on the module provider
func createVersion(t *testing.T, db *sqldb.Database, moduleProviderID int, version string,
	published, beta *bool, owner, description string) {
	t.Helper()

	moduleVersion := sqldb.ModuleVersionDB{
		ModuleProviderID: moduleProviderID,
		Version:          version,
		Beta:             false,
		Internal:         false,
		Published:        published,
	}

	if owner != "" {
		moduleVersion.Owner = &owner
	}
	if description != "" {
		moduleVersion.Description = &description
	}
	if beta != nil {
		moduleVersion.Beta = *beta
	}

	err := db.DB.Create(&moduleVersion).Error
	require.NoError(t, err)

	// Set this version as the latest version for the module provider
	// This is required for the search query to find the module
	err = db.DB.Model(&sqldb.ModuleProviderDB{}).
		Where("id = ?", moduleProviderID).
		Update("latest_version_id", moduleVersion.ID).Error
	require.NoError(t, err)
}

// CreateModuleProviderWithVerified creates a module provider with specified verified status
func CreateModuleProviderWithVerified(t *testing.T, db *sqldb.Database, namespaceID int, moduleName, providerName string, verified bool) sqldb.ModuleProviderDB {
	t.Helper()

	moduleProvider := sqldb.ModuleProviderDB{
		NamespaceID: namespaceID,
		Module:      moduleName,
		Provider:    providerName,
		Verified:    &verified,
	}

	err := db.DB.Create(&moduleProvider).Error
	require.NoError(t, err)

	return moduleProvider
}

// SetupComprehensiveProviderSearchTestData creates comprehensive provider search test data
// matching Python's integration_test_data.py provider search data.
// This creates providers in providersearch-trusted and contributed-providersearch namespaces.
func SetupComprehensiveProviderSearchTestData(t *testing.T, db *sqldb.Database) {
	t.Helper()

	// Create provider categories (matching Python's integration_provider_categories)
	createProviderCategory(t, db, "Visible Monitoring", "visible-monitoring", true)
	createProviderCategory(t, db, "Second Visible Cloud", "second-visible-cloud", true)

	// Create providersearch-trusted namespace (for trusted providers)
	// Python reference: /app/test/selenium/test_data.py - 'providersearch-trusted'
	providersearchTrustedNs := CreateNamespace(t, db, "providersearch-trusted", nil)

	// Create contributed-providersearch namespace (for contributed providers)
	// Python reference: /app/test/selenium/test_data.py - 'contributed-providersearch'
	contributedProvidersearchNs := CreateNamespace(t, db, "contributed-providersearch", nil)

	// Create GPG keys directly in namespaces (not linked to providers yet)
	gpgKeyProviderSearchTrusted := createGPGKeyInNamespace(t, db, providersearchTrustedNs.ID, "D8A89D97BB7526F33C8A2D8C39C57A3D0D24B532")
	gpgKeyContributed := createGPGKeyInNamespace(t, db, contributedProvidersearchNs.ID, "D7AA1BEFF16FA788760E54F5591EF84DC5EDCD68")

	// Get category IDs
	visibleMonitoringCat := getProviderCategoryBySlug(t, db, "visible-monitoring")

	// ===== providersearch-trusted namespace providers =====
	// Python reference: /app/test/selenium/test_data.py - providersearch-trusted providers

	// mixedsearch-trusted-result (one version)
	// Python: terraform-provider-mixedsearch-trusted-result
	provider1 := CreateProvider(t, db, providersearchTrustedNs.ID, "mixedsearch-trusted-result",
		stringPtr("Test Multiple Versions"), sqldb.ProviderTierCommunity, &visibleMonitoringCat)
	createProviderVersion(t, db, provider1.ID, "1.0.0", gpgKeyProviderSearchTrusted.ID, false)

	// mixedsearch-trusted-second-result (one version)
	// Python: terraform-provider-mixedsearch-trusted-second-result
	provider2 := CreateProvider(t, db, providersearchTrustedNs.ID, "mixedsearch-trusted-second-result",
		stringPtr("Test Multiple Versions"), sqldb.ProviderTierCommunity, &visibleMonitoringCat)
	createProviderVersion(t, db, provider2.ID, "5.2.1", gpgKeyProviderSearchTrusted.ID, false)

	// mixedsearch-trusted-result-multiversion (multiple versions)
	// Python: terraform-provider-mixedsearch-trusted-result-multiversion
	provider3 := CreateProvider(t, db, providersearchTrustedNs.ID, "mixedsearch-trusted-result-multiversion",
		stringPtr("Test Multiple Versions"), sqldb.ProviderTierCommunity, &visibleMonitoringCat)
	createProviderVersion(t, db, provider3.ID, "1.2.3", gpgKeyProviderSearchTrusted.ID, false)
	createProviderVersion(t, db, provider3.ID, "2.0.0", gpgKeyProviderSearchTrusted.ID, false)

	// ===== contributed-providersearch namespace providers =====
	// Python reference: /app/test/selenium/test_data.py - contributed-providersearch providers

	// mixedsearch-result (one version)
	// Python: terraform-provider-mixedsearch-result
	provider4 := CreateProvider(t, db, contributedProvidersearchNs.ID, "mixedsearch-result",
		stringPtr("Test Multiple Versions"), sqldb.ProviderTierCommunity, &visibleMonitoringCat)
	createProviderVersion(t, db, provider4.ID, "1.0.0", gpgKeyContributed.ID, false)

	// mixedsearch-result-multiversion (multiple versions - IMPORTANT for duplicate bug testing)
	// Python: terraform-provider-mixedsearch-result-multiversion
	provider5 := CreateProvider(t, db, contributedProvidersearchNs.ID, "mixedsearch-result-multiversion",
		stringPtr("Test Multiple Versions"), sqldb.ProviderTierCommunity, &visibleMonitoringCat)
	createProviderVersion(t, db, provider5.ID, "1.2.3", gpgKeyContributed.ID, false)
	createProviderVersion(t, db, provider5.ID, "2.0.0", gpgKeyContributed.ID, false)

	// mixedsearch-result-no-version (no versions - should be excluded from search)
	// Python: terraform-provider-mixedsearch-result-no-version
	_ = CreateProvider(t, db, contributedProvidersearchNs.ID, "mixedsearch-result-no-version",
		stringPtr("DESCRIPTION-NoVersion"), sqldb.ProviderTierCommunity, &visibleMonitoringCat)
}

// createGPGKeyInNamespace creates a GPG key directly in a namespace (not linked to a provider)
func createGPGKeyInNamespace(t *testing.T, db *sqldb.Database, namespaceID int, fingerprint string) sqldb.GPGKeyDB {
	t.Helper()

	asciiArmor := []byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nTest ASCII armor for " + fingerprint + "\n-----END PGP PUBLIC KEY BLOCK-----")
	source := "test-source"
	keyID := &fingerprint

	gpgKey := sqldb.GPGKeyDB{
		NamespaceID: namespaceID,
		ASCIIArmor:  asciiArmor,
		KeyID:       keyID,
		Fingerprint: keyID,
		Source:      &source,
	}

	err := db.DB.Create(&gpgKey).Error
	require.NoError(t, err)

	return gpgKey
}

// CreateProviderWithGPGKey creates a provider with GPG key ID (instead of creating a new GPG key)
func CreateProviderWithGPGKey(t *testing.T, db *sqldb.Database, namespaceID int, name string, description *string, tier sqldb.ProviderTier, categoryID *int, gpgKeyID int) sqldb.ProviderDB {
	t.Helper()

	provider := sqldb.ProviderDB{
		NamespaceID:        namespaceID,
		Name:               name,
		Description:        description,
		Tier:               tier,
		ProviderCategoryID: categoryID,
	}

	err := db.DB.Create(&provider).Error
	require.NoError(t, err)

	return provider
}

// createProviderVersion is a helper to create a provider version with GPG key
// It automatically sets the version as the latest version on the provider
func createProviderVersion(t *testing.T, db *sqldb.Database, providerID int, version string, gpgKeyID int, beta bool) {
	t.Helper()

	gitTag := "v" + version
	publishedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	providerVersion := sqldb.ProviderVersionDB{
		ProviderID:  providerID,
		Version:     version,
		GitTag:      &gitTag,
		GPGKeyID:    gpgKeyID,
		PublishedAt: &publishedAt,
		Beta:        beta,
	}

	err := db.DB.Create(&providerVersion).Error
	require.NoError(t, err)

	// Set this version as the latest version for the provider
	err = db.DB.Model(&sqldb.ProviderDB{}).
		Where("id = ?", providerID).
		Update("latest_version_id", providerVersion.ID).Error
	require.NoError(t, err)
}

// createProviderCategory creates a provider category
func createProviderCategory(t *testing.T, db *sqldb.Database, name, slug string, userSelectable bool) sqldb.ProviderCategoryDB {
	t.Helper()

	// First, try to find existing category by slug
	var existingCategory sqldb.ProviderCategoryDB
	err := db.DB.Where("slug = ?", slug).First(&existingCategory).Error
	if err == nil {
		// Category already exists, return it
		return existingCategory
	}

	// Category doesn't exist, create it
	namePtr := &name
	category := sqldb.ProviderCategoryDB{
		Name:           namePtr,
		Slug:           slug,
		UserSelectable: userSelectable,
	}

	err = db.DB.Create(&category).Error
	require.NoError(t, err)

	return category
}

// getProviderCategoryBySlug gets a provider category by slug
func getProviderCategoryBySlug(t *testing.T, db *sqldb.Database, slug string) int {
	t.Helper()

	var category sqldb.ProviderCategoryDB
	err := db.DB.Where("slug = ?", slug).First(&category).Error
	require.NoError(t, err)
	return category.ID
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// SetupFullyPopulatedModule creates the fullypopulated module from Python's test_data.py
// Python reference: /app/test/unit/terrareg/test_data.py - moduledetails/fullypopulated
// This creates a module with all fields populated for comprehensive testing.
//
// Returns: (namespace, moduleProvider, moduleVersions)
func SetupFullyPopulatedModule(t *testing.T, db *sqldb.Database) (sqldb.NamespaceDB, sqldb.ModuleProviderDB, []sqldb.ModuleVersionDB) {
	t.Helper()

	// Create namespace
	namespace := CreateNamespace(t, db, "moduledetails", nil)

	// Create module provider with all git configuration
	moduleProvider := CreateModuleProvider(t, db, namespace.ID, "fullypopulated", "testprovider")

	// Update module provider with all git configuration
	repoBaseURL := "https://mp-base-url.com/{namespace}/{module}-{provider}"
	repoBrowseURL := "https://mp-browse-url.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix"
	repoCloneURL := "ssh://mp-clone-url.com/{namespace}/{module}-{provider}"

	err := db.DB.Model(&moduleProvider).Updates(map[string]interface{}{
		"repo_base_url_template":   &repoBaseURL,
		"repo_browse_url_template": &repoBrowseURL,
		"repo_clone_url_template":  &repoCloneURL,
	}).Error
	require.NoError(t, err)

	published := true
	beta := true
	internal := false
	publishedAt := time.Date(2022, 1, 5, 22, 53, 12, 0, time.UTC)

	// Create all versions from Python test_data.py
	versions := []struct {
		version     string
		published   *bool
		beta        *bool
		internal    *bool
		owner       *string
		description *string
		repoBaseURL *string
		readme      *string
		varTemplate *string
	}{
		// Older version
		{"1.2.0", &published, nil, nil, nil, nil, nil, nil, nil},
		// Newer unpublished version
		{"1.6.0", nil, nil, nil, nil, nil, nil, nil, nil},
		// Newer published beta version
		{"1.6.1-beta", &published, &beta, nil, nil, nil, nil, nil, nil},
		// Unpublished and beta version
		{"1.0.0-beta", nil, &beta, nil, nil, nil, nil, nil, nil},
		// Main fully populated version
		{
			"1.5.0",
			&published,
			nil,
			&internal,
			stringPtr("This is the owner of the module"),
			stringPtr("This is a test module version for tests."),
			stringPtr("https://link-to.com/source-code-here"),
			stringPtr("# This is an exaple README!"),
			stringPtr(`[
				{"name": "name_of_application", "type": "text", "quote_value": true, "additional_help": "Provide the name of the application"},
				{"name": "variable_template_with_markdown", "type": "text", "quote_value": true, "additional_help": "This **is** some _markdown_"},
				{"name": "variable_template_with_html", "type": "text", "quote_value": true, "additional_help": "This <b>is</b> some <i>html</i>"}
			]`),
		},
	}

	createdVersions := make([]sqldb.ModuleVersionDB, 0, len(versions))
	for _, v := range versions {
		moduleVersion := sqldb.ModuleVersionDB{
			ModuleProviderID: moduleProvider.ID,
			Version:          v.version,
			Beta:             false,
			Internal:         false,
			Published:        v.published,
		}

		if v.beta != nil {
			moduleVersion.Beta = *v.beta
		}
		if v.internal != nil {
			moduleVersion.Internal = *v.internal
		}
		if v.owner != nil {
			moduleVersion.Owner = v.owner
		}
		if v.description != nil {
			moduleVersion.Description = v.description
		}
		if v.repoBaseURL != nil {
			moduleVersion.RepoBaseURLTemplate = v.repoBaseURL
		}

		// Set published_at for main version
		if v.version == "1.5.0" {
			moduleVersion.PublishedAt = &publishedAt
		}

		// Create module details if readme or variable template is provided
		if v.readme != nil || v.varTemplate != nil {
			variableTemplate := []byte{}
			if v.varTemplate != nil {
				variableTemplate = []byte(*v.varTemplate)
			}

			readmeContent := []byte{}
			if v.readme != nil {
				readmeContent = []byte(*v.readme)
			}

			moduleDetails := sqldb.ModuleDetailsDB{
				ReadmeContent:    readmeContent,
				TerraformDocs:    []byte("{}"),
				Tfsec:            []byte("{}"),
				Infracost:        []byte("{}"),
				TerraformGraph:   []byte("{}"),
				TerraformModules: []byte("{}"),
				TerraformVersion: []byte("1.0.0"),
			}

			// Store variable template in module version instead
			if v.varTemplate != nil {
				moduleVersion.VariableTemplate = variableTemplate
			}

			err := db.DB.Create(&moduleDetails).Error
			require.NoError(t, err)
			moduleVersion.ModuleDetailsID = &moduleDetails.ID
		}

		err = db.DB.Create(&moduleVersion).Error
		require.NoError(t, err)
		createdVersions = append(createdVersions, moduleVersion)

		// Set version 1.5.0 as the latest version
		if v.version == "1.5.0" {
			err = db.DB.Model(&moduleProvider).Update("latest_version_id", moduleVersion.ID).Error
			require.NoError(t, err)
		}
	}

	return namespace, moduleProvider, createdVersions
}

// SetupTestNamespaceFromPython creates all test modules from Python's testnamespace
// Python reference: /app/test/unit/terrareg/test_data.py - testnamespace
// This includes: testmodulename, lonelymodule, mock-module, unverifiedmodule, internalmodule, etc.
func SetupTestNamespaceFromPython(t *testing.T, db *sqldb.Database) sqldb.NamespaceDB {
	t.Helper()

	namespace := CreateNamespace(t, db, "testnamespace", nil)

	published := true
	beta := true
	internal := true

	// testmodulename/testprovider (ID: 1, Latest: 2.4.1, Verified: true)
	provider1 := CreateModuleProviderWithVerified(t, db, namespace.ID, "testmodulename", "testprovider", true)
	createVersion(t, db, provider1.ID, "2.4.1", &published, nil, "", "")
	createVersion(t, db, provider1.ID, "1.0.0", &published, nil, "", "")

	// lonelymodule/testprovider (ID: 2, Latest: 1.0.0, Verified: true)
	provider2 := CreateModuleProviderWithVerified(t, db, namespace.ID, "lonelymodule", "testprovider", true)
	createVersion(t, db, provider2.ID, "1.0.0", &published, nil, "", "")

	// mock-module/testprovider (ID: 3, Verified: true, Latest: 1.2.3)
	provider3 := CreateModuleProviderWithVerified(t, db, namespace.ID, "mock-module", "testprovider", true)
	createVersion(t, db, provider3.ID, "1.2.3", &published, nil, "", "")

	// unverifiedmodule/testprovider (ID: 16, Verified: false, Latest: 1.2.3)
	provider4 := CreateModuleProviderWithVerified(t, db, namespace.ID, "unverifiedmodule", "testprovider", false)
	createVersion(t, db, provider4.ID, "1.2.3", &published, nil, "", "")

	// internalmodule/testprovider (ID: 17, Verified: false, Latest: 5.2.0, Internal: true)
	provider5 := CreateModuleProviderWithVerified(t, db, namespace.ID, "internalmodule", "testprovider", false)
	createVersionWithInternal(t, db, provider5.ID, "5.2.0", &published, &internal, "", "")

	// modulenorepourl/testprovider (ID: 5, Latest: 2.2.4)
	provider6 := CreateModuleProvider(t, db, namespace.ID, "modulenorepourl", "testprovider")
	createVersion(t, db, provider6.ID, "2.2.4", &published, nil, "", "")

	// onlybeta/testprovider (ID: 18, Latest: 2.2.4-beta, Beta: true)
	provider7 := CreateModuleProvider(t, db, namespace.ID, "onlybeta", "testprovider")
	createVersion(t, db, provider7.ID, "2.2.4-beta", &published, &beta, "", "")

	// modulewithrepourl/testprovider (ID: 6, Latest: 2.1.0, has repo_clone_url_template)
	provider8 := CreateModuleProvider(t, db, namespace.ID, "modulewithrepourl", "testprovider")
	repoCloneURL := "https://github.com/test/test.git"
	err := db.DB.Model(&provider8).Update("repo_clone_url_template", &repoCloneURL).Error
	require.NoError(t, err)
	createVersion(t, db, provider8.ID, "2.1.0", nil, nil, "", "")

	// modulenotpublished/testprovider (ID: 15, Latest: None, all versions unpublished)
	// Also has git configuration templates
	provider9 := CreateModuleProvider(t, db, namespace.ID, "modulenotpublished", "testprovider")
	repoBase := "https://custom-localhost.com/{namespace}/{module}-{provider}"
	repoBrowse := "https://custom-localhost.com/{namespace}/{module}-{provider}/browse/{tag}/{path}"
	repoClone := "ssh://custom-localhost.com/{namespace}/{module}-{provider}"
	err = db.DB.Model(&provider9).Updates(map[string]interface{}{
		"repo_base_url_template":   &repoBase,
		"repo_browse_url_template": &repoBrowse,
		"repo_clone_url_template":  &repoClone,
	}).Error
	require.NoError(t, err)
	createVersion(t, db, provider9.ID, "10.2.1", nil, nil, "", "")

	// withsecurityissues/testprovider (ID: 20, Latest: 1.0.0, has tfsec data)
	provider10 := CreateModuleProvider(t, db, namespace.ID, "withsecurityissues", "testprovider")
	createVersionWithTfsec(t, db, provider10.ID, "1.0.0", &published, nil, "", "")

	// wrongversionorder/testprovider - tests version sorting
	provider11 := CreateModuleProvider(t, db, namespace.ID, "wrongversionorder", "testprovider")
	createVersion(t, db, provider11.ID, "1.5.4", &published, nil, "", "")
	createVersion(t, db, provider11.ID, "2.1.0", &published, nil, "", "")
	createVersion(t, db, provider11.ID, "0.1.1", &published, nil, "", "")
	createVersion(t, db, provider11.ID, "10.23.0", &published, nil, "", "")
	createVersion(t, db, provider11.ID, "0.1.10", &published, nil, "", "")
	createVersion(t, db, provider11.ID, "0.0.9", &published, nil, "", "")
	createVersion(t, db, provider11.ID, "0.1.09", &published, nil, "", "")
	createVersion(t, db, provider11.ID, "0.1.8", &published, nil, "", "")
	createVersion(t, db, provider11.ID, "23.2.3-beta", &published, &beta, "", "")
	createVersion(t, db, provider11.ID, "5.21.2", nil, nil, "", "") // unpublished

	// noversions/testprovider - module with no versions
	_ = CreateModuleProvider(t, db, namespace.ID, "noversions", "testprovider")

	// onlyunpublished/testprovider - module with only unpublished versions
	provider13 := CreateModuleProvider(t, db, namespace.ID, "onlyunpublished", "testprovider")
	createVersion(t, db, provider13.ID, "0.1.8", nil, nil, "", "")

	return namespace
}

// createVersionWithInternal creates a version with internal flag
func createVersionWithInternal(t *testing.T, db *sqldb.Database, moduleProviderID int, version string,
	published, internal *bool, owner, description string) {
	t.Helper()

	moduleVersion := sqldb.ModuleVersionDB{
		ModuleProviderID: moduleProviderID,
		Version:          version,
		Beta:             false,
		Internal:         false,
		Published:        published,
	}

	if internal != nil {
		moduleVersion.Internal = *internal
	}
	if owner != "" {
		moduleVersion.Owner = &owner
	}
	if description != "" {
		moduleVersion.Description = &description
	}

	err := db.DB.Create(&moduleVersion).Error
	require.NoError(t, err)

	// Set this version as the latest version for the module provider
	err = db.DB.Model(&sqldb.ModuleProviderDB{}).
		Where("id = ?", moduleProviderID).
		Update("latest_version_id", moduleVersion.ID).Error
	require.NoError(t, err)
}

// createVersionWithTfsec creates a version with Tfsec security data
// Python reference: /app/test/unit/terrareg/test_data.py - withsecurityissues
func createVersionWithTfsec(t *testing.T, db *sqldb.Database, moduleProviderID int, version string,
	published, beta *bool, owner, description string) {
	t.Helper()

	// Create module details with Tfsec data
	// Includes all fields from tfsec JSON output
	// Python reference: /app/test/selenium/test_data.py withsecurityissues test data
	tfsecJSON := `{
		"results": [
			{
				"description": "Secret explicitly uses the default key.",
				"impact": "Using AWS managed keys reduces the flexibility and control over the encryption key",
				"links": [
					"https://aquasecurity.github.io/tfsec/v1.26.0/checks/aws/ssm/secret-use-customer-key/"
				],
				"location": {
					"end_line": 4,
					"filename": "main.tf",
					"start_line": 2
				},
				"long_id": "aws-ssm-secret-use-customer-key",
				"resolution": "Use customer managed keys",
				"resource": "aws_ssm_parameter.default_key",
				"rule_description": "SSM Parameter secrets should use customer managed keys",
				"rule_id": "AVD-AWS-0098",
				"rule_provider": "aws",
				"rule_service": "ssm",
				"severity": "LOW",
				"status": 0,
				"warning": false
			},
			{
				"description": "Some security issue 2.",
				"impact": "Entire project is compromised",
				"links": [
					"https://example.com/security-issue-2/"
				],
				"location": {
					"end_line": 10,
					"filename": "main.tf",
					"start_line": 6
				},
				"long_id": "bad-code-security-issue",
				"resolution": "Fix the security issue",
				"resource": "bad_resource.example",
				"rule_description": "This is a bad security issue",
				"rule_id": "DDG-ANC-001",
				"rule_provider": "bad",
				"rule_service": "code",
				"severity": "HIGH",
				"status": 0,
				"warning": false
			}
		]
	}`

	moduleDetails := sqldb.ModuleDetailsDB{
		ReadmeContent:    []byte{},
		TerraformDocs:    []byte("{}"),
		Tfsec:            []byte(tfsecJSON),
		Infracost:        []byte("{}"),
		TerraformGraph:   []byte("{}"),
		TerraformModules: []byte("{}"),
		TerraformVersion: []byte("1.0.0"),
	}

	err := db.DB.Create(&moduleDetails).Error
	require.NoError(t, err)

	moduleVersion := sqldb.ModuleVersionDB{
		ModuleProviderID: moduleProviderID,
		Version:          version,
		Beta:             false,
		Internal:         false,
		Published:        published,
		ModuleDetailsID:  &moduleDetails.ID,
	}

	if beta != nil {
		moduleVersion.Beta = *beta
	}
	if owner != "" {
		moduleVersion.Owner = &owner
	}
	if description != "" {
		moduleVersion.Description = &description
	}

	err = db.DB.Create(&moduleVersion).Error
	require.NoError(t, err)

	// Set this version as the latest version for the module provider
	err = db.DB.Model(&sqldb.ModuleProviderDB{}).
		Where("id = ?", moduleProviderID).
		Update("latest_version_id", moduleVersion.ID).Error
	require.NoError(t, err)
}

// CreateModuleVersionWithSecurityIssues creates a module version with tfsec security data.
// Returns the created module version for use in tests.
func CreateModuleVersionWithSecurityIssues(t *testing.T, db *sqldb.Database, moduleProviderID int, version string, published *bool) sqldb.ModuleVersionDB {
	t.Helper()
	createVersionWithTfsec(t, db, moduleProviderID, version, published, nil, "", "")

	// Find and return the created module version
	var moduleVersion sqldb.ModuleVersionDB
	err := db.DB.Where("module_provider_id = ? AND version = ?", moduleProviderID, version).First(&moduleVersion).Error
	require.NoError(t, err, "Failed to find created module version")
	return moduleVersion
}
