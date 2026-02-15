package symlink

import (
	"errors"
	"os"
	"testing"

	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/model"
)

func TestCreatorCreateOnEmptyPath(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/target":          &vfst.Dir{Perm: 0755},
		"/target/file.txt": "content",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	creator := NewCreator(fs)
	err = creator.Create(model.SymlinkSpec{Source: "/link", Target: "/target"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	info, err := fs.Lstat("/link")
	if err != nil {
		t.Fatalf("failed to stat source: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("source is not a symlink")
	}

	if _, err := fs.Stat("/link/file.txt"); err != nil {
		t.Errorf("symlink does not resolve to target directory: %v", err)
	}
}

func TestCreatorCreateParentDirectories(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/target": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	creator := NewCreator(fs)
	err = creator.Create(model.SymlinkSpec{Source: "/deep/nested/link", Target: "/target"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if _, err := fs.Lstat("/deep/nested/link"); err != nil {
		t.Errorf("symlink was not created: %v", err)
	}
}

func TestCreatorUpdateWhenTargetChanges(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/old_target":         &vfst.Dir{Perm: 0755},
		"/old_target/old.txt": "old",
		"/new_target":         &vfst.Dir{Perm: 0755},
		"/new_target/new.txt": "new",
		"/link":               &vfst.Symlink{Target: "/old_target"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	creator := NewCreator(fs)
	err = creator.Create(model.SymlinkSpec{Source: "/link", Target: "/new_target"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if _, err := fs.Stat("/link/new.txt"); err != nil {
		t.Error("symlink should now resolve to new_target")
	}
	if _, err := fs.Stat("/link/old.txt"); err == nil {
		t.Error("symlink should not resolve to old_target")
	}
}

func TestCreatorNoopWhenTargetUnchanged(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/target":          &vfst.Dir{Perm: 0755},
		"/target/file.txt": "content",
		"/link":            &vfst.Symlink{Target: "/target"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	creator := NewCreator(fs)
	err = creator.Create(model.SymlinkSpec{Source: "/link", Target: "/target"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if _, err := fs.Stat("/link/file.txt"); err != nil {
		t.Errorf("symlink should still resolve to target: %v", err)
	}
}

func TestCreatorErrorOnNonEmptyDirectory(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/dir_with_files/file.txt": "data",
		"/target":                  &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	creator := NewCreator(fs)
	err = creator.Create(model.SymlinkSpec{Source: "/dir_with_files", Target: "/target"})
	if err == nil {
		t.Fatal("expected error when source directory is non-empty")
	}

	if _, statErr := fs.Stat("/dir_with_files/file.txt"); statErr != nil {
		t.Error("source directory should not be deleted when it contains files")
	}
}

func TestCreatorReplacesEmptyDirectory(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/empty_dir": &vfst.Dir{Perm: 0755},
		"/target":    &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	creator := NewCreator(fs)
	err = creator.Create(model.SymlinkSpec{Source: "/empty_dir", Target: "/target"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	info, err := fs.Lstat("/empty_dir")
	if err != nil {
		t.Fatalf("failed to stat source: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("source should be a symlink after replacing empty directory")
	}
}

func TestCreatorErrorOnRegularFile(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/regular_file": "data",
		"/target":       &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	creator := NewCreator(fs)
	err = creator.Create(model.SymlinkSpec{Source: "/regular_file", Target: "/target"})
	if err == nil {
		t.Fatal("expected error when source is a regular file")
	}
}

func TestCreateAllWithFakeCreator(t *testing.T) {
	specs := []model.SymlinkSpec{
		{Source: "/a/b", Target: "/x/y"},
		{Source: "/c/d", Target: "/z/w"},
	}

	fake := &FakeCreator{}
	err := CreateAll(fake, specs)
	if err != nil {
		t.Fatalf("CreateAll() error = %v", err)
	}

	if len(fake.Created) != 2 {
		t.Errorf("expected 2 symlinks created, got %d", len(fake.Created))
	}
	if fake.Created[0] != specs[0] {
		t.Errorf("first spec mismatch: got %v, want %v", fake.Created[0], specs[0])
	}
	if fake.Created[1] != specs[1] {
		t.Errorf("second spec mismatch: got %v, want %v", fake.Created[1], specs[1])
	}
}

func TestCreateAllStopsOnError(t *testing.T) {
	specs := []model.SymlinkSpec{
		{Source: "/a/b", Target: "/x/y"},
		{Source: "/c/d", Target: "/z/w"},
	}

	expectedErr := errors.New("test error")
	fake := &FakeCreator{Err: expectedErr}
	err := CreateAll(fake, specs)
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if len(fake.Created) != 0 {
		t.Errorf("expected no symlinks created on error, got %d", len(fake.Created))
	}
}

func TestRemoveSymlinkPreservesTarget(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/target/important_save.dat": "precious data",
		"/link":                      &vfst.Symlink{Target: "/target"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	if err := Remove(fs, "/link"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	if _, err := fs.Lstat("/link"); !os.IsNotExist(err) {
		t.Error("symlink should be removed")
	}

	if _, err := fs.Stat("/target"); err != nil {
		t.Errorf("target directory should still exist: %v", err)
	}
	if _, err := fs.Stat("/target/important_save.dat"); err != nil {
		t.Errorf("file inside target should still exist: %v", err)
	}
}

func TestRemoveRefusesNonSymlink(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/real_dir": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	err = Remove(fs, "/real_dir")
	if err == nil {
		t.Fatal("Remove() should error on non-symlink")
	}

	if _, err := fs.Stat("/real_dir"); err != nil {
		t.Error("directory should not be removed")
	}
}

func TestRemoveNonexistentIsNoop(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	err = Remove(fs, "/nonexistent/path/to/symlink")
	if err != nil {
		t.Errorf("Remove() on nonexistent path should succeed, got: %v", err)
	}
}
