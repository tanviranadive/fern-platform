package config

// TestUsers contains all test user configurations
var TestUsers = struct {
	Admin       User
	FernManager User
	FernUser    User
	AtmosUser   User
}{
	Admin: User{
		Username: "admin@fern.com",
		Password: "admin123",
		Role:     "admin",
		Teams:    []string{"all"}, // Admin has access to all teams
	},
	FernManager: User{
		Username: "fern-manager@fern.com",
		Password: "test123",
		Role:     "manager",
		Teams:    []string{"fern"},
	},
	FernUser: User{
		Username: "fern-user@fern.com",
		Password: "test123",
		Role:     "user",
		Teams:    []string{"fern"},
	},
	AtmosUser: User{
		Username: "atmos-user@fern.com",
		Password: "test123",
		Role:     "user",
		Teams:    []string{"atmos"}, // Different team, no access to fern team
	},
}

// User represents a test user configuration
type User struct {
	Username string
	Password string
	Role     string
	Teams    []string
}