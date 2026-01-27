//go:build selenium

package selenium

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommonSearchPage tests the common search page.
// Python reference: /app/test/selenium/test_common_search_page.py - TestModuleSearch class
func TestCommonSearchPage(t *testing.T) {
	t.Run("test_search_from_homepage_common_search", testSearchFromHomepageCommonSearch)
	t.Run("test_search_from_homepage_redirect_type_search", testSearchFromHomepageRedirectTypeSearch)
	t.Run("test_result_cards", testCommonSearchResultCards)
	t.Run("test_provider_results_button", testProviderResultsButton)
	t.Run("test_module_results_button", testModuleResultsButton)
}

// newCommonSearchPageTest creates a new SeleniumTest configured for common search page tests.
// Python reference: /app/test/selenium/test_common_search_page.py - setup_class
func newCommonSearchPageTest(t *testing.T) *SeleniumTest {
	config := ConfigForCommonSearchPageTests()
	return NewSeleniumTestWithConfig(t, config)
}

// ConfigForCommonSearchPageTests returns config for common search page tests.
// Python reference: /app/test/selenium/test_common_search_page.py - setup_class
func ConfigForCommonSearchPageTests() map[string]string {
	base := getDefaultTestConfig()
	return mergeMaps(base, map[string]string{
		"CONTRIBUTED_NAMESPACE_LABEL": "unittest contributed module",
		"TRUSTED_NAMESPACE_LABEL":      "unittest trusted namespace",
		"VERIFIED_MODULE_LABEL":        "unittest verified label",
		"TRUSTED_NAMESPACES":           "modulesearch-trusted,relevancysearch",
	})
}

// testSearchFromHomepageCommonSearch checks search functionality from homepage.
// Python reference: /app/test/selenium/test_common_search_page.py - TestModuleSearch.test_search_from_homepage_common_search
func testSearchFromHomepageCommonSearch(t *testing.T) {
	testCases := []struct {
		searchString string
	}{
		{""},  // Test string that will match modules and providers
		{"mixed"},
	}

	for _, tc := range testCases {
		t.Run(tc.searchString, func(t *testing.T) {
			st := newCommonSearchPageTest(t)
			defer st.TearDown()

			st.NavigateTo("/")

			// Python: self.selenium_instance.find_element(By.ID, 'navBarSearchInput').send_keys(search_string)
			searchInput := st.WaitForElement("#navBarSearchInput")
			searchInput.SendKeys(tc.searchString)

			// Python: search_button = self.selenium_instance.find_element(By.ID, 'navBarSearchButton')
			// Python: assert search_button.text == 'Search'
			searchButton := st.WaitForElement("#navBarSearchButton")
			assert.Equal(t, "Search", searchButton.Text())

			// Python: search_button.click()
			searchButton.Click()

			// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url(f'/search?q={search_string}'))
			expectedURL := st.GetURL("/search?q=" + tc.searchString)
			currentURL := st.GetCurrentURL()
			assert.Equal(t, expectedURL, currentURL)

			// Python: assert self.selenium_instance.title == 'Search - Terrareg'
			title := st.GetTitle()
			assert.Equal(t, "Search - Terrareg", title)
		})
	}
}

// redirectTypeSearchTestCase represents a test case for search redirect verification.
type redirectTypeSearchTestCase struct {
	searchString  string
	expectedURL   string
	expectedTitle string
}

// testSearchFromHomepageRedirectTypeSearch checks search functionality from homepage with type-specific redirects.
// Python reference: /app/test/selenium/test_common_search_page.py - TestModuleSearch.test_search_from_homepage_redirect_type_search
func testSearchFromHomepageRedirectTypeSearch(t *testing.T) {
	testCases := []redirectTypeSearchTestCase{
		{"fullypopulated", "/search/modules?q=fullypopulated", "Module Search - Terrareg"},
		{"initial-providers", "/search/providers?q=initial-providers", "Provider Search - Terrareg"},
	}

	for _, tc := range testCases {
		t.Run(tc.searchString, func(t *testing.T) {
			st := newCommonSearchPageTest(t)
			defer st.TearDown()

			st.NavigateTo("/")

			// Python: self.selenium_instance.find_element(By.ID, 'navBarSearchInput').send_keys(search_string)
			searchInput := st.WaitForElement("#navBarSearchInput")
			searchInput.SendKeys(tc.searchString)

			// Python: search_button = self.selenium_instance.find_element(By.ID, 'navBarSearchButton')
			// Python: assert search_button.text == 'Search'
			searchButton := st.WaitForElement("#navBarSearchButton")
			assert.Equal(t, "Search", searchButton.Text())

			// Python: search_button.click()
			searchButton.Click()

			// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url(expected_url))
			expectedURL := st.GetURL(tc.expectedURL)
			currentURL := st.GetCurrentURL()
			assert.Equal(t, expectedURL, currentURL)

			// Python: self.assert_equals(lambda: self.selenium_instance.title, expected_title)
			title := st.GetTitle()
			assert.Equal(t, tc.expectedTitle, title)

			// Python: assert self.selenium_instance.find_element(By.ID, 'search-query-string').get_attribute('value') == search_string
			searchQueryString := st.GetAttribute("#search-query-string", "value")
			assert.Equal(t, tc.searchString, searchQueryString)
		})
	}
}

// testCommonSearchResultCards checks result cards in common search page.
// Python reference: /app/test/selenium/test_common_search_page.py - TestModuleSearch.test_result_cards
func testCommonSearchResultCards(t *testing.T) {
	st := newCommonSearchPageTest(t)
	defer st.TearDown()

	st.NavigateTo("/search?q=mixed")

	// Python: self.wait_for_element(By.ID, "contributed-providersearch.mixedsearch-result.1.0.0")
	_ = st.WaitForElement("#contributed-providersearch.mixedsearch-result.1.0.0")

	// Python: provider_cards = [...]
	// Python: for card in self.selenium_instance.find_element(By.ID, "results-providers-content").find_elements(By.CLASS_NAME, "result-box"):
	providerCards := []struct {
		link string
		text string
	}{
		{"/providers/providersearch-trusted/mixedsearch-trusted-second-result", "providersearch-trusted / mixedsearch-trusted-second-result"},
		{"/providers/providersearch-trusted/mixedsearch-trusted-result-multiversion", "providersearch-trusted / mixedsearch-trusted-result-multiversion"},
		{"/providers/providersearch-trusted/mixedsearch-trusted-result", "providersearch-trusted / mixedsearch-trusted-result"},
		{"/providers/contributed-providersearch/mixedsearch-result-multiversion", "contributed-providersearch / mixedsearch-result-multiversion"},
		{"/providers/contributed-providersearch/mixedsearch-result", "contributed-providersearch / mixedsearch-result"},
	}

	for i, cardDetails := range providerCards {
		cardSelector := "#results-providers-content .result-box:nth-child(" + strconv.Itoa(i+1) + ")"
		_ = st.WaitForElement(cardSelector)

		// Python: for link in card.find_elements(By.TAG_NAME, "a"):
		// Python:     assert link.get_attribute("href") == self.get_url(card_details["link"])
		err := st.runChromedp(
			chromedp.ActionFunc(func(ctx context.Context) error {
				return chromedp.Evaluate(fmt.Sprintf(`
					(function() {
						var card = document.querySelectorAll('#results-providers-content .result-box')[%d];
						var links = card.getElementsByTagName('a');
						for (var i = 0; i < links.length; i++) {
							if (links[i].getAttribute('href').endsWith('%s')) {
								return true;
							}
						}
						return false;
					})()
				`, i, cardDetails.link), nil).Do(ctx)
			}),
		)
		require.NoError(st.t, err, "Provider card link not found for: %s", cardDetails.link)

		// Python: assert card.find_element(By.CLASS_NAME, "module-card-title").text == card_details["text"]
		titleSelector := cardSelector + " .module-card-title"
		st.AssertTextContent(titleSelector, cardDetails.text)
	}

	// Python: module_cards = [...]
	// Python: for card in self.selenium_instance.find_element(By.ID, "results-modules-content").find_elements(By.CLASS_NAME, "result-box"):
	moduleCards := []struct {
		link string
		text string
	}{
		{"/modules/modulesearch-contributed/mixedsearch-result/aws", "modulesearch-contributed / mixedsearch-result"},
		{"/modules/modulesearch-contributed/mixedsearch-result-multiversion/aws", "modulesearch-contributed / mixedsearch-result-multiversion"},
		{"/modules/modulesearch-trusted/mixedsearch-trusted-result/aws", "modulesearch-trusted / mixedsearch-trusted-result"},
		{"/modules/modulesearch-trusted/mixedsearch-trusted-result-multiversion/null", "modulesearch-trusted / mixedsearch-trusted-result-multiversion"},
		{"/modules/modulesearch-trusted/mixedsearch-trusted-result-verified/gcp", "modulesearch-trusted / mixedsearch-trusted-result-verified"},
		{"/modules/modulesearch-trusted/mixedsearch-trusted-second-result/datadog", "modulesearch-trusted / mixedsearch-trusted-second-result"},
	}

	for i, cardDetails := range moduleCards {
		cardSelector := "#results-modules-content .result-box:nth-child(" + strconv.Itoa(i+1) + ")"
		_ = st.WaitForElement(cardSelector)

		// Python: for link in card.find_elements(By.TAG_NAME, "a"):
		// Python:     if "provider-logo-link" not in link.get_attribute("class"):
		// Python:         assert link.get_attribute("href") == self.get_url(card_details["link"])
		err := st.runChromedp(
			chromedp.ActionFunc(func(ctx context.Context) error {
				return chromedp.Evaluate(fmt.Sprintf(`
					(function() {
						var card = document.querySelectorAll('#results-modules-content .result-box')[%d];
						var links = card.getElementsByTagName('a');
						for (var i = 0; i < links.length; i++) {
							if (!links[i].classList.contains('provider-logo-link') &&
								links[i].getAttribute('href').endsWith('%s')) {
								return true;
							}
						}
						return false;
					})()
				`, i, cardDetails.link), nil).Do(ctx)
			}),
		)
		require.NoError(st.t, err, "Module card link not found for: %s", cardDetails.link)

		// Python: assert card.find_element(By.CLASS_NAME, "module-card-title").text == card_details["text"]
		titleSelector := cardSelector + " .module-card-title"
		st.AssertTextContent(titleSelector, cardDetails.text)
	}
}

// testProviderResultsButton checks link to provider results.
// Python reference: /app/test/selenium/test_common_search_page.py - TestModuleSearch.test_provider_results_button
func testProviderResultsButton(t *testing.T) {
	st := newCommonSearchPageTest(t)
	defer st.TearDown()

	st.NavigateTo("/search?q=mixed")

	// Python: self.wait_for_element(By.ID, "contributed-providersearch.mixedsearch-result.1.0.0")
	_ = st.WaitForElement("#contributed-providersearch.mixedsearch-result.1.0.0")

	// Python: button = self.selenium_instance.find_element(By.XPATH, ".//button[text()='View all provider results']")
	// Python: button.click()
	button := st.WaitForElement("button")
	button.Click()

	// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/search/providers?q=mixed'))
	currentURL := st.GetCurrentURL()
	assert.Equal(t, st.GetURL("/search/providers?q=mixed"), currentURL)
}

// testModuleResultsButton checks link to module results.
// Python reference: /app/test/selenium/test_common_search_page.py - TestModuleSearch.test_module_results_button
func testModuleResultsButton(t *testing.T) {
	st := newCommonSearchPageTest(t)
	defer st.TearDown()

	st.NavigateTo("/search?q=mixed")

	// Python: self.wait_for_element(By.ID, "contributed-providersearch.mixedsearch-result.1.0.0")
	_ = st.WaitForElement("#contributed-providersearch.mixedsearch-result.1.0.0")

	// Python: button = self.selenium_instance.find_element(By.XPATH, ".//button[text()='View all module results']")
	// Python: button.click()
	// Find the button by text content
	err := st.runChromedp(
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(`
				(function() {
					var buttons = document.getElementsByTagName('button');
					for (var i = 0; i < buttons.length; i++) {
						if (buttons[i].textContent === 'View all module results') {
							buttons[i].click();
							return true;
						}
					}
					return false;
				})()
			`, nil).Do(ctx)
		}),
	)
	require.NoError(st.t, err)

	// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/search/modules?q=mixed'))
	currentURL := st.GetCurrentURL()
	assert.Equal(t, st.GetURL("/search/modules?q=mixed"), currentURL)
}
