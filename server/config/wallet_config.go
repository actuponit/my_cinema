package config

import (
	"fmt"
	"os"
	"strings"
)

type WalletConfig struct {
	SupabaseURL         string
	SupabaseJWTSecret   string
	VerifyAPIBaseURL    string
	VerifyAPIKey        string
	VerifyWebhookSecret string
	WalletWebURL        string
	ActionSecret        string
}

func NewWalletConfig() (*WalletConfig, error) {
	cfg := &WalletConfig{
		SupabaseURL:         strings.TrimRight(os.Getenv("SUPABASE_URL"), "/"),
		SupabaseJWTSecret:   os.Getenv("SUPABASE_JWT_SECRET"),
		VerifyAPIBaseURL:    strings.TrimRight(getEnvOrDefault("VERIFY_API_BASE_URL", "https://checkoutapi.verify.et"), "/"),
		VerifyAPIKey:        os.Getenv("VERIFY_API_KEY"),
		VerifyWebhookSecret: os.Getenv("VERIFY_WEBHOOK_SECRET"),
		WalletWebURL:        strings.TrimRight(os.Getenv("WALLET_WEB_URL"), "/"),
		ActionSecret:        os.Getenv("HASURA_ACTION_SECRET"),
	}

	required := map[string]string{
		"SUPABASE_URL":          cfg.SupabaseURL,
		"VERIFY_API_KEY":        cfg.VerifyAPIKey,
		"VERIFY_WEBHOOK_SECRET": cfg.VerifyWebhookSecret,
		"WALLET_WEB_URL":        cfg.WalletWebURL,
		"HASURA_ACTION_SECRET":  cfg.ActionSecret,
	}
	var missing []string
	for name, v := range required {
		if v == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}
	return cfg, nil
}
