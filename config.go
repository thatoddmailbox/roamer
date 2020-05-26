package roamer

// An EnvironmentConfig struct defines configuration parameters related to the environment's setup.
type EnvironmentConfig struct {
	// MigrationDirectory defines where the migrations directory is, relative to the location of the config file.
	MigrationDirectory string

	// MinimumVersion defines the minimum version of roamer required for this environment.
	MinimumVersion string
}

// A Config struct defines some configuration parameters for roamer.
type Config struct {
	Environment EnvironmentConfig
}

// A LocalDatabaseConfig struct defines configuration parameters for the database connection.
type LocalDatabaseConfig struct {
	Driver DriverType
	DSN    string
}

// A LocalConfig struct defines several local configuration parameters for roamer.
// This is mainly used for describing a database connection
type LocalConfig struct {
	Database LocalDatabaseConfig
}

// DefaultConfig contains the default configuration options, used when creating a new environment.
var DefaultConfig = Config{
	Environment: EnvironmentConfig{
		MigrationDirectory: "migrations/",
		MinimumVersion:     GetVersionString(),
	},
}

// DefaultLocalConfig contains the default configuration options, used when creating a new environment.
var DefaultLocalConfig = LocalConfig{
	Database: LocalDatabaseConfig{
		Driver: DriverTypeMySQL,
		DSN:    "user:password@tcp(localhost:3306)/dbname",
	},
}
