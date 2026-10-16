package utils

import (
	"cinema-server/config"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestSupabaseAuth_HS256(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-supabase-jwt-secret-string-32-chars-long"
	cfg := &config.WalletConfig{
		SupabaseURL:       "https://test.supabase.co",
		SupabaseJWTSecret: secret,
	}

	makeToken := func(sub string, iss string, aud string, exp time.Time) string {
		claims := jwt.RegisteredClaims{
			Subject:   sub,
			Issuer:    iss,
			Audience:  jwt.ClaimStrings{aud},
			ExpiresAt: jwt.NewNumericDate(exp),
		}
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := tok.SignedString([]byte(secret))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}
		return signed
	}

	validUserID := uuid.New().String()

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantUser   string
	}{
		{
			name:       "missing header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "malformed header",
			authHeader: "Basic abc",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "valid HS256 token",
			authHeader: "Bearer " + makeToken(validUserID, "https://test.supabase.co/auth/v1", "authenticated", time.Now().Add(time.Hour)),
			wantStatus: http.StatusOK,
			wantUser:   validUserID,
		},
		{
			name:       "expired token",
			authHeader: "Bearer " + makeToken(validUserID, "https://test.supabase.co/auth/v1", "authenticated", time.Now().Add(-time.Hour)),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong audience",
			authHeader: "Bearer " + makeToken(validUserID, "https://test.supabase.co/auth/v1", "wrong-aud", time.Now().Add(time.Hour)),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong issuer",
			authHeader: "Bearer " + makeToken(validUserID, "https://wrong.supabase.co/auth/v1", "authenticated", time.Now().Add(time.Hour)),
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			var capturedUser string
			r.GET("/protected", SupabaseAuth(cfg), func(c *gin.Context) {
				capturedUser = c.GetString("supabase_user_id")
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, w.Code)
			}
			if tc.wantStatus == http.StatusOK && capturedUser != tc.wantUser {
				t.Fatalf("expected user %s, got %s", tc.wantUser, capturedUser)
			}
		})
	}
}

func TestSupabaseAuth_ES256_JWKS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Generate ECDSA private key (P-256) for ES256
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ecdsa key: %v", err)
	}

	kid := "test-key-id-123"
	xBytes := privKey.PublicKey.X.Bytes()
	yBytes := privKey.PublicKey.Y.Bytes()
	// Pad to 32 bytes if necessary
	pad := func(b []byte, size int) []byte {
		if len(b) >= size {
			return b
		}
		padded := make([]byte, size)
		copy(padded[size-len(b):], b)
		return padded
	}
	xB64 := base64.RawURLEncoding.EncodeToString(pad(xBytes, 32))
	yB64 := base64.RawURLEncoding.EncodeToString(pad(yBytes, 32))

	jwksJSON := fmt.Sprintf(`{"keys":[{"kty":"EC","crv":"P-256","alg":"ES256","use":"sig","kid":"%s","x":"%s","y":"%s"}]}`, kid, xB64, yB64)

	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/v1/.well-known/jwks.json" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(jwksJSON))
			return
		}
		http.NotFound(w, r)
	}))
	defer jwksServer.Close()

	cfg := &config.WalletConfig{
		SupabaseURL: jwksServer.URL,
	}

	validUserID := uuid.New().String()

	claims := jwt.RegisteredClaims{
		Subject:   validUserID,
		Issuer:    jwksServer.URL + "/auth/v1",
		Audience:  jwt.ClaimStrings{"authenticated"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tok.Header["kid"] = kid
	tokenString, err := tok.SignedString(privKey)
	if err != nil {
		t.Fatalf("failed to sign token with ECDSA: %v", err)
	}

	r := gin.New()
	var capturedUser string
	r.GET("/protected", SupabaseAuth(cfg), func(c *gin.Context) {
		capturedUser = c.GetString("supabase_user_id")
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
	if capturedUser != validUserID {
		t.Fatalf("expected user %s, got %s", validUserID, capturedUser)
	}
}

