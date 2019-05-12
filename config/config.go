package config

// Config struct
type Config struct {
	CacheDir   string
	ConfigDir  string
	ConfigFile string
}

// Files represents the config files / dirs
var files Config

func init() {
	files = Config{}
}

// SetConfig is a setter for config
func SetConfig(config *Config) {
	files = *config
}

// GetConfig is a getter for config
func GetConfig() *Config {
	return &files
}
