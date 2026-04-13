package config

import (
	"os"
	"strings"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	Port         string
	DatabaseURL  string
	GoogleAPIKey string
	GCSBucket    string
	AllowedUsers []string
	AuthPassword string
}

// Load reads configuration from environment variables.
func Load() *Config {
	allowedUsers := getEnv("ALLOWED_USERS", "sunjin,velvet-panda,neon-otter,cosmic-tanuki,mellow-fox,turbo-finch,hazy-lynx,ember-crane,drift-moose,glitch-heron,plum-badger")
	users := strings.Split(allowedUsers, ",")
	for i := range users {
		users[i] = strings.TrimSpace(users[i])
	}

	return &Config{
		Port:         getEnv("PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://sunjinlee:db@localhost:5432/vibecomposer_dev?sslmode=disable"),
		GoogleAPIKey: getEnv("GOOGLE_API_KEY", ""),
		GCSBucket:    getEnv("GCS_BUCKET", ""),
		AllowedUsers: users,
		AuthPassword: getEnv("AUTH_PASSWORD", "vibecheck"),
	}
}

// IsAllowedUser checks if a username is in the allowed list.
func (c *Config) IsAllowedUser(username string) bool {
	for _, u := range c.AllowedUsers {
		if u == username {
			return true
		}
	}
	return false
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
