// Package googleauth verifies Google ID tokens. See docs/AUTH.md for the
// checks and why each one matters.
package googleauth

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/coreos/go-oidc/v3/oidc"
)

const (
	issuer = "https://accounts.google.com"
	// Google's signing keys in JWK format. They rotate; go-oidc caches them and
	// refetches when a token is signed with a key it hasn't seen.
	jwksURL = "https://www.googleapis.com/oauth2/v3/certs"
)

// Identity is what a verified ID token tells us about the user.
type Identity struct {
	// Subject is Google's stable, never-reused user ID (the sub claim).
	Subject string
	// Name is empty if the token has no name claim (the profile scope wasn't
	// requested).
	Name string
	// Nonce is not checked here: that needs the database, so the caller does it.
	Nonce string
}

// Verifier verifies Google ID tokens issued for one of our OAuth clients.
type Verifier struct {
	verifier  *oidc.IDTokenVerifier
	clientIDs []string
}

// New returns a Verifier that accepts tokens whose audience is in clientIDs.
// Keys are fetched lazily on the first Verify, using ctx for the requests.
func New(ctx context.Context, clientIDs []string) *Verifier {
	return newVerifier(oidc.NewRemoteKeySet(ctx, jwksURL), clientIDs)
}

func newVerifier(keySet oidc.KeySet, clientIDs []string) *Verifier {
	return &Verifier{
		// go-oidc checks the signature, issuer and expiry. Its client ID check
		// takes a single ID, so it's skipped here and Verify checks the audience
		// against the whole list instead.
		verifier:  oidc.NewVerifier(issuer, keySet, &oidc.Config{SkipClientIDCheck: true}),
		clientIDs: clientIDs,
	}
}

// Verify checks rawIDToken and returns the identity it asserts.
func (v *Verifier) Verify(ctx context.Context, rawIDToken string) (Identity, error) {
	token, err := v.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return Identity{}, fmt.Errorf("verify id token: %w", err)
	}
	if err := v.checkAudience(token.Audience); err != nil {
		return Identity{}, err
	}

	var claims struct {
		Name string `json:"name"`
	}
	if err := token.Claims(&claims); err != nil {
		return Identity{}, fmt.Errorf("decode id token claims: %w", err)
	}
	return Identity{Subject: token.Subject, Name: claims.Name, Nonce: token.Nonce}, nil
}

// checkAudience requires every audience to be one of ours, not just one of
// them, per OIDC Core 3.1.3.7: reject tokens with "additional audiences not
// trusted by the Client".
func (v *Verifier) checkAudience(audience []string) error {
	if len(audience) == 0 {
		return errors.New("id token has no audience")
	}
	for _, aud := range audience {
		if !slices.Contains(v.clientIDs, aud) {
			return fmt.Errorf("id token audience %q is not one of our client IDs", aud)
		}
	}
	return nil
}
