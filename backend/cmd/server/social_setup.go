// path: backend/cmd/server/social_setup.go
package main

import (
	"log"

	"github.com/techappsUT/social-queue/internal/config"
	"github.com/techappsUT/social-queue/internal/social"
	"github.com/techappsUT/social-queue/internal/social/adapters"
)

// setupSocialAdapters initializes and registers all social platform adapters
func setupSocialAdapters(cfg *config.Config) *social.AdapterRegistry {
	registry := social.NewAdapterRegistry()

	// Register Twitter adapter
	if cfg.Twitter.ClientID != "" && cfg.Twitter.ClientSecret != "" {
		twitterAdapter := adapters.NewTwitterAdapter(
			cfg.Twitter.ClientID,
			cfg.Twitter.ClientSecret,
		)
		if err := registry.Register(twitterAdapter); err != nil {
			log.Printf("Warning: Failed to register Twitter adapter: %v", err)
		} else {
			log.Println("✓ Twitter adapter registered successfully")
		}
	} else {
		log.Println("⚠ Twitter credentials not configured, skipping adapter registration")
	}

	// Register Facebook adapter
	if cfg.Facebook.AppID != "" && cfg.Facebook.AppSecret != "" {
		facebookAdapter := adapters.NewFacebookAdapter(
			cfg.Facebook.AppID,
			cfg.Facebook.AppSecret,
		)
		if err := registry.Register(facebookAdapter); err != nil {
			log.Printf("Warning: Failed to register Facebook adapter: %v", err)
		} else {
			log.Println("✓ Facebook adapter registered successfully")
		}
	} else {
		log.Println("⚠ Facebook credentials not configured, skipping adapter registration")
	}

	// Register LinkedIn adapter
	if cfg.LinkedIn.ClientID != "" && cfg.LinkedIn.ClientSecret != "" {
		linkedinAdapter := adapters.NewLinkedInAdapter(
			cfg.LinkedIn.ClientID,
			cfg.LinkedIn.ClientSecret,
		)
		if err := registry.Register(linkedinAdapter); err != nil {
			log.Printf("Warning: Failed to register LinkedIn adapter: %v", err)
		} else {
			log.Println("✓ LinkedIn adapter registered successfully")
		}
	} else {
		log.Println("⚠ LinkedIn credentials not configured, skipping adapter registration")
	}

	// TODO: Add more platforms as you implement them
	// - Instagram adapter
	// - TikTok adapter
	// - YouTube adapter

	// Log summary
	platforms := registry.ListPlatforms()
	if len(platforms) == 0 {
		log.Println("⚠ WARNING: No social platform adapters registered! Configure platform credentials in environment variables.")
	} else {
		log.Printf("✓ Total %d platform adapter(s) registered: %v", len(platforms), platforms)
	}

	return registry
}

// setupFacebookWebhooks initializes Facebook webhook handlers
func setupFacebookWebhooks(cfg *config.Config) *adapters.FacebookWebhookHandler {
	if cfg.Facebook.AppSecret == "" || cfg.Facebook.WebhookVerifyToken == "" {
		log.Println("⚠ Facebook webhook credentials not configured, skipping webhook handler")
		return nil
	}

	handler := adapters.NewFacebookWebhookHandler(
		cfg.Facebook.AppSecret,
		cfg.Facebook.WebhookVerifyToken,
	)

	log.Println("✓ Facebook webhook handler initialized")
	return handler
}

// validateSocialConfig checks if essential social platform configurations are present
func validateSocialConfig(cfg *config.Config) error {
	// At least one platform should be configured
	hasAtLeastOne := false

	if cfg.Twitter.ClientID != "" && cfg.Twitter.ClientSecret != "" {
		hasAtLeastOne = true
	}
	if cfg.Facebook.AppID != "" && cfg.Facebook.AppSecret != "" {
		hasAtLeastOne = true
	}
	if cfg.LinkedIn.ClientID != "" && cfg.LinkedIn.ClientSecret != "" {
		hasAtLeastOne = true
	}

	if !hasAtLeastOne {
		log.Println("⚠ WARNING: No social platform credentials configured. The application will start but social features will be unavailable.")
	}

	return nil
}
