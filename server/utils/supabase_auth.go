package utils

import (
	"cinema-server/config"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// SupabaseAuth validates a Supabase access token. It supports:
// 1. Asymmetric keys via Supabase JWKS (ES256, RS256) at SUPABASE_URL/auth/v1/.well-known/jwks.json
// 2. Legacy HS256 project secret if SUPABASE_JWT_SECRET is configured.
func SupabaseAuth(cfg *config.WalletConfig) gin.HandlerFunc {
	validMethods := []string{"ES256", "RS256"}
	if cfg.SupabaseJWTSecret != "" {
		validMethods = append(validMethods, "HS256")
	}

	parser := jwt.NewParser(
		jwt.WithValidMethods(validMethods),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(cfg.SupabaseURL+"/auth/v1"),
		jwt.WithAudience("authenticated"),
	)

	var jwks keyfunc.Keyfunc
	if cfg.SupabaseURL != "" {
		jwksURL := cfg.SupabaseURL + "/auth/v1/.well-known/jwks.json"
		var err error
		jwks, err = keyfunc.NewDefault([]string{jwksURL})
		if err != nil {
			log.Printf("WARN: failed to initialize Supabase JWKS from %s: %v", jwksURL, err)
		}
	}

	secret := []byte(cfg.SupabaseJWTSecret)

	keyFunc := func(token *jwt.Token) (any, error) {
		if token.Method.Alg() == "HS256" {
			if len(secret) == 0 {
				return nil, fmt.Errorf("HS256 token received but SUPABASE_JWT_SECRET is not set")
			}
			return secret, nil
		}
		if jwks == nil {
			return nil, fmt.Errorf("no JWKS provider available for algorithm %s", token.Method.Alg())
		}
		return jwks.Keyfunc(token)
	}

	return func(c *gin.Context) {
		sub, ok := supabaseSubject(parser, keyFunc, c.GetHeader("Authorization"))
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "unauthorized", "message": "Invalid or missing token"})
			return
		}
		c.Set("supabase_user_id", sub)
		c.Next()
	}
}

func supabaseSubject(parser *jwt.Parser, keyFunc jwt.Keyfunc, header string) (string, bool) {
	scheme, token, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "bearer") || token == "" {
		return "", false
	}
	var claims jwt.RegisteredClaims
	if _, err := parser.ParseWithClaims(token, &claims, keyFunc); err != nil {
		return "", false
	}
	if _, err := uuid.Parse(claims.Subject); err != nil {
		return "", false
	}
	return claims.Subject, true
}
