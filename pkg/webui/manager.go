package webui

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"

	"github.com/Masterminds/sprig"
	"github.com/ssoroka/slice"
)

var (
	ErrNoTemplate        = errors.New("at least one template is required")
	ErrAlreadyRegistered = errors.New("template already registered")
	ErrNotRegistered     = errors.New("template not registered")
	ErrInvalidTemplate   = errors.New("invalid template")
	ErrRenderFailure     = errors.New("could not render template")
)

// Manager is an interface that handles view templates
type Manager interface {
	// AddBase saves a file as a base file and will inject this
	// file in each render call
	AddBase(filePath string) error

	// Register creates a new template from a list of filepaths
	// The filepaths syntax depends on the manager implementation
	// It returns a key to identify the view and an error in case
	// it cannot build the template
	Register(filePaths ...string) (string, error)

	// Render writes a registered template using some values to a
	// writer
	Render(w io.Writer, key string, values any) error
}

// defaultManager is a concrete implementation of Manager
// it uses an FS object to resolve paths
type defaultManager struct {
	// fs represents the file system where the templates are
	// stored
	fs fs.FS

	// bases is the internal data store of the base
	// files
	bases []string

	// views is the internal data store of the registered
	// templates
	views map[string]*template.Template
}

// NewManager creates a new manager using the defaultManager
// implementation
func NewManager(fs fs.FS) (Manager, error) {
	return &defaultManager{
		fs:    fs,
		bases: []string{},
		views: map[string]*template.Template{},
	}, nil
}

func (m *defaultManager) AddBase(filePath string) error {
	if slice.Contains(m.bases, filePath) {
		return ErrAlreadyRegistered
	}

	m.bases = append(m.bases, filePath)

	return nil
}

// Register combines multiple paths from fs in a template
func (m *defaultManager) Register(filePaths ...string) (string, error) {
	if len(filePaths) == 0 {
		return "", ErrNoTemplate
	}

	key := filePaths[len(filePaths)-1]

	if _, ok := m.views[key]; ok {
		return "", ErrAlreadyRegistered
	}

	name := key
	if len(m.bases) >= 0 {
		name = m.bases[0]
	}

	tpl, err := template.
		New(name).
		Funcs(sprig.FuncMap()).
		ParseFS(m.fs, append(m.bases, filePaths...)...)

	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidTemplate, err)
	}

	m.views[key] = tpl

	return key, nil
}

// Render writes the registered templates and loads all values received
// under the Values object
func (m *defaultManager) Render(w io.Writer, fileName string, values any) error {
	if _, ok := m.views[fileName]; !ok {
		return ErrNotRegistered
	}

	if err := m.views[fileName].Execute(w, map[string]any{
		"Values": values,
	}); err != nil {
		return fmt.Errorf("%w: %v", ErrRenderFailure, err)
	}

	return nil
}
