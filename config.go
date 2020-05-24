package roamer

// A Config struct defines some configuration parameters for roamer.
type Config struct {
	// MigrationDirectory defines where the migrations directory is, relative to the location of the config file.
	MigrationDirectory string
}

// A LocalDatabaseConfig struct defines configuration parameters for the database connection.
type LocalDatabaseConfig struct {
	Driver string
	DSN    string
}

// A LocalConfig struct defines several local configuration parameters for roamer.
// This is mainly used for describing a database connection
type LocalConfig struct {
	Database LocalDatabaseConfig
}

// DefaultConfig contains the default configuration options, used when creating a new environment.
var DefaultConfig = Config{
	MigrationDirectory: "migrations/",
}

// DefaultLocalConfig contains the default configuration options, used when creating a new environment.
var DefaultLocalConfig = LocalConfig{
	Database: LocalDatabaseConfig{
		Driver: "mysql",
		DSN:    "user:password@tcp(localhost:3306)/dbname",
	},
}
