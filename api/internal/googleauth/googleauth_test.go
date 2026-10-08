package googleauth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-jose/go-jose/v4"
	"github.com/google/go-cmp/cmp"
)

const (
	webClientID   = "web-client.apps.googleusercontent.com"
	otherClientID = "someone-elses-app.apps.googleusercontent.com"
)

func TestVerify(t *testing.T) {
	t.Parallel()

	googleKey := newKey(t)
	attackerKey := newKey(t)
	verifier := newVerifier(&oidc.StaticKeySet{PublicKeys: []crypto.PublicKey{&googleKey.PublicKey}}, []string{webClientID})

	validClaims := func() map[string]any {
		return map[string]any{
			"iss":   issuer,
			"aud":   webClientID,
			"sub":   "google-sub-1",
			"name":  "Alice",
			"nonce": "nonce-1",
			"iat":   time.Now().Unix(),
			"exp":   time.Now().Add(time.Hour).Unix(),
		}
	}

	tests := []struct {
		name    string
		key     *rsa.PrivateKey
		modify  func(claims map[string]any)
		want    Identity
		wantErr bool
	}{
		{
			name:   "valid token",
			key:    googleKey,
			modify: func(map[string]any) {},
			want:   Identity{Subject: "google-sub-1", Name: "Alice", Nonce: "nonce-1"},
		},
		{
			name:    "issued for another app",
			key:     googleKey,
			modify:  func(c map[string]any) { c["aud"] = otherClientID },
			wantErr: true,
		},
		{
			name:    "additional untrusted audience",
			key:     googleKey,
			modify:  func(c map[string]any) { c["aud"] = []string{webClientID, otherClientID} },
			wantErr: true,
		},
		{
			name:    "expired",
			key:     googleKey,
			modify:  func(c map[string]any) { c["exp"] = time.Now().Add(-time.Minute).Unix() },
			wantErr: true,
		},
		{
			name:    "different issuer",
			key:     googleKey,
			modify:  func(c map[string]any) { c["iss"] = "https://evil.example.com" },
			wantErr: true,
		},
		{
			name:    "not signed by Google",
			key:     attackerKey,
			modify:  func(map[string]any) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			claims := validClaims()
			tt.modify(claims)

			got, err := verifier.Verify(t.Context(), sign(t, tt.key, claims))

			if tt.wantErr {
				if err == nil {
					t.Errorf("Verify() = %+v, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Verify() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func newKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	return key
}

func sign(t *testing.T, key *rsa.PrivateKey, claims map[string]any) string {
	t.Helper()
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, nil)
	if err != nil {
		t.Fatalf("create signer: %v", err)
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	signed, err := signer.Sign(payload)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	token, err := signed.CompactSerialize()
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	return token
}
