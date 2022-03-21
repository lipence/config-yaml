package yaml

import (
	"fmt"
	"regexp"

	"github.com/lipence/config"
	"github.com/lipence/gabs-yaml/v2"
)

const (
	Name            = "yaml"
	configPathRegex = `.*\.(yaml|yml)`
)

type Loader struct{}

func (l *Loader) Type() string {
	return Name
}

func (l *Loader) PathPattern() *regexp.Regexp {
	return regexp.MustCompile(configPathRegex)
}

func (l *Loader) AllowDir() bool {
	return false
}

func (l *Loader) Load(path string, files map[string][]byte) (val config.Value, err error) {
	var result *gabs.Container
	if content, ok := files[path]; !ok {
		return nil, fmt.Errorf("%w (path: %s)", config.ErrPathNotFound, path)
	} else if result, err = gabs.ParseYAML(content); err != nil {
		return nil, fmt.Errorf("%w (path: %s)", err, path)
	}
	return &value{result: result}, nil
}

func (l *Loader) Clear() {}
