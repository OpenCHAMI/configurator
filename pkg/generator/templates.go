package generator

import (
	"bytes"
	"fmt"
	"os"

	"github.com/OpenCHAMI/configurator/pkg/util"
	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
)

type Template struct {
	Contents []byte `json:"contents"`
}

func (t *Template) LoadFromFile(path string) error {
	// skip loading template if path is a directory with no error
	if isDir, err := util.IsDirectory(path); err == nil && isDir {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to test if template path is directory: %w", err)
	}

	// try and read the contents of the file
	// NOTE: we don't care if this is actually a Jinja template
	// or not...at least for now.
	contents, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}
	t.Contents = contents
	return nil
}

func (t *Template) IsEmpty() bool {
	return len(t.Contents) <= 0
}

// Wrapper function to slightly abstract away some of the nuances with using gonja
// into a single function call. This function is *mostly* for convenience and
// simplication. If no paths are supplied, then no templates will be applied and
// there will be no output.
//
// The "FileList" returns a slice of byte arrays in the same order as the argument
// list supplied, but with the Jinja templating applied.
func ApplyTemplates(mappings Mappings, templates map[string]Template) (FileMap, error) {
	var (
		data    = exec.NewContext(mappings)
		outputs = FileMap{}
	)

	for path, template := range templates {
		// load jinja template from file
		t, err := gonja.FromBytes(template.Contents)
		if err != nil {
			return nil, fmt.Errorf("failed to read template from file: %w", err)
		}

		// execute/render jinja template
		b := bytes.Buffer{}
		if err = t.Execute(&b, data); err != nil {
			return nil, fmt.Errorf("failed to execute: %w", err)
		}
		outputs[path] = b.Bytes()
	}

	return outputs, nil
}

// Wrapper function similiar to "ApplyTemplates" but takes file paths as arguments.
// This function will load templates from a file instead of using file contents.
func ApplyTemplateFromFiles(mappings Mappings, paths ...string) (FileMap, error) {
	var (
		data    = exec.NewContext(mappings)
		outputs = FileMap{}
	)

	for _, path := range paths {
		// load jinja template from file
		t, err := gonja.FromFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read template from file: %w", err)
		}

		// execute/render jinja template
		b := bytes.Buffer{}
		if err = t.Execute(&b, data); err != nil {
			return nil, fmt.Errorf("failed to execute: %w", err)
		}
		outputs[path] = b.Bytes()
	}

	return outputs, nil
}
