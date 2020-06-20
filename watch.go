package watch

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"time"
)

// Watch exposes two methods for determing when a file's content has changed.
type Watch interface {
	// Updated returns whether or not the file has changed since last called.
	// It always returns true on the first call and is not safe for concurrent
	// use.
	//
	// It does not interact or conflict with OnInterval signals.
	Updated() (bool, error)

	// OnInterval starts a go routine that will emit to a channel when the
	// watched files changes. This will emit at most once per change and checks
	// for updates as specified by the provided interval duration.
	OnInterval(interval time.Duration) (<-chan struct{}, context.CancelFunc)
}

// Watched is an interface representing an object that can be observed for
// change
type Watched interface {
	// If we hit an error trying to get content this controls behavior of the
	// Watch. Return true if the watch should treat content fetch errors as
	// an update; false if it should not.
	FailOpen() bool

	// Content returns the content that should be compared or an error if it
	// could not be fetched.
	Content() ([]byte, error)
}

type watcher struct {
	target   Watched
	lastHash []byte
}

// File constructs a Watch for a file.
func File(path string) Watch {
	return NewWatched(FileTarget(path, true))
}

// NewWatched constructs a watch for a target that can be reified into a series
// of bytes.
func NewWatched(target Watched) Watch {
	return &watcher{target, nil}
}

func (w *watcher) Updated() (bool, error) {
	newHash, diff, err := w.targetDiff(w.lastHash)
	if err == nil {
		w.lastHash = newHash
	}
	return diff, err
}

func (w *watcher) targetDiff(token []byte) ([]byte, bool, error) {
	content, err := w.target.Content()
	if err != nil {
		return nil, w.target.FailOpen(), fmt.Errorf("Unable to get target content: %v", err.Error())
	}

	contentSum := md5.Sum(content)
	// slicify for go
	hash := contentSum[:]

	if byteSliceMatch(hash, token) {
		return token, false, nil
	}
	return hash, true, nil
}

func (w *watcher) OnInterval(
	interval time.Duration,
) (<-chan struct{}, context.CancelFunc) {
	ticker := time.NewTicker(interval)

	ch := make(chan struct{})
	ctx, cancelFn := context.WithCancel(context.Background())

	go func(ticker *time.Ticker, done <-chan struct{}, updatedCh chan<- struct{}) {
		var lastHash []byte
		cancelled := false
		for !cancelled {
			select {
			case <-done:
				cancelled = true

			case <-ticker.C:
				if checkedHash, updated, err := w.targetDiff(lastHash); err == nil {
					if updated {
						lastHash = checkedHash
						select {
						case updatedCh <- struct{}{}:
						case <-done:
							cancelled = true
						}
					}
				}
			}
		}
		close(updatedCh)
	}(ticker, ctx.Done(), ch)

	return ch, cancelFn
}

type watchedFile struct {
	path     string
	failOpen bool
}

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

func byteSliceMatch(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
