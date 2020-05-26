package roamer

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/BurntSushi/toml"
	"github.com/go-sql-driver/mysql"
)

// ErrEnvironmentMissingConfig is returned when the environment is missing a roamer.toml file.
var ErrEnvironmentMissingConfig = errors.New("roamer: environment is missing roamer.toml file")

// ErrEnvironmentMissingLocalConfig is returned when the environment is missing a local config file.
var ErrEnvironmentMissingLocalConfig = errors.New("roamer: environment is missing local config file")

// ErrEnvironmentWasFile is returned when you provide a file as your environment path.
var ErrEnvironmentWasFile = errors.New("roamer: environment path is a file, not a folder! make sure you provide the *path* to your roamer.toml, not the actual file.")

// ErrVersionTooOld is returned when the environment requires a newer version of roamer.
var ErrVersionTooOld = errors.New("roamer: this environment requires a newer version of roamer")

// An Environment is the context in which roamer operates. It contains migrations and configuration data.
// Do not create this struct manually; use the NewEnvironment function instead.
type Environment struct {
	Config
	LocalConfig

	db     *sql.DB
	driver driver

	migrations     []Migration
	migrationsByID map[string]Migration

	fs         http.FileSystem
	pathOnDisk string
}

// GetHistoryTableName gets the name of the table roamer is using to track history.
func (e *Environment) GetHistoryTableName() string {
	return tableNameRoamerHistory
}

func (e *Environment) readFile(filename string) ([]byte, error) {
	migrationFile, err := e.fs.Open(filename)
	if err != nil {
		return nil, err
	}
	migrationData, err := ioutil.ReadAll(migrationFile)
	migrationFile.Close()
	if err != nil {
		return nil, err
	}

	return migrationData, nil
}

// NewEnvironment creates a new environment, reading from the given config and http.FileSystem and using the given *sql.DB.
func NewEnvironment(config Config, localConfig LocalConfig, db *sql.DB, fs http.FileSystem) (*Environment, error) {
	env := Environment{
		Config:      config,
		LocalConfig: localConfig,

		db: db,

		fs: fs,
	}

	if env.Config.Environment.MinimumVersion != "" {
		currentVersion := getVersion()
		minimumVersion, err := version.NewVersion(env.Config.Environment.MinimumVersion)
		if err != nil {
			return nil, err
		}

		if minimumVersion.GreaterThan(currentVersion) {
			return nil, ErrVersionTooOld
		}
	}

	var err error

	if env.LocalConfig.Database.Driver != DriverTypeMySQL && env.LocalConfig.Database.Driver != DriverTypeSQLite3 {
		return nil, fmt.Errorf("roamer: did not recognize driver type '%s'", env.LocalConfig.Database.Driver)
	}
	if env.LocalConfig.Database.Driver == DriverTypeSQLite3 && !sqliteAvailable {
		return nil, errors.New("roamer: sqlite support not available")
	}

	// test that the db works
	err = env.db.Ping()
	if err != nil {
		return nil, err
	}

	// set up the driver
	if env.LocalConfig.Database.Driver == DriverTypeMySQL {
		env.driver = &driverMySQL{
			db: env.db,
		}
	} else if env.LocalConfig.Database.Driver == DriverTypeSQLite3 {
		env.driver = &driverSQLite{
			db: env.db,
		}
	}

	// scan the migrations directory
	migrationsDir, err := fs.Open("")
	if err != nil {
		return nil, err
	}

	dirEntries, err := migrationsDir.Readdir(0)
	if err != nil {
		return nil, err
	}

	filenames := []string{}
	for _, dirEntry := range dirEntries {
		filenames = append(filenames, dirEntry.Name())
	}

	sort.Strings(filenames)

	baseNames := []string{}
	for _, filename := range filenames {
		if strings.HasSuffix(filename, "_down.sql") {
			baseName := strings.Replace(filename, "_down.sql", "", -1)
			baseNames = append(baseNames, baseName)
		} else if strings.HasSuffix(filename, "_up.sql") {
			baseName := strings.Replace(filename, "_up.sql", "", -1)

			// we need the matching down migration to exist
			exists := false
			for _, existingBaseName := range baseNames {
				if existingBaseName == baseName {
					exists = true
					break
				}
			}

			if !exists {
				return nil, fmt.Errorf("roamer: migration file '%s_down.sql' did not have matching up migration", baseName)
			}
		} else {
			return nil, fmt.Errorf("roamer: migration file '%s' did not end in recognized suffixes '_down.sql' or '_up.sql'", filename)
		}
	}

	env.migrations = []Migration{}
	env.migrationsByID = map[string]Migration{}
	for i, baseName := range baseNames {
		parts := strings.Split(baseName, "_")
		id := parts[0]

		_, existsAlready := env.migrationsByID[id]
		if existsAlready {
			return nil, fmt.Errorf("roamer: there are two migrations with ID %s", id)
		}

		downPath := baseName + "_down.sql"
		upPath := baseName + "_up.sql"

		// read the description from the down migration
		downFile, err := env.readFile(downPath)
		if err != nil {
			return nil, err
		}
		matches := reMigrationDescription.FindAllSubmatch(downFile, -1)
		if len(matches) == 0 {
			return nil, fmt.Errorf("roamer: migration file '%s' is missing a description line", downPath)
		}
		if len(matches) > 1 {
			return nil, fmt.Errorf("roamer: migration file '%s' has too many description lines", downPath)
		}

		description := string(matches[0][1])

		env.migrations = append(env.migrations, Migration{
			ID:          id,
			Description: description,

			Index: i,

			downPath: downPath,
			upPath:   upPath,
		})
		env.migrationsByID[id] = env.migrations[len(env.migrations)-1]
	}

	return &env, nil
}

// NewEnvironmentFromDisk creates a new environment with the given path.
func NewEnvironmentFromDisk(basePath string, localConfigName string) (*Environment, error) {
	// validate the path
	envInfo, err := os.Stat(basePath)
	if err != nil {
		return nil, err
	}

	if !envInfo.IsDir() {
		return nil, ErrEnvironmentWasFile
	}

	config := DefaultConfig
	localConfig := DefaultLocalConfig

	// get the config path and read it
	configPath := path.Join(basePath, "roamer.toml")
	configFile, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrEnvironmentMissingConfig
		}

		return nil, err
	}
	metadata, err := toml.DecodeReader(configFile, &config)
	if err != nil {
		return nil, err
	}
	if len(metadata.Undecoded()) != 0 {
		return nil, UndecodedConfigError{"roamer.toml", metadata.Undecoded()}
	}

	// get the local config path and read it
	configLocalPath := path.Join(basePath, "roamer."+localConfigName+".toml")
	configLocalFile, err := os.Open(configLocalPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrEnvironmentMissingLocalConfig
		}

		return nil, err
	}
	metadata, err = toml.DecodeReader(configLocalFile, &localConfig)
	if err != nil {
		return nil, err
	}
	if len(metadata.Undecoded()) != 0 {
		return nil, UndecodedConfigError{"roamer.local.toml", metadata.Undecoded()}
	}

	fullMigrationsPath := path.Join(basePath, config.Environment.MigrationDirectory)

	if localConfig.Database.Driver != DriverTypeMySQL && localConfig.Database.Driver != DriverTypeSQLite3 {
		return nil, fmt.Errorf("roamer: did not recognize driver type '%s'", localConfig.Database.Driver)
	}
	if localConfig.Database.Driver == DriverTypeSQLite3 && !sqliteAvailable {
		return nil, errors.New("roamer: sqlite support not available")
	}

	// connect to the db
	dsn := localConfig.Database.DSN

	if localConfig.Database.Driver == DriverTypeMySQL {
		config, err := mysql.ParseDSN(dsn)
		if err != nil {
			return nil, err
		}

		config.MultiStatements = true

		dsn = config.FormatDSN()
	}

	// try to connect to the database
	db, err := sql.Open(string(localConfig.Database.Driver), dsn)
	if err != nil {
		return nil, err
	}

	env, err := NewEnvironment(config, localConfig, db, http.Dir(fullMigrationsPath))
	if err != nil {
		return nil, err
	}

	env.pathOnDisk = fullMigrationsPath

	return env, nil
}
