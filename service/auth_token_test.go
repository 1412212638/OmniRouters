package service

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestDashboardAuthTokenRoundTripAndIsolation(t *testing.T) {
	previous := common.SessionSecret
	common.SessionSecret = "test-session-secret-with-sufficient-entropy"
	t.Cleanup(func() { common.SessionSecret = previous })
	identity := AuthIdentity{UserID: 42, SessionID: "session-1", UserAuthVersion: 3, SessionVersion: 2}
	token, expiresAt, err := IssueAccessToken(identity)
	require.NoError(t, err)
	require.Positive(t, expiresAt)
	parsed, err := ParseAccessToken(token)
	require.NoError(t, err)
	require.Equal(t, identity, parsed)
	binding, err := BindVerificationOperation(VerificationOperation{Scope: VerificationScopeChannelKeyRead, Context: []byte(`{"channel_id":123}`)})
	require.NoError(t, err)
	proofClaims := authClaims{TokenUse: securityProofTokenUse, SessionID: identity.SessionID, UserAuthVersion: identity.UserAuthVersion, SessionVersion: identity.SessionVersion, RegisteredClaims: jwt.RegisteredClaims{Issuer: authTokenIssuer, Subject: "42", Audience: jwt.ClaimStrings{authTokenAudience}, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)), IssuedAt: jwt.NewNumericDate(time.Now()), NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Second)), ID: "proof"}}
	proof, err := jwt.NewWithClaims(jwt.SigningMethodHS256, proofClaims).SignedString(authSigningKey(securityProofTokenUse))
	require.NoError(t, err)
	_ = binding
	_, err = ParseAccessToken(proof)
	require.ErrorIs(t, err, ErrAuthTokenInvalid)
}

func TestDashboardAuthTokenRejectsTamperingAndExpiry(t *testing.T) {
	previous := common.SessionSecret
	common.SessionSecret = "test-session-secret-with-sufficient-entropy"
	t.Cleanup(func() { common.SessionSecret = previous })
	token, _, err := IssueAccessToken(AuthIdentity{UserID: 42, SessionID: "session-1", UserAuthVersion: 1, SessionVersion: 1})
	require.NoError(t, err)
	tampered := token[:len(token)-2] + "x" + token[len(token)-1:]
	_, err = ParseAccessToken(tampered)
	require.ErrorIs(t, err, ErrAuthTokenInvalid)
	expired := authClaims{TokenUse: accessTokenUse, SessionID: "session-1", UserAuthVersion: 1, SessionVersion: 1, RegisteredClaims: jwt.RegisteredClaims{Issuer: authTokenIssuer, Subject: "42", Audience: jwt.ClaimStrings{authTokenAudience}, ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)), IssuedAt: jwt.NewNumericDate(time.Now().Add(-2*time.Minute)), NotBefore: jwt.NewNumericDate(time.Now().Add(-2*time.Minute)), ID: "expired"}}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expired).SignedString(authSigningKey(accessTokenUse))
	require.NoError(t, err)
	_, internal, err := ParseDashboardAccessToken(raw)
	require.True(t, internal)
	require.ErrorIs(t, err, ErrAuthTokenExpired)
}
