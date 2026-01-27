//go:build selenium

package selenium

import (
	"context"
	"fmt"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	integrationTestUtils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// TestCreateModuleProvider tests the create module provider page.
// Python reference: /app/test/selenium/test_create_module_provider.py - TestCreateModuleProvider class
func TestCreateModuleProvider(t *testing.T) {
	t.Run("test_page_details", testCreateModuleProviderPageDetails)
	t.Run("test_create_basic", testCreateModuleProviderBasic)
	t.Run("test_create_against_namespace_with_display_name", testCreateModuleProviderWithDisplayName)
	t.Run("test_with_git_path", testCreateModuleProviderWithGitPath)
	t.Run("test_with_git_tag_format_none", func(t *testing.T) {
		testCreateModuleProviderGitTagFormat(t, "", false, false, "v{version}")
	})
	t.Run("test_with_git_tag_format_empty", func(t *testing.T) {
		testCreateModuleProviderGitTagFormat(t, "", true, false, "{version}")
	})
	t.Run("test_with_git_tag_format_invalid", func(t *testing.T) {
		testCreateModuleProviderGitTagFormat(t, "testgittag", false, true, "")
	})
	t.Run("test_with_git_tag_format_custom", func(t *testing.T) {
		testCreateModuleProviderGitTagFormat(t, "unittestvalue{version}", false, false, "unittestvalue{version}")
	})
	t.Run("test_with_git_tag_format_major", func(t *testing.T) {
		testCreateModuleProviderGitTagFormat(t, "releases/v{major}", false, false, "releases/v{major}")
	})
	t.Run("test_with_git_tag_format_minor", func(t *testing.T) {
		testCreateModuleProviderGitTagFormat(t, "releases/v{minor}", false, false, "releases/v{minor}")
	})
	t.Run("test_with_git_tag_format_patch", func(t *testing.T) {
		testCreateModuleProviderGitTagFormat(t, "releases/v{patch}", false, false, "releases/v{patch}")
	})
	t.Run("test_unauthenticated", testCreateModuleProviderUnauthenticated)
	t.Run("test_duplicate_module", testCreateModuleProviderDuplicate)
	t.Run("test_creating_with_invalid_git_tag", testCreateModuleProviderInvalidGitTag)
	t.Run("test_creating_with_invalid_module_name", testCreateModuleProviderInvalidModuleName)
	t.Run("test_creating_with_invalid_provider", testCreateModuleProviderInvalidProvider)
}

// newCreateModuleProviderTest creates a new SeleniumTest configured for module provider tests.
func newCreateModuleProviderTest(t *testing.T) *SeleniumTest {
	st := NewSeleniumTestWithConfig(t, ConfigForCreateModuleProviderTests())

	// Setup test data - create namespaces that exist in Python's integration_test_data
	// Python reference: /app/test/selenium/test_data.py - integration_test_data
	db := st.server.GetDB()
	_ = integrationTestUtils.CreateNamespace(t, db, "testmodulecreation", nil)
	_ = integrationTestUtils.CreateNamespace(t, db, "moduledetails", nil)

	// Create namespace with display name (for test_create_against_namespace_with_display_name)
	// Python: 'withdisplayname': {'display_name': 'A Display Name'}
	displayName := "A Display Name"
	namespaceWithDisplayName := sqldb.NamespaceDB{
		Namespace:     "withdisplayname",
		DisplayName:   &displayName,
		NamespaceType: sqldb.NamespaceTypeNone,
	}
	require.NoError(t, db.DB.Create(&namespaceWithDisplayName).Error)

	return st
}

// testCreateModuleProviderPageDetails tests that the page contains required information.
// Python reference: /app/test/selenium/test_create_module_provider.py - TestCreateModuleProvider.test_page_details
func testCreateModuleProviderPageDetails(t *testing.T) {
	st := newCreateModuleProviderTest(t)
	defer st.TearDown()

	performAdminAuthentication(st, "test-admin-token")

	st.NavigateTo("/create-module")

	// Python: assert self.selenium_instance.find_element(By.CLASS_NAME, 'breadcrumb').text == 'Create Module'
	st.AssertTextContent(".breadcrumb", "Create Module")

	// Verify form exists
	_ = st.WaitForElement("#create-module-form")

	// Verify namespace dropdown has expected default value
	// Python: assert input.text == 'Custom' (for git provider dropdown)
	// Python: assert input.get_attribute('value') == 'v{version}' (for git tag format)
	// Note: For input fields, we need to check the value attribute, not text content
	st.AssertAttributeValue("#create-module-git-tag-format", "value", "v{version}")
}

// testCreateModuleProviderBasic tests creating module provider with basic inputs.
// Python reference: /app/test/selenium/test_create_module_provider.py - TestCreateModuleProvider.test_create_basic
func testCreateModuleProviderBasic(t *testing.T) {
	st := newCreateModuleProviderTest(t)
	defer st.TearDown()

	performAdminAuthentication(st, "test-admin-token")

	st.NavigateTo("/create-module")

	// Python: Select(self.selenium_instance.find_element(By.ID, 'create-module-namespace')).select_by_visible_text('testmodulecreation')
	st.SelectOptionByVisibleText("#create-module-namespace", "testmodulecreation")

	// Python: self._fill_out_field_by_label('Module Name', 'minimal-module')
	fillOutModuleFieldByLabel(st, "Module Name", "minimal-module")

	// Python: self._fill_out_field_by_label('Provider', 'testprovider')
	fillOutModuleFieldByLabel(st, "Provider", "testprovider")

	// Python: self._fill_out_field_by_label('Git tag format', 'vunit{version}test')
	fillOutModuleFieldByLabel(st, "Git tag format", "vunit{version}test")

	// Python: self._click_create()
	clickCreateModuleButton(st)

	// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/testmodulecreation/minimal-module/testprovider'))
	st.WaitForURL("/modules/testmodulecreation/minimal-module/testprovider")
	currentURL := st.GetCurrentURL()
	assert.Equal(t, st.GetURL("/modules/testmodulecreation/minimal-module/testprovider"), currentURL)
}

// testCreateModuleProviderWithDisplayName tests creating against namespace with display name.
// Python reference: /app/test/selenium/test_create_module_provider.py - TestCreateModuleProvider.test_create_against_namespace_with_display_name
func testCreateModuleProviderWithDisplayName(t *testing.T) {
	st := newCreateModuleProviderTest(t)
	defer st.TearDown()

	performAdminAuthentication(st, "test-admin-token")

	st.NavigateTo("/create-module")

	// Python: Select(self.selenium_instance.find_element(By.ID, 'create-module-namespace')).select_by_visible_text('A Display Name')
	st.SelectOptionByVisibleText("#create-module-namespace", "A Display Name")

	fillOutModuleFieldByLabel(st, "Module Name", "minimal-module")
	fillOutModuleFieldByLabel(st, "Provider", "testprovider")
	fillOutModuleFieldByLabel(st, "Git tag format", "vunit{version}test")

	clickCreateModuleButton(st)

	// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/withdisplayname/minimal-module/testprovider'))
	st.WaitForURL("/modules/withdisplayname/minimal-module/testprovider")
	currentURL := st.GetCurrentURL()
	assert.Equal(t, st.GetURL("/modules/withdisplayname/minimal-module/testprovider"), currentURL)
}

// testCreateModuleProviderWithGitPath tests creating module provider with custom git path.
// Python reference: /app/test/selenium/test_create_module_provider.py - TestCreateModuleProvider.test_with_git_path
func testCreateModuleProviderWithGitPath(t *testing.T) {
	st := newCreateModuleProviderTest(t)
	defer st.TearDown()

	performAdminAuthentication(st, "test-admin-token")

	st.NavigateTo("/create-module")

	st.SelectOptionByVisibleText("#create-module-namespace", "testmodulecreation")
	fillOutModuleFieldByLabel(st, "Module Name", "with-git-path")
	fillOutModuleFieldByLabel(st, "Provider", "testprovider")

	// Python: self._fill_out_field_by_label('Git path', './testmodulesubdir')
	fillOutModuleFieldByLabel(st, "Git path", "./testmodulesubdir")

	clickCreateModuleButton(st)

	st.WaitForURL("/modules/testmodulecreation/with-git-path/testprovider")
	currentURL := st.GetCurrentURL()
	assert.Equal(t, st.GetURL("/modules/testmodulecreation/with-git-path/testprovider"), currentURL)
}

// gitTagFormatTestCase represents a test case for git tag format.
type gitTagFormatTestCase struct {
	gitTagFormat              string
	shouldShowValidationError bool
	shouldError               bool
	expectedGitTagFormat      string
}

// testCreateModuleProviderGitTagFormat tests git tag format validation.
// Python reference: /app/test/selenium/test_create_module_provider.py - TestCreateModuleProvider.test_with_git_tag_format
func testCreateModuleProviderGitTagFormat(t *testing.T, gitTagFormat string, shouldShowValidationError, shouldError bool, expectedGitTagFormat string) {
	st := newCreateModuleProviderTest(t)
	defer st.TearDown()

	performAdminAuthentication(st, "test-admin-token")

	st.NavigateTo("/create-module")

	st.SelectOptionByVisibleText("#create-module-namespace", "testmodulecreation")
	fillOutModuleFieldByLabel(st, "Module Name", "with-git-path")
	fillOutModuleFieldByLabel(st, "Provider", "testprovider")

	if gitTagFormat != "" {
		fillOutModuleFieldByLabel(st, "Git tag format", gitTagFormat)
	}

	clickCreateModuleButton(st)

	// Python: Check if form validation is shown
	if shouldShowValidationError {
		// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'create-module-git-tag-format').get_attribute('validationMessage'), 'Please fill out this field.')
		validationMessage := st.GetAttribute("#create-module-git-tag-format", "validationMessage")
		assert.Equal(t, "Please fill out this field.", validationMessage)
		currentURL := st.GetCurrentURL()
		assert.Equal(t, st.GetURL("/create-module"), currentURL)
	} else if shouldError {
		// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'create-error').is_displayed(), True)
		// Python: assert error.text == "Invalid git tag format. Must contain one placeholder: {version}, {major}, {minor}, {patch}."
		st.AssertElementVisible("#create-error")
		st.AssertTextContent("#create-error", "Invalid git tag format. Must contain one placeholder: {version}, {major}, {minor}, {patch}.")
		currentURL := st.GetCurrentURL()
		assert.Equal(t, st.GetURL("/create-module"), currentURL)
	} else {
		// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/testmodulecreation/with-git-path/testprovider'))
		st.WaitForURL("/modules/testmodulecreation/with-git-path/testprovider")
		currentURL := st.GetCurrentURL()
		assert.Equal(t, st.GetURL("/modules/testmodulecreation/with-git-path/testprovider"), currentURL)
	}
}

// testCreateModuleProviderUnauthenticated tests creating module when not authenticated.
// Python reference: /app/test/selenium/test_create_module_provider.py - TestCreateModuleProvider.test_unauthenticated
func testCreateModuleProviderUnauthenticated(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()

	st.DeleteCookiesAndLocalStorage()

	st.NavigateTo("/create-module")

	fillOutModuleFieldByLabel(st, "Module Name", "with-git-path")
	fillOutModuleFieldByLabel(st, "Provider", "testprovider")

	clickCreateModuleButton(st)

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'create-error').is_displayed(), True)
	// Python: assert error.text == "You must be logged in to perform this action.\nIf you were previously logged in, please re-authentication and try again.")
	st.AssertElementVisible("#create-error")
	errorText := st.WaitForElement("#create-error").Text()
	assert.Contains(t, errorText, "You must be logged in to perform this action")
	assert.Contains(t, errorText, "If you were previously logged in, please re-authentication and try again")

	currentURL := st.GetCurrentURL()
	assert.Equal(t, st.GetURL("/create-module"), currentURL)
}

// testCreateModuleProviderDuplicate tests creating a module that already exists.
// Python reference: /app/test/selenium/test_create_module_provider.py - TestCreateModuleProvider.test_duplicate_module
func testCreateModuleProviderDuplicate(t *testing.T) {
	st := newCreateModuleProviderTest(t)
	defer st.TearDown()

	performAdminAuthentication(st, "test-admin-token")

	st.NavigateTo("/create-module")

	// Python: Select(self.selenium_instance.find_element(By.ID, 'create-module-namespace')).select_by_visible_text('moduledetails')
	st.SelectOptionByVisibleText("#create-module-namespace", "moduledetails")

	fillOutModuleFieldByLabel(st, "Module Name", "fullypopulated")
	fillOutModuleFieldByLabel(st, "Provider", "testprovider")

	clickCreateModuleButton(st)

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'create-error').is_displayed(), True)
	// Python: assert error.text == "Module provider already exists"
	st.AssertElementVisible("#create-error")
	st.AssertTextContent("#create-error", "Module provider already exists")

	currentURL := st.GetCurrentURL()
	assert.Equal(t, st.GetURL("/create-module"), currentURL)
}

// testCreateModuleProviderInvalidGitTag tests creating module with invalid git tag format.
// Python reference: /app/test/selenium/test_create_module_provider.py - TestCreateModuleProvider.test_creating_with_invalid_git_tag
func testCreateModuleProviderInvalidGitTag(t *testing.T) {
	st := newCreateModuleProviderTest(t)
	defer st.TearDown()

	performAdminAuthentication(st, "test-admin-token")

	st.NavigateTo("/create-module")

	st.SelectOptionByVisibleText("#create-module-namespace", "moduledetails")
	fillOutModuleFieldByLabel(st, "Module Name", "fullypopulated")
	fillOutModuleFieldByLabel(st, "Provider", "invalidgittag")

	// Python: self._fill_out_field_by_label('Git tag format', "doesnotcontainplaceholder")
	fillOutModuleFieldByLabel(st, "Git tag format", "doesnotcontainplaceholder")

	clickCreateModuleButton(st)

	st.AssertElementVisible("#create-error")
	st.AssertTextContent("#create-error", "Invalid git tag format. Must contain one placeholder: {version}, {major}, {minor}, {patch}.")

	currentURL := st.GetCurrentURL()
	assert.Equal(t, st.GetURL("/create-module"), currentURL)
}

// testCreateModuleProviderInvalidModuleName tests creating module with invalid module name.
// Python reference: /app/test/selenium/test_create_module_provider.py - TestCreateModuleProvider.test_creating_with_invalid_module_name
func testCreateModuleProviderInvalidModuleName(t *testing.T) {
	st := newCreateModuleProviderTest(t)
	defer st.TearDown()

	performAdminAuthentication(st, "test-admin-token")

	st.NavigateTo("/create-module")

	st.SelectOptionByVisibleText("#create-module-namespace", "moduledetails")
	fillOutModuleFieldByLabel(st, "Module Name", "Invalid Module Name")
	fillOutModuleFieldByLabel(st, "Provider", "testprovider")

	clickCreateModuleButton(st)

	st.AssertElementVisible("#create-error")
	st.AssertTextContent("#create-error", "Module name is invalid")

	currentURL := st.GetCurrentURL()
	assert.Equal(t, st.GetURL("/create-module"), currentURL)
}

// testCreateModuleProviderInvalidProvider tests creating module with invalid provider name.
// Python reference: /app/test/selenium/test_create_module_provider.py - TestCreateModuleProvider.test_creating_with_invalid_provider
func testCreateModuleProviderInvalidProvider(t *testing.T) {
	st := newCreateModuleProviderTest(t)
	defer st.TearDown()

	performAdminAuthentication(st, "test-admin-token")

	st.NavigateTo("/create-module")

	st.SelectOptionByVisibleText("#create-module-namespace", "moduledetails")
	fillOutModuleFieldByLabel(st, "Module Name", "fullypopulated")
	fillOutModuleFieldByLabel(st, "Provider", "Invalid Provider Name")

	clickCreateModuleButton(st)

	st.AssertElementVisible("#create-error")
	st.AssertTextContent("#create-error", "Module provider name is invalid")

	currentURL := st.GetCurrentURL()
	assert.Equal(t, st.GetURL("/create-module"), currentURL)
}

// Helper functions for module provider tests

// fillOutModuleFieldByLabel finds input field by label and fills out input.
// Python reference: /app/test/selenium/test_create_module_provider.py - _fill_out_field_by_label
func fillOutModuleFieldByLabel(st *SeleniumTest, label, input string) {
	err := st.runChromedp(
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(fmt.Sprintf(`
				(function() {
					var form = document.getElementById('create-module-form');
					var labels = form.getElementsByTagName('label');
					for (var i = 0; i < labels.length; i++) {
						if (labels[i].textContent === %q) {
							var parent = labels[i].parentElement;
							var inputElem = parent.querySelector('input');
							if (inputElem) {
								inputElem.value = '';
								inputElem.value = %q;
								var event = new Event('input', { bubbles: true });
								inputElem.dispatchEvent(event);
								return true;
							}
						}
					}
					return false;
				})()
			`, label, input), nil).Do(ctx)
		}),
	)
	require.NoError(st.t, err, "Failed to find field with label: %s", label)
}

// clickCreateModuleButton clicks the Create button.
// Python reference: /app/test/selenium/test_create_module_provider.py - _click_create
func clickCreateModuleButton(st *SeleniumTest) {
	// Python: self.selenium_instance.find_element(By.XPATH, "//button[text()='Create']").click()
	// Note: The button doesn't have type='submit', we need to find by text
	err := st.runChromedp(
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(`
				(function() {
					var buttons = document.getElementsByTagName('button');
					for (var i = 0; i < buttons.length; i++) {
						if (buttons[i].textContent === 'Create') {
							buttons[i].click();
							return true;
						}
					}
					return false;
				})()
			`, nil).Do(ctx)
		}),
	)
	require.NoError(st.t, err, "Failed to find and click Create button")
}
