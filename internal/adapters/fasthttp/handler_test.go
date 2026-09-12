package fasthttpadp

import (
	"crypto/sha1"
	"encoding/base64"
	"strings"
	"testing"
)

func TestVerifyRequestSignature(t *testing.T) {
	password := "super-secret"
	body := []byte(`{"status":"paid","external_id":"payment_123"}`)
	encodedBody := base64.URLEncoding.EncodeToString(body)
	sum := sha1.Sum([]byte(password + encodedBody + password))
	expected := base64.URLEncoding.EncodeToString(sum[:])
	rawExpected := strings.TrimSuffix(expected, "=")

	if !verifyRequestSignature(body, password, expected) {
		t.Fatalf("expected valid padded-url signature for payload")
	}
	if !verifyRequestSignature(body, password, rawExpected) {
		t.Fatalf("expected valid raw-url signature for payload")
	}
	if verifyRequestSignature(body, password, "invalid") {
		t.Fatal("expected invalid signature to be rejected")
	}
}
