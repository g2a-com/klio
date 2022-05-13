package manager

import (
	"fmt"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/dependency/registry"
	"github.com/g2a-com/klio/internal/lock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
)

//===========================  INDEX HANDLER MOCK ===========================

type mockIndexHandler struct {
	mock.Mock
}

func (m *mockIndexHandler) LoadDependencyIndex(filePath string) error {
	args := m.Called(filePath)
	return args.Error(0)
}

func (m *mockIndexHandler) SaveDependencyIndex() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockIndexHandler) GetEntries() []dependency.DependenciesIndexEntry {
	args := m.Called()
	return args.Get(0).([]dependency.DependenciesIndexEntry)
}

func (m *mockIndexHandler) SetEntries(entries []dependency.DependenciesIndexEntry) {
	m.Called(entries)
}

//=========================== REGISTRY MOCK ===========================

type mockRegistry struct {
	mock.Mock
}

func (r *mockRegistry) Update() error {
	_ = r.Called()
	return nil
}

func (r *mockRegistry) GetHighestBreaking(dep dependency.Dependency) (*registry.Entry, error) {
	args := r.Called(dep)
	return args.Get(0).(*registry.Entry), args.Error(1)
}

func (r *mockRegistry) GetHighestNonBreaking(dep dependency.Dependency) (*registry.Entry, error) {
	args := r.Called(dep)
	return args.Get(0).(*registry.Entry), args.Error(1)
}

func (r *mockRegistry) GetExactMatch(dep dependency.Dependency) (*registry.Entry, error) {
	args := r.Called(dep)
	return args.Get(0).(*registry.Entry), args.Error(1)
}

//=========================== REGISTRY HTTP SERVER STUB ===========================

type testHandler struct{}

func (th *testHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.URL.Path != fmt.Sprintf("/%s/%s.tar.gz", "registry/commands", dependencyName) {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//hard dependency in resources
	file, err := afero.NewOsFs().Open(fmt.Sprintf("%s.tar.gz", dependencyName))
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	all, err := io.ReadAll(file)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = rw.Write(all)
}

//=========================== LOCK MOCK ===========================

type AcquireLockError struct{}

func (e *AcquireLockError) Error() string { return "acquire lock error" }

type mockLock struct {
	mock.Mock
	FailingAcquire bool
}

func newMockLock(_ string) (lock.Lock, error) {
	ml := new(mockLock)
	ml.On("Acquire")
	ml.On("Release")
	return ml, nil
}
func newMockLockFailingToAcquire(_ string) (lock.Lock, error) {
	ml := &mockLock{FailingAcquire: true}
	ml.On("Acquire")
	ml.On("Release")
	return ml, nil
}

func (ml *mockLock) Acquire() error {
	_ = ml.Called()
	if ml.FailingAcquire {
		return &AcquireLockError{}
	}
	return nil
}

func (ml *mockLock) Release() error {
	_ = ml.Called()
	return nil
}

//=========================== LOCAL FS MOCK ===========================

var fsPaths = []string{
	"valid/artifactory/path",
	"some/path/nobody/cares/about",
}

func getMockFs() afero.Fs {
	fs := afero.NewMemMapFs()

	for _, p := range fsPaths {
		_ = fs.MkdirAll(p, 0x777)
	}

	return fs
}
