package selenium

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultTimeout = 30 * time.Second
	pollInterval  = 100 * time.Millisecond
)

// SeleniumTest provides base functionality for Selenium tests.
// This is the Go equivalent of Python's test.selenium.SeleniumTest class.
//
// Python reference: /app/test/selenium/__init__.py - SeleniumTest class
type SeleniumTest struct {
	t              *testing.T
	server         *TestServer
	baseURL        string
	AllocCtx       context.Context // Exported chromedp allocator context
	allocCancel    context.CancelFunc
	ctxCancel      context.CancelFunc // Context cancel function for chromedp context
}

// NewSeleniumTest creates a new Selenium test instance.
// This is the Go equivalent of Python's setup_class method.
func NewSeleniumTest(t *testing.T) *SeleniumTest {
	st := &SeleniumTest{
		t: t,
	}

	st.setupServer()
	st.setupBrowser()

	return st
}

// NewSeleniumTestWithConfig creates a new Selenium test instance with custom config.
// This allows individual test classes to override configuration like Python's setup_class.
// Python reference: /app/test/selenium/test_homepage.py - TestHomepage.setup_class
func NewSeleniumTestWithConfig(t *testing.T, configOverrides map[string]string) *SeleniumTest {
	st := &SeleniumTest{
		t: t,
	}

	st.server = NewTestServer(st.t, configOverrides)
	st.baseURL = st.server.baseURL
	st.setupBrowser()

	return st
}

// setupServer starts a test Terrareg server with the actual application.
// This is the Go equivalent of Python's _setup_server method.
// Python reference: /app/test/selenium/__init__.py - SeleniumTest._setup_server()
func (st *SeleniumTest) setupServer() {
	// Create test server with default configuration
	// Individual tests can override this by calling NewTestServer directly
	configOverrides := ConfigForAdminTokenTests()
	st.server = NewTestServer(st.t, configOverrides)
	st.baseURL = st.server.baseURL
}

// setupBrowser initializes the Chrome browser for testing.
// This is the Go equivalent of Python's Selenium/Firefox setup.
func (st *SeleniumTest) setupBrowser() {
	// Create Chrome DP allocator options
	// Use headless mode unless RUN_INTERACTIVELY is set
	opts := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", os.Getenv("RUN_INTERACTIVELY") == ""),
		chromedp.Flag("disable-gpu", "true"),
		chromedp.Flag("no-sandbox", "true"),
		chromedp.Flag("disable-dev-shm-usage", "true"),
		chromedp.WindowSize(1920, 1080),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	}

	// Create the allocator context
	allocatorCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	st.allocCancel = allocCancel

	// Create the browser context from the allocator context
	// This is the standard chromedp pattern - the allocator is inherited from the parent
	ctx, cancel := chromedp.NewContext(allocatorCtx, chromedp.WithLogf(log.Printf))
	st.AllocCtx = ctx
	st.ctxCancel = cancel

	// Set a 20-second timeout for all chromedp operations
	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	st.ctxCancel = func() {
		cancel() // Cancel the timeout context
		st.ctxCancel() // Cancel the chromedp context
	}
	st.AllocCtx = ctx

	// Allocate the browser by running an initial task
	// The executor will be embedded in the context after this call
	if err := chromedp.Run(st.AllocCtx); err != nil {
		st.t.Fatalf("Failed to start browser: %v", err)
	}
}

// TearDown cleans up the Selenium test resources.
// This is the Go equivalent of Python's teardown_class method.
func (st *SeleniumTest) TearDown() {
	// Close browser context
	if st.ctxCancel != nil {
		st.ctxCancel()
	}

	// Close allocator
	if st.allocCancel != nil {
		st.allocCancel()
	}

	// Shutdown server
	if st.server != nil {
		st.server.Shutdown()
	}
}

// GetURL returns the full URL for a given path.
// This is the Go equivalent of Python's get_url method.
// Python reference: /app/test/selenium/__init__.py - get_url()
func (st *SeleniumTest) GetURL(path string) string {
	return st.baseURL + path
}

// runChromedp is a helper method to run chromedp actions with the proper context.
// This ensures that all chromedp operations have access to the browser executor.
func (st *SeleniumTest) runChromedp(actions ...chromedp.Action) error {
	return chromedp.Run(st.AllocCtx, actions...)
}

// NavigateTo navigates the browser to a specific URL.
func (st *SeleniumTest) NavigateTo(path string) {
	url := st.GetURL(path)
	err := st.runChromedp(chromedp.Navigate(url))
	require.NoError(st.t, err, "Failed to navigate to %s", url)
}

// WaitForElement waits for an element to be present and optionally visible.
// This is the Go equivalent of Python's wait_for_element method.
// Python reference: /app/test/selenium/__init__.py - wait_for_element()
func (st *SeleniumTest) WaitForElement(selector string, opts ...ElementOption) *Element {
	opt := &elementOptions{
		timeout:         defaultTimeout,
		ensureDisplayed: true,
	}
	for _, o := range opts {
		o(opt)
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), opt.timeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			require.Fail(st.t, "Element not found within timeout: %s", selector)
		case <-ticker.C:
			var found bool
			var visible bool

			// Use runChromedp to properly access the browser
			err := st.runChromedp(
				chromedp.ActionFunc(func(ctx context.Context) error {
					// Check if element exists
					var textContent string
					err := chromedp.Text(selector, &textContent, chromedp.ByQuery).Do(ctx)
					if err != nil {
						found = false
						return nil
					}
					found = true

					if !opt.ensureDisplayed {
						return nil
					}

					// Check if visible
					return chromedp.Evaluate(fmt.Sprintf(`
						(function() {
							var el = document.querySelector(%q);
							if (!el) return false;
							var rect = el.getBoundingClientRect();
							return rect.width > 0 && rect.height > 0;
						})()
					`, selector), &visible).Do(ctx)
				}),
			)

			if err == nil && found {
				if !opt.ensureDisplayed || visible {
					return &Element{
						selector: selector,
						ctx:      st.AllocCtx,
						st:       st,
					}
				}
			}
		}
	}
}

// GetTitle returns the page title.
func (st *SeleniumTest) GetTitle() string {
	var title string
	err := st.runChromedp(chromedp.Title(&title))
	require.NoError(st.t, err, "Failed to get page title")
	return title
}

// GetCurrentURL returns the current browser URL.
func (st *SeleniumTest) GetCurrentURL() string {
	var url string
	err := st.runChromedp(chromedp.Location(&url))
	require.NoError(st.t, err, "Failed to get current URL")
	return url
}

// DeleteCookiesAndLocalStorage clears all cookies and local storage.
// This is the Go equivalent of Python's delete_cookies_and_local_storage method.
// Python reference: /app/test/selenium/__init__.py - delete_cookies_and_local_storage()
func (st *SeleniumTest) DeleteCookiesAndLocalStorage() {
	// Navigate to a simple page first
	st.NavigateTo("/")

	// Clear cookies and local storage using chromedp.Run
	err := st.runChromedp(
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Use JavaScript to clear cookies
			return chromedp.Evaluate(`document.cookie.split(";").forEach(function(c) { document.cookie = c.replace(/^ +/, "").replace(/=.*/, "=;expires=" + new Date().toUTCString() + ";path=/"); });`, nil).Do(ctx)
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Clear local storage
			return chromedp.Evaluate("window.localStorage.clear();", nil).Do(ctx)
		}),
	)
	if err != nil {
		log.Printf("Warning: Failed to clear cookies/local storage: %v", err)
	}
}

// AssertTextContent asserts that an element contains specific text.
func (st *SeleniumTest) AssertTextContent(selector, expectedText string) {
	var text string
	err := st.runChromedp(chromedp.Text(selector, &text, chromedp.ByQuery))
	require.NoError(st.t, err, "Element not found: %s", selector)
	assert.Contains(st.t, text, expectedText, "Element text does not contain expected value")
}

// AssertElementVisible asserts that an element is visible.
func (st *SeleniumTest) AssertElementVisible(selector string) {
	var visible bool
	err := st.runChromedp(chromedp.Evaluate(fmt.Sprintf(`
		(function() {
			var el = document.querySelector(%q);
			if (!el) return false;
			var rect = el.getBoundingClientRect();
			return rect.width > 0 && rect.height > 0;
		})()
	`, selector), &visible))
	require.NoError(st.t, err, "Element not found: %s", selector)
	assert.True(st.t, visible, "Element exists but is not visible: %s", selector)
}

// AssertElementNotVisible asserts that an element either doesn't exist or is not visible.
func (st *SeleniumTest) AssertElementNotVisible(selector string) {
	var visible bool
	err := st.runChromedp(chromedp.Evaluate(fmt.Sprintf(`
		(function() {
			var el = document.querySelector(%q);
			if (!el) return true;
			var rect = el.getBoundingClientRect();
			return rect.width > 0 && rect.height > 0;
		})()
	`, selector), &visible))
	if err != nil {
		// Element doesn't exist or error - which is fine for "not visible"
		return
	}
	assert.False(st.t, visible, "Element should not be visible: %s", selector)
}

// AssertElementExists asserts that an element exists in the DOM.
func (st *SeleniumTest) AssertElementExists(selector string) {
	var text string
	err := st.runChromedp(chromedp.Text(selector, &text, chromedp.ByQuery))
	require.NoError(st.t, err, "Element not found: %s", selector)
}

// AssertElementNotExists asserts that an element does not exist in the DOM.
func (st *SeleniumTest) AssertElementNotExists(selector string) {
	var text string
	err := st.runChromedp(chromedp.Text(selector, &text, chromedp.ByQuery))
	assert.Error(st.t, err, "Element should not exist: %s", selector)
}

// ElementOption is a function that modifies element options.
type ElementOption func(*elementOptions)

// elementOptions holds options for WaitForElement.
type elementOptions struct {
	timeout         time.Duration
	ensureDisplayed bool
}

// WithTimeout sets a custom timeout for WaitForElement.
func WithTimeout(timeout time.Duration) ElementOption {
	return func(o *elementOptions) {
		o.timeout = timeout
	}
}

// WithoutVisibilityCheck disables the visibility check in WaitForElement.
func WithoutVisibilityCheck() ElementOption {
	return func(o *elementOptions) {
		o.ensureDisplayed = false
	}
}

// Element represents a DOM element.
// This is the Go equivalent of Python's WebElement.
type Element struct {
	selector string
	ctx      context.Context
	st       *SeleniumTest
}

// Text returns the text content of the element.
func (e *Element) Text() string {
	var text string
	err := e.st.runChromedp(chromedp.Text(e.selector, &text, chromedp.ByQuery))
	require.NoError(e.st.t, err, "Failed to get text for element: %s", e.selector)
	return text
}

// Click clicks the element.
func (e *Element) Click() {
	err := e.st.runChromedp(chromedp.Click(e.selector, chromedp.ByQuery))
	require.NoError(e.st.t, err, "Failed to click element: %s", e.selector)
}

// IsDisplayed returns true if the element is visible.
func (e *Element) IsDisplayed() bool {
	var visible bool
	err := e.st.runChromedp(chromedp.Evaluate(fmt.Sprintf(`
		(function() {
			var el = document.querySelector(%q);
			if (!el) return false;
			var rect = el.getBoundingClientRect();
			return rect.width > 0 && rect.height > 0;
		})()
	`, e.selector), &visible))
	if err != nil {
		return false
	}
	return visible
}

// Exists returns true if the element exists in the DOM.
func (e *Element) Exists() bool {
	var text string
	err := e.st.runChromedp(chromedp.Text(e.selector, &text, chromedp.ByQuery))
	return err == nil
}

// SendKeys sends keystrokes to the element.
func (e *Element) SendKeys(keys string) {
	err := e.st.runChromedp(chromedp.SendKeys(e.selector, keys, chromedp.ByQuery))
	require.NoError(e.st.t, err, "Failed to send keys to element: %s", e.selector)
}
