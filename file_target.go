package watch

import (
	"fmt"
	"io/ioutil"
)

type watchedFile struct {
	path     string
	failOpen bool
}

var _ Watched = &watchedFile{}

// FileTarget constructs a Watched wrapper for a file at a given path and allows
// selection of whether failing to access the file should result in an update
// signal.
func FileTarget(path string, failOpen bool) Watched {
	return watchedFile{path, failOpen}
}

func (wf watchedFile) FailOpen() bool { return wf.failOpen }
func (wf watchedFile) Content() ([]byte, error) {
	configContents, err := ioutil.ReadFile(wf.path)
	if err != nil {
		return nil, fmt.Errorf("Unable to read config file: %v", err)
	}
	return configContents, nil
}
