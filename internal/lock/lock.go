package lock

import (
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
	lockOwner, _ := l.lockFile.GetOwner()
	log.Debugf("acquiring lock for process %d", lockOwner.Pid)
	return l.lockFile.TryLock()
}

func (l *lock) Release() error {
	lockOwner, _ := l.lockFile.GetOwner()
	log.Debugf("releasing lock for process %d", lockOwner.Pid)
	return l.lockFile.Unlock()
}
