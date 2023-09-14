package config

// CORS defines the Faucet CORS configuration
type CORS struct {
	// A list of origins a cross-domain request can be executed from.
	// If the special '*' value is present in the list, all origins will be allowed.
	// An origin may contain a wildcard (*) to replace 0 or more characters (i.e.: http://*.domain.com).
	// Only one wildcard can be used per origin
	AllowedOrigins []string `toml:"cors_allowed_origins"`

	// A list of non-simple headers the client is allowed to use with cross-domain requests
	AllowedHeaders []string `toml:"cors_allowed_headers"`

	// A list of methods the client is allowed to use with cross-domain requests
	AllowedMethods []string `toml:"cors_allowed_methods"`
}
