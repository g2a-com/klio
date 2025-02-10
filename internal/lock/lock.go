package lock

import (
	"fmt"

	"github.com/g2a-com/klio/internal/log"
	"github.com/nightlyone/lockfile"
)

type Lock interface {
	Acquire() error
	Release() error
}

type lock struct {
	lockFile lockfile.Lockfile
}

func New(lockPath string) (Lock, error) {
	l, err := lockfile.New(lockPath)
	if err != nil {
		return nil, err
	}
	return &lock{lockFile: l}, nil
}

func (l *lock) Acquire() error {
	err := l.lockFile.TryLock()
	if err != nil {
		return err
	}
	lockOwner, err := l.lockFile.GetOwner()
	if err != nil {
		return err
	}
	log.Debugf("acquiring lock for process %d", lockOwner.Pid)
	return nil
}

func (l *lock) Release() error {
	lockOwner, err := l.lockFile.GetOwner()
	if err != nil {
		return fmt.Errorf("failed releasing a lock: %v", err)
	}
	log.Debugf("releasing lock for process %d", lockOwner.Pid)
	return l.lockFile.Unlock()
}
