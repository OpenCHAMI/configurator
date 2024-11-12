package generator

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/OpenCHAMI/configurator/pkg/util"
	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
)

func LoadTemplate(path string) (Template, error) {
	// skip loading template if path is a directory with no error
	if isDir, err := util.IsDirectory(path); err == nil && isDir {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to test if template path is directory: %w", err)
	}

	// try and read the contents of the file
	// NOTE: we don't care if this is actually a Jinja template
	// or not...at least for now.
	return os.ReadFile(path)
}

func LoadTemplates(paths []string, opts ...util.Option) (map[string]Template, error) {
	var (
		templates = make(map[string]Template)
		params    = util.ToDict(opts...)
	)

	for _, path := range paths {
		err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
			// skip trying to load generator plugin if directory or error
			if info.IsDir() || err != nil {
				return nil
			}

			// load the contents of the template
			template, err := LoadTemplate(path)
			if err != nil {
				return fmt.Errorf("failed to load generator in directory '%s': %w", path, err)
			}

			// show the templates loaded if verbose flag is set
			if util.GetVerbose(params) {
				fmt.Printf("-- loaded tempalte '%s'\n", path)
			}

			// map each template by the path it was loaded from
			templates[path] = template
			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to walk directory: %w", err)
		}
	}

	return templates, nil
}

// Wrapper function to slightly abstract away some of the nuances with using gonja
// into a single function call. This function is *mostly* for convenience and
// simplication. If no paths are supplied, then no templates will be applied and
// there will be no output.
//
// The "FileList" returns a slice of byte arrays in the same order as the argument
// list supplied, but with the Jinja templating applied.
func ApplyTemplates(mappings Mappings, contents ...[]byte) (FileList, error) {
	var (
		data    = exec.NewContext(mappings)
		outputs = FileList{}
	)

	for _, b := range contents {
		// load jinja template from file
		t, err := gonja.FromBytes(b)
		if err != nil {
			return nil, fmt.Errorf("failed to read template from file: %w", err)
		}

		// execute/render jinja template
		b := bytes.Buffer{}
		if err = t.Execute(&b, data); err != nil {
			return nil, fmt.Errorf("failed to execute: %w", err)
		}
		outputs = append(outputs, b.Bytes())
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
