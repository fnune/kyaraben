package testutil

import (
	"testing"

	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/model"
)

func NewTestFS(t *testing.T, root map[string]any) vfs.FS {
	t.Helper()
	fs, cleanup, err := vfst.NewTestFS(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	return fs
}

type FakeResolver struct {
	ConfigDir string
	DataDir   string
	HomeDir   string
}

func (r FakeResolver) UserConfigDir() (string, error) { return r.ConfigDir, nil }
func (r FakeResolver) UserDataDir() (string, error)   { return r.DataDir, nil }
func (r FakeResolver) UserHomeDir() (string, error)   { return r.HomeDir, nil }

var _ model.BaseDirResolver = FakeResolver{}
