package lock

import (
	"github.com/g2a-com/klio/internal/log"

	"github.com/nightlyone/lockfile"
)

func Acquire(path string) error {
	log.Debugf("Acquiring lock: %s", path)
	l, err := lockfile.New(path)
	if err != nil {
		return err
	}
	return l.TryLock()
}

func Release(path string) error {
	log.Debugf("Releasing lock: %s", path)
	l, err := lockfile.New(path)
	if err != nil {
		return err
	}
	return l.Unlock()
}
