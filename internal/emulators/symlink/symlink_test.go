package symlink

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

func TestOSCreatorCreateOnEmptyPath(t *testing.T) {
	tmpDir := t.TempDir()
	source := filepath.Join(tmpDir, "link")
	target := filepath.Join(tmpDir, "target")

	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}

	creator := OSCreator{}
	err := creator.Create(model.SymlinkSpec{Source: source, Target: target})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	info, err := os.Lstat(source)
	if err != nil {
		t.Fatalf("failed to stat source: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("source is not a symlink")
	}

	resolvedTarget, err := os.Readlink(source)
	if err != nil {
		t.Fatalf("Readlink() error = %v", err)
	}
	if resolvedTarget != target {
		t.Errorf("symlink points to %s, want %s", resolvedTarget, target)
	}
}

func TestOSCreatorCreateParentDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	source := filepath.Join(tmpDir, "deep", "nested", "link")
	target := filepath.Join(tmpDir, "target")

	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}

	creator := OSCreator{}
	err := creator.Create(model.SymlinkSpec{Source: source, Target: target})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if _, err := os.Lstat(source); err != nil {
		t.Errorf("symlink was not created: %v", err)
	}
}

func TestOSCreatorUpdateWhenTargetChanges(t *testing.T) {
	tmpDir := t.TempDir()
	source := filepath.Join(tmpDir, "link")
	oldTarget := filepath.Join(tmpDir, "old_target")
	newTarget := filepath.Join(tmpDir, "new_target")

	if err := os.MkdirAll(oldTarget, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(newTarget, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(oldTarget, source); err != nil {
		t.Fatal(err)
	}

	creator := OSCreator{}
	err := creator.Create(model.SymlinkSpec{Source: source, Target: newTarget})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	resolvedTarget, err := os.Readlink(source)
	if err != nil {
		t.Fatalf("Readlink() error = %v", err)
	}
	if resolvedTarget != newTarget {
		t.Errorf("symlink points to %s, want %s", resolvedTarget, newTarget)
	}
}

func TestOSCreatorNoopWhenTargetUnchanged(t *testing.T) {
	tmpDir := t.TempDir()
	source := filepath.Join(tmpDir, "link")
	target := filepath.Join(tmpDir, "target")

	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, source); err != nil {
		t.Fatal(err)
	}

	creator := OSCreator{}
	err := creator.Create(model.SymlinkSpec{Source: source, Target: target})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	resolvedTarget, err := os.Readlink(source)
	if err != nil {
		t.Fatalf("Readlink() error = %v", err)
	}
	if resolvedTarget != target {
		t.Errorf("symlink points to %s, want %s", resolvedTarget, target)
	}
}

func TestOSCreatorErrorOnNonEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	source := filepath.Join(tmpDir, "dir_with_files")
	target := filepath.Join(tmpDir, "target")

	if err := os.MkdirAll(source, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "file.txt"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}

	creator := OSCreator{}
	err := creator.Create(model.SymlinkSpec{Source: source, Target: target})
	if err == nil {
		t.Fatal("expected error when source directory is non-empty")
	}

	if _, statErr := os.Stat(filepath.Join(source, "file.txt")); statErr != nil {
		t.Error("source directory should not be deleted when it contains files")
	}
}

func TestOSCreatorReplacesEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	source := filepath.Join(tmpDir, "empty_dir")
	target := filepath.Join(tmpDir, "target")

	if err := os.MkdirAll(source, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}

	creator := OSCreator{}
	err := creator.Create(model.SymlinkSpec{Source: source, Target: target})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	info, err := os.Lstat(source)
	if err != nil {
		t.Fatalf("failed to stat source: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("source should be a symlink after replacing empty directory")
	}
}

func TestOSCreatorErrorOnRegularFile(t *testing.T) {
	tmpDir := t.TempDir()
	source := filepath.Join(tmpDir, "regular_file")
	target := filepath.Join(tmpDir, "target")

	if err := os.WriteFile(source, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}

	creator := OSCreator{}
	err := creator.Create(model.SymlinkSpec{Source: source, Target: target})
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
