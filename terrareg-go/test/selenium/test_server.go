package selenium

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	mathrand "math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	domainConfigService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/container"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/version"
)

// generateTestSigningKey generates a test RSA signing key and saves it to a file.
// This is required for Terraform OIDC tests.
func generateTestSigningKey(t *testing.T) string {
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	require.NoError(t, err, "Failed to generate RSA key")

	// Encode private key to PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyBytes := pem.EncodeToMemory(privateKeyPEM)

	// Create a temporary file for the signing key
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "signing_key.pem")

	err = os.WriteFile(keyPath, privateKeyBytes, 0600)
	require.NoError(t, err, "Failed to write signing key file")

	return keyPath
}

// TestServer manages a test instance of the Terrareg server.
// This is the Go equivalent of Python's test.selenium.SeleniumTest server setup.
// Python reference: /app/test/selenium/__init__.py - SeleniumTest._setup_server()
type TestServer struct {
	t               *testing.T
	container       *container.Container
	httpServer      *http.Server
	port            int
	baseURL         string
	db              *sqldb.Database
	configOverrides map[string]string
	logger          zerolog.Logger
	serverCtx       context.Context
	serverCancel    context.CancelFunc
	serverWg        sync.WaitGroup
}

// NewTestServer creates and starts a new test Terrareg server.
// This is the Go equivalent of Python's setup_class method.
// Python reference: /app/test/selenium/__init__.py - SeleniumTest.setup_class()
func NewTestServer(t *testing.T, configOverrides map[string]string) *TestServer {
	ts := &TestServer{
		t:               t,
		configOverrides: configOverrides,
		logger:          zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger(),
	}

	// Generate test signing key for Terraform OIDC
	signingKeyPath := generateTestSigningKey(t)
	if ts.configOverrides == nil {
		ts.configOverrides = make(map[string]string)
	}
	ts.configOverrides["TERRAFORM_OIDC_IDP_SIGNING_KEY_PATH"] = signingKeyPath

	ts.setup()
	return ts
}

// setup initializes the test server following the bootstrap pattern from cmd/server/main.go
func (ts *TestServer) setup() {
	// Set default config values if not provided
	// Python reference: /app/test/selenium/__init__.py - _get_database_path() returns 'temp-selenium.db'
	defaults := map[string]string{
		"LISTEN_PORT":                      "5000", // Valid port (will be overridden to random port after bootstrap)
		"PUBLIC_URL":                       "http://127.0.0.1:5000",
		"DATABASE_URL":                     "sqlite:///temp-selenium.db", // File-based DB like Python tests
		"SECRET_KEY":                       "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		"ADMIN_AUTHENTICATION_TOKEN":       "test-admin-token",
		"ALLOW_MODULE_HOSTING":             "true",
		"DEBUG":                            "true",
		"SESSION_COOKIE_NAME":              "terrareg_session",
		"SESSION_EXPIRY_MINS":              "60",
		"ADMIN_SESSION_EXPIRY_MINS":         "60",
		"SESSION_REFRESH_MINS":             "5",
		"TRUSTED_NAMESPACE_LABEL":          "Trusted",
		"CONTRIBUTED_NAMESPACE_LABEL":      "Contributed",
		"VERIFIED_MODULE_LABEL":            "Verified",
		"ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER": "true",
		"ALLOW_CUSTOM_GIT_URL_MODULE_VERSION":  "true",
		"AUTO_CREATE_NAMESPACE":            "true",
		"AUTO_CREATE_MODULE_PROVIDER":      "true",
		"DISABLE_ANALYTICS":                "false",
		"AUTO_PUBLISH_MODULE_VERSIONS":     "true",
		"MODULE_VERSION_REINDEX_MODE":      "legacy",
		"PRODUCT":                          "terraform",
		"DEFAULT_TERRAFORM_VERSION":        "1.5.7",
		"MANAGE_TERRAFORM_RC_FILE":         "false",
		"MODULES_DIRECTORY":                "modules",
		"EXAMPLES_DIRECTORY":               "examples",
		"PROVIDER_SOURCES":                 "[]",
		"PROVIDER_CATEGORIES":              `[{"id": 1, "name": "Example Category", "slug": "example-category", "user-selectable": true}]`,
		"GITHUB_URL":                       "https://github.com",
		"GITHUB_API_URL":                   "https://api.github.com",
		"GITHUB_LOGIN_TEXT":                "Login with Github",
		"OPENID_CONNECT_LOGIN_TEXT":        "Login using OpenID Connect",
		"SAML2_LOGIN_TEXT":                 "Login using SAML",
		"INFRACOST_TLS_INSECURE_SKIP_VERIFY": "false",
		"ALLOW_UNIDENTIFIED_DOWNLOADS":     "false",
		// Terraform OIDC settings (will be overridden with generated key)
		"TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		"TERRAFORM_OIDC_IDP_SESSION_EXPIRY":       "3600",
	}

	// Merge user overrides with defaults
	for k, v := range ts.configOverrides {
		defaults[k] = v
	}
	ts.configOverrides = defaults

	// Set environment variables (equivalent to Python's unittest.mock.patch)
	for k, v := range ts.configOverrides {
		os.Setenv(k, v)
	}

	// Bootstrap the application (following cmd/server/main.go pattern)
	ts.bootstrap()

	// Find an available port in 20000-21000 range (like Python tests)
	port := ts.findAvailablePort()
	ts.port = port
	ts.baseURL = fmt.Sprintf("http://127.0.0.1:%d", port)

	// Create HTTP server with the chi router from container.Server
	// Override the ListenPort in the container's InfraConfig
	ts.container.InfraConfig.ListenPort = port
	ts.container.InfraConfig.PublicURL = ts.baseURL

	// Get the router from the container's Server
	router := ts.container.Server.GetRouter()

	// Create our own http.Server to control the port
	ts.serverCtx, ts.serverCancel = context.WithCancel(context.Background())
	ts.httpServer = &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%d", port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in background
	ts.serverWg.Add(1)
	go func() {
		defer ts.serverWg.Done()
		ts.t.Logf("Test server listening on %s", ts.baseURL)
		if err := ts.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ts.t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to be ready
	ts.waitForServer()
}

// bootstrap bootstraps the application following the pattern in cmd/server/main.go
func (ts *TestServer) bootstrap() {
	// Load configuration using the new configuration service
	versionReader := version.NewVersionReader()
	configService := domainConfigService.NewConfigurationService(
		domainConfigService.ConfigurationServiceOptions{},
		versionReader,
	)

	domainConfig, infraConfig, err := configService.LoadConfiguration()
	require.NoError(ts.t, err, "Failed to load configuration")

	ts.t.Logf("Configuration loaded - port: %d, public_url: %s", infraConfig.ListenPort, infraConfig.PublicURL)

	// Initialize database
	ts.t.Log("Connecting to database")
	ts.db, err = sqldb.NewDatabase(infraConfig.DatabaseURL, infraConfig.Debug)
	require.NoError(ts.t, err, "Failed to connect to database")

	ts.t.Log("Database connected successfully")

	// Run auto-migration for all models (from cmd/server/main.go)
	ts.t.Log("Running database auto-migration")
	err = ts.autoMigrate()
	require.NoError(ts.t, err, "Failed to auto-migrate database")

	// Initialize dependency injection container with new configuration architecture
	ts.t.Log("Initializing application container")
	ts.container, err = container.NewContainer(
		domainConfig,
		infraConfig,
		configService,
		ts.logger,
		ts.db,
	)
	require.NoError(ts.t, err, "Failed to create container")
}

// autoMigrate runs GORM auto-migration for all models
// From cmd/server/main.go
func (ts *TestServer) autoMigrate() error {
	return ts.db.DB.AutoMigrate(
		&sqldb.SessionDB{},
		&sqldb.TerraformIDPAuthorizationCodeDB{},
		&sqldb.TerraformIDPAccessTokenDB{},
		&sqldb.TerraformIDPSubjectIdentifierDB{},
		&sqldb.UserGroupDB{},
		&sqldb.NamespaceDB{},
		&sqldb.UserGroupNamespacePermissionDB{},
		&sqldb.GitProviderDB{},
		&sqldb.NamespaceRedirectDB{},
		&sqldb.ModuleDetailsDB{},
		&sqldb.ModuleProviderDB{},
		&sqldb.ModuleVersionDB{},
		&sqldb.ModuleProviderRedirectDB{},
		&sqldb.SubmoduleDB{},
		&sqldb.AnalyticsDB{},
		&sqldb.ProviderAnalyticsDB{},
		&sqldb.ExampleFileDB{},
		&sqldb.ModuleVersionFileDB{},
		&sqldb.GPGKeyDB{},
		&sqldb.ProviderSourceDB{},
		&sqldb.ProviderCategoryDB{},
		&sqldb.RepositoryDB{},
		&sqldb.ProviderDB{},
		&sqldb.ProviderVersionDB{},
		&sqldb.ProviderVersionDocumentationDB{},
		&sqldb.ProviderVersionBinaryDB{},
		// Note: AuthenticationTokenDB excluded due to ENUM type incompatibility with SQLite
		&sqldb.AuditHistoryDB{},
	)
}

// findAvailablePort finds an available port in the range 20000-21000 (like Python tests).
// Python reference: /app/test/selenium/__init__.py - _setup_server() method
func (ts *TestServer) findAvailablePort() int {
	// Try ports in range 20000-21000 (matching Python's range)
	for i := 0; i < 100; i++ {
		port := 20000 + mathrand.Intn(1000)
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			listener.Close()
			return port
		}
	}
	// If all ports in range are busy, use system-assigned port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(ts.t, err, "Failed to find any available port")
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port
}

// waitForServer waits for the server to be ready to accept connections.
func (ts *TestServer) waitForServer() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			ts.t.Fatal("Server did not start within timeout")
		case <-time.After(100 * time.Millisecond):
			client := &http.Client{Timeout: 1 * time.Second}
			resp, err := client.Get(ts.baseURL + "/")
			if err == nil {
				resp.Body.Close()
				return
			}
		}
	}
}

// GetURL returns the full URL for a given path.
// Python reference: /app/test/selenium/__init__.py - get_url()
func (ts *TestServer) GetURL(path string) string {
	return ts.baseURL + path
}

// GetPort returns the port the server is listening on.
func (ts *TestServer) GetPort() int {
	return ts.port
}

// GetContainer returns the DI container (useful for setting up test data).
func (ts *TestServer) GetContainer() *container.Container {
	return ts.container
}

// GetDB returns the database connection (useful for setting up test data).
func (ts *TestServer) GetDB() *sqldb.Database {
	return ts.db
}

// Shutdown stops the test server and cleans up resources.
// Python reference: /app/test/selenium/__init__.py - _teardown_server()
func (ts *TestServer) Shutdown() {
	// Shutdown HTTP server
	if ts.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ts.httpServer.Shutdown(ctx)
	}

	// Cancel server context
	if ts.serverCancel != nil {
		ts.serverCancel()
	}

	// Wait for server goroutines to finish
	done := make(chan struct{})
	go func() {
		ts.serverWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Server stopped successfully
	case <-time.After(10 * time.Second):
		ts.t.Log("Warning: Server did not stop within timeout")
	}

	// Close database
	if ts.db != nil {
		ts.db.Close()
	}
}
