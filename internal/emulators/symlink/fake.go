package symlink

import "github.com/fnune/kyaraben/internal/model"

type FakeCreator struct {
	Created []model.SymlinkSpec
	Err     error
}

func (f *FakeCreator) Create(spec model.SymlinkSpec) error {
	if f.Err != nil {
		return f.Err
	}
	f.Created = append(f.Created, spec)
	return nil
}
