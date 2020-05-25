package roamer

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/go-sql-driver/mysql"
)

// ErrEnvironmentMissingConfig is returned when the environment is missing a roamer.toml file.
var ErrEnvironmentMissingConfig = errors.New("roamer: environment is missing roamer.toml file")

// ErrEnvironmentMissingLocalConfig is returned when the environment is missing a roamer.local.toml file.
var ErrEnvironmentMissingLocalConfig = errors.New("roamer: environment is missing roamer.local.toml file")

// ErrEnvironmentWasFile is returned when you provide a file as your environment path.
var ErrEnvironmentWasFile = errors.New("roamer: environment path is a file, not a folder! make sure you provide the *path* to your roamer.toml, not the actual file.")

// An Environment is the context in which roamer operates. It contains migrations and configuration data.
// Do not create this struct manually; use the NewEnvironment function instead.
type Environment struct {
	Config Config
	LocalConfig

	db     *sql.DB
	driver driver

	migrations     []Migration
	migrationsByID map[string]Migration

	basePath           string
	fullMigrationsPath string
}

// GetHistoryTableName gets the name of the table roamer is using to track history.
func (e *Environment) GetHistoryTableName() string {
	return tableNameRoamerHistory
}

// NewEnvironment creates a new environment with the given path.
func NewEnvironment(basePath string) (*Environment, error) {
	// validate the path
	envInfo, err := os.Stat(basePath)
	if err != nil {
		return nil, err
	}

	if !envInfo.IsDir() {
		return nil, ErrEnvironmentWasFile
	}

	env := Environment{
		Config:      DefaultConfig,
		LocalConfig: DefaultLocalConfig,

		basePath: basePath,
	}

	// get the config path and read it
	configPath := path.Join(basePath, "roamer.toml")
	configFile, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrEnvironmentMissingConfig
		}

		return nil, err
	}
	metadata, err := toml.DecodeReader(configFile, &env.Config)
	if err != nil {
		return nil, err
	}
	if len(metadata.Undecoded()) != 0 {
		return nil, UndecodedConfigError{"roamer.toml", metadata.Undecoded()}
	}

	// get the local config path and read it
	configLocalPath := path.Join(basePath, "roamer.local.toml")
	configLocalFile, err := os.Open(configLocalPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrEnvironmentMissingLocalConfig
		}

		return nil, err
	}
	metadata, err = toml.DecodeReader(configLocalFile, &env.LocalConfig)
	if err != nil {
		return nil, err
	}
	if len(metadata.Undecoded()) != 0 {
		return nil, UndecodedConfigError{"roamer.local.toml", metadata.Undecoded()}
	}

	env.fullMigrationsPath = path.Join(basePath, env.Config.MigrationDirectory)

	if env.LocalConfig.Database.Driver != "mysql" {
		return nil, fmt.Errorf("roamer: did not recognize driver name '%s'", env.LocalConfig.Database.Driver)
	}

	dsn := env.LocalConfig.Database.DSN

	if env.LocalConfig.Database.Driver == "mysql" {
		config, err := mysql.ParseDSN(dsn)
		if err != nil {
			return nil, err
		}

		config.MultiStatements = true

		env.driver = &driverMySQL{
			db:     env.db,
			config: config,
		}

		dsn = config.FormatDSN()
	}

	// try to connect to the database
	env.db, err = sql.Open(env.LocalConfig.Database.Driver, dsn)
	if err != nil {
		return nil, err
	}

	// test that it worked
	err = env.db.Ping()
	if err != nil {
		return nil, err
	}

	// set up the driver
	if env.LocalConfig.Database.Driver == "mysql" {
		config, err := mysql.ParseDSN(dsn)
		if err != nil {
			return nil, err
		}

		env.driver = &driverMySQL{
			db:     env.db,
			config: config,
		}
	}

	// scan the migrations directory
	migrationsDir, err := os.Open(env.fullMigrationsPath)
	err = env.db.Ping()
	if err != nil {
		return nil, err
	}

	filenames, err := migrationsDir.Readdirnames(0)
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
	for _, baseName := range baseNames {
		fullPath := path.Join(env.fullMigrationsPath, baseName)

		parts := strings.Split(baseName, "_")
		id := parts[0]

		_, existsAlready := env.migrationsByID[id]
		if existsAlready {
			return nil, fmt.Errorf("roamer: there are two migrations with ID %s", id)
		}

		downPath := fullPath + "_down.sql"
		upPath := fullPath + "_up.sql"

		// read the description from the down migration
		downFile, err := ioutil.ReadFile(downPath)
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

			downPath: downPath,
			upPath:   upPath,
		})
		env.migrationsByID[id] = env.migrations[len(env.migrations)-1]
	}

	return &env, nil
}
