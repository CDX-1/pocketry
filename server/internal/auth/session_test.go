package auth

import "testing"

// checks that session token is generated and non-empty
func TestGenerateSessionToken(t *testing.T) {
	token, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken returned error: %v", err)
	}

	if token == "" {
		t.Fatal("expected token to not be empty")
	}
}

// checks that session tokens are unique
func TestGenerateSessionTokenProducesUniqueTokens(t *testing.T) {
	token1, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken returned error: %v", err)
	}

	token2, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken returned error: %v", err)
	}

	if token1 == token2 {
		t.Fatal("expected tokens to be unique")
	}
}

// checks that session token hashing is consistent, secure, and non-empty
func TestHashSessionToken(t *testing.T) {
	token := "example-token"

	hash1 := HashSessionToken(token)
	hash2 := HashSessionToken(token)

	if hash1 == "" || hash2 == "" {
		t.Fatal("expected hashes to not be empty")
	}

	if hash1 != hash2 {
		t.Fatal("expected hashes to be identical")
	}

	if hash1 == token {
		t.Fatal("expected hash to not be equal to token")
	}
}
 
// checks that bearer token is extracted correctly
func TestExtractBearerToken(t *testing.T) {
	token, err := ExtractBearerToken("Bearer token123")
	if err != nil {
		t.Fatalf("ExtractBearerToken returned error: %v", err)
	}

	if token != "token123" {
		t.Fatalf("expected token to be 'token123'")
	}
}


// checks that bearer token is rejected when header is malformed
func TestExtractBearerTokenRejectsBadHeader(t *testing.T) {
	_, err := ExtractBearerToken("abc123")
	if err == nil {
		t.Fatal("expected error for missing Bearer prefix")
	}
}