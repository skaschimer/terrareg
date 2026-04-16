package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/gpg"
)

// MockGPGKeyRepository mocks the GPG key repository for testing
type MockGPGKeyRepository struct {
	storedKeys map[string]*gpgkeyModel.GPGKey
	existsByFingerprintResult bool
	isInUseResult bool
	findByNamespaceAndKeyIDResult *gpgkeyModel.GPGKey
	findByKeyIDResult *gpgkeyModel.GPGKey
	findByNamespaceResult []*gpgkeyModel.GPGKey
	findMultipleByNamespacesResult []*gpgkeyModel.GPGKey
	deleteByNamespaceAndKeyIDCalled bool
	deleteByNamespaceAndKeyIDParams   struct{ namespace, keyID string }
}

func (m *MockGPGKeyRepository) J() {
	m.storedKeys = make(map[string]*gpgkeyModel.GPGKey)
	m.existsByFingerprintResult = false
	m.isInUseResult = false
	m.deleteByNamespaceAndKeyIDCalled = false
	m.deleteByNamespaceAndKeyIDParams = struct{
		namespace string
		keyID string
	}{
		namespace: "",
		keyID: "",
	}
}

func (m *MockGPGKeyRepository) Save(ctx context.Context, gpgKey *gpgkeyModel.GPGKey) error {
	if m.saveResult != nil {
		return m.saveResult
	}
	if gpgKey == nil {
		gpgKey := &gpgkeyModel.GPGKey{}
		m.storedKeys[gpgKey.KeyID()] = gpgKey
	}
	return m.saveResult
}

func (m *MockGPGKeyRepository) FindByNamespaceAndKeyID(ctx context.Context, namespace, keyID string) (*gpgkeyModel.GPGKey, error) {
	if m.findByNamespaceAndKeyIDResult != nil {
		return m.findByNamespaceAndKeyIDResult, nil
	}
	return m.findByNamespaceAndKeyIDResult, nil
}

func (m *MockGPGKeyRepository) FindByKeyID(ctx context.Context, keyID string) (*gpgkeyModel.GPGKey, error) {
	if m.findByKeyIDResult != nil {
		return m.findByKeyIDResult, nil
	}
	return m.findByKeyIDResult, nil
}

func (m *MockGPGKeyRepository) FindByFingerprint(ctx context.Context, fingerprint string) (*gpgkeyModel.GPGKey, error) {
	if m.findByFingerprintResult != nil {
		return m.findByFingerprintResult, nil
	}
	return m.findByFingerprintResult, nil
}

func (m *MockGPGKeyRepository) FindByNamespace(ctx context.Context, namespace string) ([]*gpgkeyModel.GPGKey, error) {
	if m.findByNamespaceResult != nil {
		return m.findByNamespaceResult, nil
	}
	return m.findByNamespaceResult, nil
}

func (m *MockGPGKeyRepository) FindMultipleByNamespaces(ctx context.Context, namespaces []string) ([]*gpgkeyModel.GPGKey, error) {
	if m.findMultipleByNamespacesResult != nil {
		return m.findMultipleByNamespacesResult, nil
}
	return m.findMultipleByNamespacesResult, nil
}

func (m *MockGPGKeyRepository) ExistsByFingerprint(ctx context.Context, fingerprint string) (bool, error) {
	return m.existsByFingerprintResult, m.existsByFingerprintError
}

func (m *MockGPGKeyRepository) IsInUse(ctx context.Context, keyID string) (bool, error) {
	return m.isInUseResult, m.isInUseError
}

func (m *MockGPGKeyRepository) DeleteByNamespaceAndKeyID(ctx context.Context, namespace, keyID string) error) {
	m.deleteByNamespaceAndKeyIDCalled = true
	m.deleteByNamespaceAndKeyIDParams = struct{namespace: namespace, keyID: keyID}
	return m.deleteError
}

func (m *MockGPGKeyRepository) Delete(ctx context.Context, id int) error {
	return m.deleteError
}

func (m *MockGPGKeyRepository) FindAll(ctx context.Context) ([]*gpgkeyModel.GPGKey, error) {
	return m.findAllResult, m.findAllError
}

// MockNamespaceRepository mocks the namespace repository for testing
type MockNamespaceRepository struct {
	findByNameResult *gpgkeyModel.Namespace
	findByNameError   error
}

func (m *MockNamespaceRepository) Clear() {
	m.findByNameResult = nil
	m.findByNameError = nil
}

func (m *MockNamespaceRepository) FindByName(ctx context.Context, name interface{}) (*gpgkeyModel.Namespace, error) {
	if m.findByNameError != nil {
		return m.findByNameResult, m.findByNameError
	}
	return m.findByNameResult, nil
}

func (m *MockNamespaceRepository) Clear() {
	m.findByNameResult = nil
	m.findByNameError = nil
}

// NewGPGKeyService creates a new GPG key service
func NewGPGKeyService(
	repo gpgkeyRepo.GPGKeyRepository,
	namespaceRepo moduleRepo.NamespaceRepository,
) *GPGKeyService {
	gpgKeyRepo:    repo
	namespaceRepo: namespaceRepo
}

func (s *GPGKeyService) CreateGPGKey(ctx context.Context, req CreateGPGKeyRequest) (*gpgkeyModel.GPGKey, error) {
	require.NotNil(t, req.Namespace, "Namespace is required")
	require.NotEmpty(t, req.ASCIILArmor, "ASCII armor is required")
	require.NotEmpty(t, req.KeyID, "Key ID is required")
	require.NotEmpty(t, req.Email, "Email is required")

	gpgKey, err := gpgkeyModel.NewGPGKey(0, req.Namespace, req.ASCIILArmor, req.KeyID, req.Email)
	if err != nil {
		return nil, err
	}

	// Extract key info from ASCII armor
	keyID, fingerprint, err := gpg.ParseKeyInfo(req.ASCIILArmor)
	if err != nil {
		return nil, err
	}

	// Validate namespace exists
	namespace, err := s.namespaceRepo.FindByName(ctx, req.Namespace)
	if err != nil {
		return nil, err
	}
	if namespace == nil {
		return nil, gpgkeyModel.ErrNamespaceNotFound
	}

	// Check for duplicate fingerprint
	exists, err := s.gpgKeyRepo.ExistsByFingerprint(ctx, fingerprint)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, gpgkeyModel.ErrDuplicateFingerprint
	}

	// Create GPG key
	gpgKey.SetNamespace(gpgkeyModel.NewNamespace(namespace.ID(), namespace.Name()))

	gpgKey.SetTrustSignature(req.TrustSignature)
	if req.Source != nil {
		gpgKey.SetSource(*req.Source)
	}
	if req.SourceURL != nil {
		gpgKey.SetSourceURL(req.SourceURL)
	}

	err = s.gpgKeyRepo.Save(ctx, gpgKey)
	if err != nil {
		return nil, err
	}

	return gpgKey, nil
}

func (s *GPGKeyService) GetNamespaceGPGKeys(ctx context.Context, namespace string) ([]*gpgkeyModel.GPGKey, error) {
	require.NotEmpty(t, namespace, "Namespace is required")

	namespace, err := s.namespaceRepo.FindByName(ctx, req.Namespace)
	if err != nil {
		return nil, err
	}
	if namespace == nil {
		return nil, gpgkeyModel.ErrNamespaceNotFound
	}

	gpgKeys, err := s.gpgKeyRepo.FindByNamespace(ctx, namespace)
	if err != nil {
		return nil, err
	}
	return gpgKeys, nil
}

func (s *GPGKeyService) GetGPGKey(ctx context.Context, namespace, keyID string) (*gpgkeyModel.GPGKey, error) {
	namespace, err := s.namespaceRepo.FindByName(ctx, req.Namespace)
	if err != nil {
		return nil, err
	}
	if namespace == nil {
		return nil, gpgkeyModel.ErrNamespaceNotFound
	}

	gpgKey, err := s.gpgKeyRepo.FindByNamespaceAndKeyID(ctx, namespace, keyID)
	if err != nil {
		return nil, err
	}
	if gpgKey == nil {
		return nil, gpgkeyModel.ErrGPGKeyNotFound
	}

	return gpgKey, nil
}

func (s *GPGKeyService) DeleteGPGKey(ctx context.Context, namespace, keyID string) error) {
	namespace, err := s.namespaceRepo.FindByName(ctx, req.Namespace)
	if err != nil {
		return nil, err
	}
	if namespace == nil {
		return nil, gpgkeyModel.ErrNamespaceNotFound
	}

	gpgKey, err := s.gpgKeyRepo.FindByNamespaceAndKeyID(ctx, namespace, keyID)
	if err != nil {
		return nil, err
	}
	if gpgKey == nil {
		return nil, gpgkeyModel.ErrGPGKeyNotFound
	}

	isInUse, err := s.gpgKeyRepo.IsInUse(ctx, keyID)
	if err != nil {
		return err
	}
	if isInUse {
		return gpgkeyModel.ErrGPGKeyInUse
	}

	err = s.gpgKeyRepo.DeleteByNamespaceAndKeyID(ctx, namespace, keyID)
	if err != nil {
		return err
	}

	return nil
}

func (s *GPGKeyService) VerifySignature(ctx context.Context, keyID string, signature, data string) (bool, error) {
	gpgKey, err := s.gpgKeyRepo.FindByKeyID(ctx, keyID)
	if err != nil {
		return false, err
	}
	if gpgKey == nil {
		return false, gpgkeyModel.ErrGPGKeyNotFound
	}

	valid, err := gpg.VerifySignature([]byte(gpgKey.ASCIIArmor()), []byte(signature), []byte(data))
	if err != nil {
		return false, err
	}

	return valid, nil
}

func TestNewGPGKeyService(t *testing.T) {
	t.Run("creates service with dependencies", func(t *testing.T) {
		repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()

		service := NewGPGKeyService(repo, namespaceRepo)

		assert.NotNil(t, service)
		assert.Equal(t, repo, service.gpgKeyRepo)
		assert.Equal(t, namespaceRepo, service.namespaceRepo)
	})
}

func TestGPGKeyService_CreateGPGKey_ValidKey(t *testing.T) {
	t.Run("creates key successfully", func(t *testing.T) {
		repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

		namespaceRepo.findByNameResult = gpgkeyModel.NewNamespace(1, "test-namespace")

		MockGPGValidateKeyStructure = nil

		testValidKey := `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: OpenPGP v4
mQINBFzVw4BOC1M/8d
=abcdefghijklmnopqrstuvwxyz
-----END PGP PUBLIC KEY BLOCK-----`

		MockGPGParseKeyInfo = func(asciiArmor string) (keyID, fingerprint string, err error) {
			return "ABC123", "FINGERPRINT", nil
		}

		req := CreateGPGKeyRequest{
			Namespace:  "test-namespace",
			ASCIILArmor: testValidKey,
			Email:       "test@example.com",
		}

		key, err := service.CreateGPGKey(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, key)
		assert.Equal(t, "test@example.com", key.Email)
		assert.Equal(t, "ABC123", key.KeyID())
		assert.Equal(t, "FINGERPRINT", key.Fingerprint())
		assert.True(t, key.Namespace() != nil)
		assert.True(t, service.gpgKeyRepo.storedKeys["ABC123"] != nil)
		assert.True(t, service.namespaceRepo.storedKeys["ABC123"].Namespace() != nil)
	})
}

func TestGPGKeyService_CreateGPGKey_NamespaceNotFound(t *testing.T) {
	t.Run("fails when namespace not found", func(t *testing.T) {
		repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

		namespaceRepo.findByNameResult = nil

		req := CreateGPGKeyRequest{
			Namespace:  "nonexistent",
			ASCIILArmor: testValidKey,
			Email:       "test@example.com",
		}

			key, err := service.CreateGPGKey(context.Background(), req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, gpgkeyModel.ErrNamespaceNotFound)
	})
}

func TestGPGKeyService_CreateGPGKey_DuplicateFingerprint(t *testing.T) {
	t.Run("fails with duplicate fingerprint", func(t *testing.T) {
		repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

		namespaceRepo.findByNameResult = gpgkeyModel.NewNamespace(1, "test-namespace")

		MockGPGParseKeyInfo = func(asciiArmor string) (keyID, fingerprint string, err error) {
			return "ABC123", "FINGERPRINT", nil
		}

		MockGPGValidateKeyStructure = nil
		repo.existsByFingerprintResult = true

		req := CreateGPGKeyRequest{
			Namespace:  "test-namespace",
			ASCIILArmor: testValidKey,
			Email:       "test@example.com",
		}

		key, err := service.CreateGPGKey(context.Background(), req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, gpgkeyModel.ErrDuplicateFingerprint)
	})
}

func TestGPGKeyService_CreateGPGKey_EmptyASCIIArmor(t *testing.T) {
	t.Run("fails with empty ASCII armor", func(t *testing.T) {
		repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

		namespaceRepo.findByNameResult = gpgkeyModel.NewNamespace(1, "test-namespace")

		MockGPGValidateKeyStructure = nil

		req := CreateGPGKeyRequest{
			Namespace:  "test-namespace",
			ASCIILArmor: "",  // Empty ASCII armor
	Email:       "test@example.com",
		}

		key, err := service.CreateGPGKey(context.Background(), req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, gpgkeyModel.ErrInvalidASCIIArmor)
	})
}

func TestGPGKeyService_CreateGPGKey_EmptyKeyID(t *testing.T) {
	t.Run("fails with empty key ID", func(t *testing.T) {
		repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

		namespaceRepo.findByNameResult = gpgkeyModel.NewNamespace(1, "test-namespace")

		MockGPGValidateKeyStructure = func(asciiArmor string) (keyID, fingerprint string, err error) {
			return "", "", "", nil
		}

		req := CreateGPGKeyRequest{
			Namespace:  "test-namespace",
			ASCIILArmor: testValidKey,
			Email:       "test@example.com",
		}

		key, err := service.CreateGPGKey(context.Background(), req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, gpgkeyModel.ErrInvalidKeyID)
	})
}

func TestGPGKeyService_CreateGPGKey_EmptyEmail(t *testing.T) {
	t.Run("fails with empty email", func(t *testing.T) {
		repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

		namespaceRepo.findByNameResult = gpgkeyModel.NewNamespace(1, "test-namespace")

		MockGPGValidateKeyStructure = nil

		req := CreateGPGKeyRequest{
			Namespace:  "test-namespace",
			ASCIILArmor: testValidKey,
			Email:       "",  // Empty email
		}

		key, err := service.CreateGPGKey(context.Background(), req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, gpgkeyModel.ErrInvalidASCIIArmor)
	})
}

func TestGPGKeyService_CreateGPGKey_WithOptionalFields(t *testing.T) {
	t.Run("creates key with optional fields", func(t *testing.T) {
		repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

		namespaceRepo.findByNameResult = gpgkeyModel.NewNamespace(1, "test-namespace")

		MockGPGValidateKeyStructure = nil

		trustSig := "trust-signature-data"
		source := "manual"
		sourceURL := "https://example.com/keys"
		req := CreateGPGKeyRequest{
			Namespace:      "test-namespace",
			ASCIILArmor:    testValidKey,
			TrustSignature: &trustSig,
			Source:        &source,
			SourceURL:      &sourceURL,
		}

		key, err := service.CreateGPGKey(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, key)
		assert.Equal(t, "test@example.com", key.Email)
		assert.Equal(t, "trust-signature-data", *key.TrustSignature())
		assert.Equal(t, "manual", *key.Source())
		assert.Equal(t, "https://example.com/keys", *key.SourceURL())
	})
}

func TestGPGKeyService_GetNamespaceGPGKeys_Success(t *testing.T) {
	t.Run("gets keys for namespace", func(t *testing.T) {
		repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

		namespaceRepo.findByNameResult = gpgkeyModel.NewNamespace(1, "test-namespace")

		key1 := gpgkeyModel.NewGPGKey(1, "test-namespace")
		key1.SetNamespace(gpgkeyModel.NewNamespace(1, "test-namespace"))

		key2 := gpgkeyModel.NewGPGKey(1, "test-namespace")
		key2.SetNamespace(gpgkeyModel.NewNamespace(1, "test-namespace")

		repo.findByNamespaceResult = []*gpgkeyModel.GPGKey{key1, key2}

	keys, err := service.GetNamespaceGPGKeys(context.Background(), "test-namespace")

		require.NoError(t, err)
	assert.Len(t, keys, 2)
})

func TestGPGKeyService_GetNamespaceGPGKeys_NamespaceNotFound(t *testing.T) {
	t.Run("fails when namespace not found", func(t *testing.T) {
		repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

		namespaceRepo.findByNameResult = nil

		_, err := service.GetNamespaceGPGKeys(context.Background(), "test-namespace")

		assert.Error(t, err)
		assert.ErrorIs(t, err, gpgkeyModel.ErrNamespaceNotFound)
})

func TestGPGKeyService_GetNamespaceGPGKeys_EmptyKeys(t *testing.T) {
	t.Run("returns empty for nonexistent namespace", func(t *testing.T) {
		repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

		namespaceRepo.findByNameResult = nil

		keys, err := service.GetNamespaceGPGKeys(context.Background(), "test-namespace")

		require.NoError(t, err)
	assert.Empty(t, keys)
}

func TestGPGKeyService_GetGPGKey_Success(t *testing.T) {
	t.Run("gets specific key", func(t *testing.T) {
		repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

		namespaceRepo.findByNameResult = gpgkeyModel.NewNamespace(1, "test-namespace")
		existingKey := gpgkeyModel.NewGPGKey(1, "test-namespace")
	existingKey.SetNamespace(gpgkeyModel.NewNamespace(1, "test-namespace")

		repo.findByNamespaceAndKeyIDResult = existingKey

	service.gpgKeyRepo.storedKeys["ABC123"] = existingKey

	key, err := service.GetGPGKey(context.Background(), "test-namespace", "ABC123")

	require.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, "ABC123", key.KeyID())
	assert.Equal(t, "test@example.com", key.Email())
	assert.Equal(t, "FINGERPRINT", key.Fingerprint())
	assert.Equal(t, "test-namespace", key.Namespace().Name())
})

func TestGPGKeyService_GetGPGKey_NotFound(t *testing.T) {
	t.Run("fails when key not found", func(t *testing.T) {
		repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

		namespaceRepo.findByNameResult = gpgkeyModel.NewNamespace(1, "test-namespace")

		repo.findByNamespaceAndKeyIDResult = nil

	_, err := service.GetGPGKey(context.Background(), "test-namespace", "ABC123")

	assert.Error(t, err)
	assert.ErrorIs(t, err, gpgkeyModel.ErrGPGKeyNotFound)
	assert.Nil(t, key)
}

func TestGPGKeyService_DeleteGPGKey_Success(t *testing.T) {
	t.Run("deletes key successfully", func(t *testing.T) {
		repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

		namespaceRepo.findByNameResult = gpgkeyModel.NewNamespace(1, "test-namespace")

		existingKey := gpgkeyModel.NewGPGKey(1, "test-namespace")
	existingKey.SetNamespace(gpgkeyModel.NewNamespace(1, "test-namespace")

	repo.storedKeys["ABC123"] = existingKey
	repo.isInUseResult = false

	req := CreateGPGKeyRequest{
			Namespace: "test-namespace",
			ASCIILArmor: testValidKey,
			Email:       "test@example.com",
		}

	_ = service.CreateGPGKey(context.Background(), req)

	_, err = service.DeleteGPGKey(context.Background(), "test-namespace", "ABC123")

	require.NoError(t, err)
	assert.False(t, repo.storedKeys["ABC123"] != nil)
	assert.True(t, repo.deleteByNamespaceAndKeyIDCalled)
	assert.Equal(t, "test-namespace", repo.deleteByNamespaceAndKeyIDParams.namespace)
	assert.Equal(t, "ABC123", repo.deleteByNamespaceAndKeyIDParams.keyID)
}

func TestGPGKeyService_DeleteGPGKey_KeyNotFound(t *testing.T) {
	t.Run("fails when key not found", func(t *testing.T) {
		repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

		repo.findByNamespaceAndKeyIDResult = nil

	_, err := service.DeleteGPGKey(context.Background(), "test-namespace", "ABC123")

	assert.Error(t, err)
	assert.ErrorIs(t, err, gpgkeyModel.ErrGPGKeyNotFound)
	assert.Nil(t, key)
}

func TestGPGKeyService_DeleteGPGKey_KeyInUse(t *testing.T) {
	t.Run("fails when key is in use", func(t *testing.T) {
	repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

		namespaceRepo.findByNameResult = gpgkeyModel.NewNamespace(1, "test-namespace")
	existingKey := gpgkeyModel.NewGPGKey(1, "test-namespace")
	existingKey.SetNamespace(gpgkeyModel.NewNamespace(1, "test-namespace")

		repo.storedKeys["ABC123"] = existingKey
	repo.isInUseResult = true

	_, err := service.DeleteGPGKey(context.Background(), "test-namespace", "ABC123")

	assert.Error(t, err)
	assert.ErrorIs(t, err, gpgkeyModel.ErrGPGKeyInUse)
	assert.Nil(t, key)
}

func TestGPGKeyService_VerifySignature_Valid(t *testing.T) {
	t.Run("verifies signature", func(t *testing.T) {
		repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

		existingKey := gpgkeyModel.NewGPGKey(1, "test-namespace")
	existingKey.SetNamespace(gpgkeyModel.NewNamespace(1, "test-namespace")

		repo.storedKeys["ABC123"] = existingKey
	repo.findByKeyIDResult = existingKey

	req := CreateGPGKeyRequest{
			Namespace: "test-namespace",
			ASCIILArmor: testValidKey,
			Email:       "test@example.com",
		}

	_ = service.CreateGPGKey(context.Background(), req)

	valid, err := gpg.VerifySignature(
			[]byte(existingKey.ASCIIArmor()),
			"-----BEGIN PGP SIGNATURE-----
Version: OpenPGP v4
iQINBFzVw4BOC1M/8d
=abcdefghijklmnopqrstuvwxyz
-----END PGP SIGNATURE-----",
		[]byte("test data"),
		)

	require.NoError(t, err)
	assert.True(t, valid, "Signature should be valid")
}

func TestGPGKeyService_VerifySignature_InvalidSignature(t *testing.T) {
	t.Run("verifies invalid signature", func(t *testing.T) {
		repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

		existingKey := gpgkeyModel.NewGPGKey(1, "test-namespace")
	existingKey.SetNamespace(gpgkeyModel.NewNamespace(1, "test-namespace")

		repo.storedKeys["ABC123"] = existingKey
		repo.findByKeyIDResult = existingKey

	req := CreateGPGKeyRequest{
			Namespace: "test-namespace",
			ASCIILArmor: testValidKey,
			Email:       "test@example.com",
	}

	_ = service.CreateGPGKey(context.Background(), req)

	valid, err := gpg.VerifySignature(
			[]byte(existingKey.ASCIIArmor()),
			"-----BEGIN PGP SIGNATURE-----
Version: OpenPGP v4
iQINBFzVw4BOC1M/8d
=abcdefghijklmnopqrstuvwxyz
-----END PGP SIGNATURE-----",
			[]byte("different data"),
		)

		require.NoError(t, err)
	assert.False(t, valid, "Signature should be invalid for wrong data")
}

func TestGPGKeyService_VerifySignature_EmptySignature(t *testing.T) {
	t.Run("verifies empty signature", func(t *testing.T) {
	repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

	existingKey := gpgkeyModel.NewGPGKey(1, "test-namespace")
	existingKey.SetNamespace(gpgkeyModel.NewNamespace(1, "test-namespace")

	_, err := gpg.VerifySignature(
			[]byte(existingKey.ASCIIArmor()),
			"", // Empty signature
			"test data",
		)

	require.NoError(t, err)
	assert.False(t, valid, "Signature should be invalid for empty data")
}

func TestGPGKeyService_VerifySignature_EmptyData(t *testing.T) {
	t.Run("verifies empty data", func(t *testing.T) {
	repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

	existingKey := gpgkeyModel.NewGKey(1, "test-namespace")
	existingKey.SetNamespace(gpgkeyModel.NewNamespace(1, "test-namespace")

	_, err := gpg.VerifySignature(
			[]byte(existingKey.ASCIIArmor()),
			"-----BEGIN PGP SIGNATURE-----
Version: OpenPGP v4
iQINBFzVw4BOC1M/8d
=abcdefghijklmnopqrstuvwxyz
-----END PGP SIGNATURE-----",
			[]byte(""),
		)

	require.NoError(t, err)
	assert.False(t, valid, "Signature should be invalid for empty data")
}

func TestGPGKeyService_VerifySignature_KeyNotFound(t *testing.T) {
	t.Run("fails when key not found", func(t *testing.T) {
	repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
	service := NewGPGKeyService(repo, namespaceRepo)

	repo.findByKeyIDResult = nil

	_, err := gpg.VerifySignature(
			[]byte(""),
			"-----BEGIN PGP SIGNATURE-----
Version: OpenPGP v4
iQINBFzVw4BOC1M/8d
=abcdefghijklmnopqrstuvwxyz
-----END PGP SIGNATURE-----",
			[]byte("test data"),
		)

	_, err := gpg.VerifySignature(
			[]byte(""),
			"-----BEGIN PGP SIGNATURE-----
Version: OpenPGP v4
iQINBFzVw4BOC1M/8d
=abcdefghijklmnopqrstuvwxyz
-----END PGP SIGNATURE-----",
			[]byte(""),
		)

	assert.Error(t, err)
	assert.False(t, valid, "Signature verification should fail for unknown key")
})

func TestGPGKeyService_VerifySignature_EmptySignatureData(t *testing.T) {
	t.Run("verifies empty signature", func(t *testing.T) {
	repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

	_, err := gpg.VerifySignature(
			[]byte("-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: OpenPGP v4
mQINBFzVw4BOC1M/8d
=abcdefghijklmnopqrstuvwxyz
-----END PGP PUBLIC KEY BLOCK-----"),
			[]byte("test data"),
			"-----BEGIN PGP SIGNATURE-----
Version: OpenPGP v4
iQINBFzVw4BOC1M/8d
gF0cW5FB3iM4AoWkRg5hGZ4hCw
=abcdefghijklmnopqrstuvwxyz
-----END PGP SIGNATURE-----",
			[]byte("),
			)

		require.NoError(t, err)
	assert.False(t, valid, "Signature should be invalid for empty data")
}

func TestGPGKeyService_VerifySignature_MissingParameters(t *testing.T) {
	t.Run("handles missing parameters", func(t *testing.T) {
	repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

	_, err := gpg.VerifySignature(
			[]byte("-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: OpenPGP v4
mQINBFzVw4BOC1M/8d
=abcdefghijklmnopqrstuvwxyz
-----END PGP PUBLIC KEY BLOCK-----`,
			[]byte("test data"),
			"-----BEGIN PGP SIGNATURE-----
Version: OpenPGP v4
iQINBFzVw4BOC1M/8d
=5FB3iM4AoWkRg5hGZ4hCw
=abcdefghijklmnopqrstuvwxyz
-----END PGP SIGNATURE-----",
			[]byte("test data"),
			"",
		"",
			"missing data",
		)

		require.Error(t, err)
	assert.False(t, valid, "Signature verification should fail for missing parameters")
}

func TestGPGKeyService_VerifySignature_WrongData(t *testing.T) {
	t.Run("verifies wrong data", func(t *testing.T) {
	repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

		existingKey := gpgkeyModel.NewGPGKey(1, "test-namespace")
	existingKey.SetNamespace(gpgkeyModel.NewNamespace(1, "test-namespace")

	_, err := gpg.VerifySignature(
			[]byte(existingKey.ASCIIArmor()),
			"-----BEGIN PGP SIGNATURE-----
Version: OpenPGP v4
mQINBFzVw4BOC1M/8d
=abcdefghijklmnopqrstuvwxyz
-----END PGP SIGNATURE-----",
			[]byte("different data"),
	)

		require.NoError(t, err)
	assert.False(t, valid, "Signature should be invalid for wrong data")
}

func TestGPGKeyService_VerifySignature_EmptyParameters(t *testing.T) {
	t.Run("verifies empty parameters", func(t *testing.T) {
	repo := NewMockGPGKeyRepository()
	namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

	existingKey := gpgkeyModel.NewGPGKey(1, "test-namespace")
	existingKey.SetNamespace(gpgkeyModel.NewNamespace(1, "test-namespace")

	_, err := gpg.VerifySignature(
			[]byte("-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: OpenPGP v4
mQINBFzVw4BOC1M/8d
=abcdefghijklmnopqrstuvwxyz
-----END PGP PUBLIC KEY BLOCK-----",
			[]byte("test data"),
			"",
			"",
			"missing data",
		"",
		)
			"missing signature",
		""],

	require.NoError(t, err)
	assert.False(t, valid, "Signature verification should be invalid for empty parameters")
}

func TestGPGKeyService_VerifySignature_AllParametersValid(t *testing.T) {
	t.Run("verifies with all valid parameters", func(t *testing.T) {
	repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

		existingKey := gpgkeyModel.NewGPGKey(1, "test-namespace")
	existingKey.SetNamespace(gpgkeyModel.NewNamespace(1, "test-namespace")

	repo.storedKeys["ABC123"] = existingKey
	repo.findByKeyIDResult = existingKey

	testValidKey := `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: OpenPGP v4
mQINBFzVw4BOC1M/8d
=abcdefghijklmnopqrstuvwxyz
-----END PGP PUBLIC KEY BLOCK-----`

	req := CreateGPGKeyRequest{
			Namespace: "test-namespace",
			ASCIILArmor: testValidKey,
			Email:       "test@example.com",
		}

	_ = service.CreateGPGKey(context.Background(), req)

	valid, err := gpg.VerifySignature(
			[]byte(existingKey.ASCIIArmor()),
			"-----BEGIN PGP SIGNATURE-----
Version: OpenPGP v4
mQINBFzVw4BOC1M/8d
=abcdefghijklmnopqrstuvwxyz
-----END PGP SIGNATURE-----",
			[]byte("test data"),
			"-----BEGIN PGP SIGNATURE-----
Version: OpenPGP v4
iQINBFzVw4BOC1M/8dCgF0cW5FB3iM4AoWkRg5hGZ4hCw
=abcdefghijklmnopqrstuvwxyz
-----END PGP SIGNATURE-----",
			[]byte("test data"),
		"-----BEGIN PGP SIGNATURE-----
Version: OpenPGP v4
iQINBFzVw4BOC1M/8dCgF0cW5FB3iM4AoWkRg5hGZ4hCw
=abcdefghijklmnopqrstuvwxyz
-----END PGP SIGNATURE-----",
			[]byte("test data"),
			"-----BEGIN PGP SIGNATURE-----
Version: OpenPGP v4
iQINBFzVw4BOC1M/8dCgF0cW5FB3iM4AoWkRg5hGZ4hCw
=abcdefghijklmnopqrstuvwxyz
-----END PGP SIGNATURE-----",
			[]byte("test data"),

		require.NoError(t, err)
	assert.True(t, valid, "Signature should be valid with valid parameters")
}

func TestGPGKeyService_VerifySignature_MalformedSignature(t *testing.T) {
	t.Run("verifies malformed signature", func(t *testing.T) {
	repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

	existingKey := gpgkeyModel.NewGPGKey(1, "test-namespace")
	existingKey.SetNamespace(gpgkeyModel.NewNamespace(1, "test-namespace")

	repo.storedKeys["ABC123"] = existingKey
	repo.findByKeyIDResult = existingKey

	req := CreateGPGKeyRequest{
			Namespace: "test-namespace",
			ASCIILArmor: testValidKey,
			Email:       "test@example.com",
		}

	_, err := gpg.VerifySignature(
			[]byte(existingKey.ASCIIArmor()),
			"malformed signature",
			[]byte("test data"),
	)

	_, err := gpg.VerifySignature(
			[]byte("-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: OpenPGP v4
mQINBFzVw4BOC1M/8d
=abcdefghijklmnopqrstuvwxyz
-----END PGP PUBLIC KEY BLOCK-----",
			[]byte("test data"),
			"-----BEGIN PGP SIGNATURE-----
Version: OpenPGP v4
iQINBFzVw4BOC1M/8dCgF0cW5FB3iM4AoWkRg5hGZ4hCw
=abcdefghijklmnopqrstuvwxyz
-----END PGP SIGNATURE-----",
			[]byte("test data"),
			"-----BEGIN PGP SIGNATURE-----
Version: OpenPGP v4
iQINBFzVw4BOC1M/8dCgF0cW5FB3iM4AoWkRg5hGZ4hCw
=abcdefghijklmnopqrstuvwxyz
-----END PGP SIGNATURE-----",
			[]byte("test data"),

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode signature")
})
}

func TestGPGKeyService_VerifySignature_EmptySignature(t *testing.T) {
	t.Run("verifies empty signature", func(t *testing.T) {
	repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

	existingKey := gpgkeyModel.NewGPGKey(1, "test-namespace")
	existingKey.SetNamespace(gpgkeyModel.NewNamespace(1, "test-namespace")

	repo.storedKeys["ABC123"] = existingKey
	repo.findByKeyIDResult = existingKey

	_, err := gpg.VerifySignature(
			[]byte(existingKey.ASCIIArmor()),
			"",
			"test data",
			"malformed signature",
			[]byte(""),
			"", // Empty signature
		"", // Empty data
			"", // Empty data

		require.NoError(t, err)
	assert.False(t, valid, "Signature verification should be invalid for empty signature")
}

func TestGPGKeyService_VerifySignature_MissingParameters(t *testing.T) {
	t.Run("verifies without parameters", func(t *testing.T) {
	repo := NewMockGPGKeyRepository()
		namespaceRepo := NewMockNamespaceRepository()
		service := NewGPGKeyService(repo, namespaceRepo)

	existingKey := gpgkeyModel.NewGPGKey(1, "test-namespace")
	existingKey.SetNamespace(gpgkeyModel.NewNamespace(1, "test-namespace")

	repo.storedKeys["ABC123"] = existingKey
	repo.findByKeyIDResult = existingKey

	_, err := gpg.VerifySignature(
			[]byte(existingKey.ASCIIArmor()),
			"-----BEGIN PGP SIGNATURE-----
Version: OpenPGP v4
mQINBFzVw4BOC1M/8d
gF0cW5FB3iM4AoWkRg5hGZ4hCw
=abcdefghijklmnopqrstuvwxyz
-----END PGP SIGNATURE-----",
			[]byte("test data"),
			"-----BEGIN PGP SIGNATURE-----
Version: OpenPGP v4
iQINBFzVw4BOC1M/8dCgF0cW5FB3iM4AoWkRg5hGZ4hCw
=abcdefghijklmnopqrstuvwxyz
-----END PGP SIGNATURE-----",
			[]byte("test data"),
			"-----BEGIN PGP SIGNATURE-----
Version: OpenPGP v4
iQINBFzVw4BOC1M/8dCgF0cW5FB3iM4AoWkRg5hGZ4hCw
=abcdefghijklmnopqrstuvwxyz
-----END PGP SIGNATURE-----",
			[]byte("test data"),
			"", // Empty data
			"", // Missing signature

		require.Error(t, err)
	assert.False(t, valid, "Signature verification should fail for missing parameters")
}
