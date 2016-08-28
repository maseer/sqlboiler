// Package sqlboiler has types and methods useful for generating code that
// acts as a fully dynamic ORM might.
package main

import (
	"go/build"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/vattle/sqlboiler/bdb"
	"github.com/vattle/sqlboiler/bdb/drivers"
)

const (
	templatesDirectory          = "templates"
	templatesSingletonDirectory = "templates/singleton"

	templatesTestDirectory          = "templates_test"
	templatesSingletonTestDirectory = "templates_test/singleton"

	templatesTestMainDirectory = "templates_test/main_test"
)

// State holds the global data needed by most pieces to run
type State struct {
	Config *Config

	Driver bdb.Interface
	Tables []bdb.Table

	Templates              *templateList
	TestTemplates          *templateList
	SingletonTemplates     *templateList
	SingletonTestTemplates *templateList

	TestMainTemplate *template.Template
}

// New creates a new state based off of the config
func New(config *Config) (*State, error) {
	s := &State{
		Config: config,
	}

	err := s.initDriver(config.DriverName)
	if err != nil {
		return nil, err
	}

	// Connect to the driver database
	if err = s.Driver.Open(); err != nil {
		return nil, errors.Wrap(err, "unable to connect to the database")
	}

	err = s.initTables(config.ExcludeTables)
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize tables")
	}

	err = s.initOutFolder()
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize the output folder")
	}

	err = s.initTemplates()
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize templates")
	}

	return s, nil
}

// Run executes the sqlboiler templates and outputs them to files based on the
// state given.
func (s *State) Run(includeTests bool) error {
	singletonData := &templateData{
		Tables:          s.Tables,
		DriverName:      s.Config.DriverName,
		UseLastInsertID: s.Driver.UseLastInsertID(),
		PkgName:         s.Config.PkgName,
		NoHooks:         s.Config.NoHooks,

		StringFuncs: templateStringMappers,
	}

	if err := generateSingletonOutput(s, singletonData); err != nil {
		return errors.Wrap(err, "singleton template output")
	}

	if includeTests {
		if err := generateTestMainOutput(s, singletonData); err != nil {
			return errors.Wrap(err, "unable to generate TestMain output")
		}

		if err := generateSingletonTestOutput(s, singletonData); err != nil {
			return errors.Wrap(err, "unable to generate singleton test template output")
		}
	}

	for _, table := range s.Tables {
		if table.IsJoinTable {
			continue
		}

		data := &templateData{
			Tables:          s.Tables,
			Table:           table,
			DriverName:      s.Config.DriverName,
			UseLastInsertID: s.Driver.UseLastInsertID(),
			PkgName:         s.Config.PkgName,
			NoHooks:         s.Config.NoHooks,

			StringFuncs: templateStringMappers,
		}

		// Generate the regular templates
		if err := generateOutput(s, data); err != nil {
			return errors.Wrap(err, "unable to generate output")
		}

		// Generate the test templates
		if includeTests {
			if err := generateTestOutput(s, data); err != nil {
				return errors.Wrap(err, "unable to generate test output")
			}
		}
	}

	return nil
}

// Cleanup closes any resources that must be closed
func (s *State) Cleanup() error {
	s.Driver.Close()
	return nil
}

// initTemplates loads all template folders into the state object.
func (s *State) initTemplates() error {
	var err error

	basePath, err := getBasePath(s.Config.BaseDir)
	if err != nil {
		return err
	}

	s.Templates, err = loadTemplates(filepath.Join(basePath, templatesDirectory))
	if err != nil {
		return err
	}

	s.SingletonTemplates, err = loadTemplates(filepath.Join(basePath, templatesSingletonDirectory))
	if err != nil {
		return err
	}

	s.TestTemplates, err = loadTemplates(filepath.Join(basePath, templatesTestDirectory))
	if err != nil {
		return err
	}

	s.SingletonTestTemplates, err = loadTemplates(filepath.Join(basePath, templatesSingletonTestDirectory))
	if err != nil {
		return err
	}

	s.TestMainTemplate, err = loadTemplate(filepath.Join(basePath, templatesTestMainDirectory), s.Config.DriverName+"_main.tpl")
	if err != nil {
		return err
	}

	return nil
}

var basePackage = "github.com/vattle/sqlboiler"

func getBasePath(baseDirConfig string) (string, error) {
	if len(baseDirConfig) > 0 {
		return baseDirConfig, nil
	}

	p, _ := build.Default.Import(basePackage, "", build.FindOnly)
	if p != nil && len(p.Dir) > 0 {
		return p.Dir, nil
	}

	return os.Getwd()
}

// initDriver attempts to set the state Interface based off the passed in
// driver flag value. If an invalid flag string is provided an error is returned.
func (s *State) initDriver(driverName string) error {
	// Create a driver based off driver flag
	switch driverName {
	case "postgres":
		s.Driver = drivers.NewPostgresDriver(
			s.Config.Postgres.User,
			s.Config.Postgres.Pass,
			s.Config.Postgres.DBName,
			s.Config.Postgres.Host,
			s.Config.Postgres.Port,
			s.Config.Postgres.SSLMode,
		)
	case "mock":
		s.Driver = &drivers.MockDriver{}
	}

	if s.Driver == nil {
		return errors.New("An invalid driver name was provided")
	}

	return nil
}

// initTables retrieves all "public" schema table names from the database.
func (s *State) initTables(exclude []string) error {
	var err error
	s.Tables, err = bdb.Tables(s.Driver, exclude...)
	if err != nil {
		return errors.Wrap(err, "unable to fetch table data")
	}

	if len(s.Tables) == 0 {
		return errors.New("no tables found in database")
	}

	if err := checkPKeys(s.Tables); err != nil {
		return err
	}

	return nil
}

// initOutFolder creates the folder that will hold the generated output.
func (s *State) initOutFolder() error {
	return os.MkdirAll(s.Config.OutFolder, os.ModePerm)
}

// checkPKeys ensures every table has a primary key column
func checkPKeys(tables []bdb.Table) error {
	var missingPkey []string
	for _, t := range tables {
		if t.PKey == nil {
			missingPkey = append(missingPkey, t.Name)
		}
	}

	if len(missingPkey) != 0 {
		return errors.Errorf("primary key missing in tables (%s)", strings.Join(missingPkey, ", "))
	}

	return nil
}
