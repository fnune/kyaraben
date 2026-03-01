package importscanner

import "github.com/fnune/kyaraben/internal/model"

type ImportMode string

const (
	ImportModeCopy       ImportMode = "copy"
	ImportModeReorganize ImportMode = "reorganize"
)

type DataType string

const (
	DataTypeROMs        DataType = "roms"
	DataTypeBIOS        DataType = "bios"
	DataTypeSaves       DataType = "saves"
	DataTypeStates      DataType = "states"
	DataTypeScreenshots DataType = "screenshots"
	DataTypeGamelists   DataType = "gamelists"
	DataTypeMedia       DataType = "media"
)

type ImportReport struct {
	SourcePath   string           `json:"sourcePath"`
	ESDEPath     string           `json:"esdePath,omitempty"`
	KyarabenPath string           `json:"kyarabenPath"`
	Mode         ImportMode       `json:"mode"`
	Systems      []SystemReport   `json:"systems"`
	Frontends    []FrontendReport `json:"frontends,omitempty"`
	Unclassified []FileInfo       `json:"unclassified,omitempty"`
	Summary      DiffSummary      `json:"summary"`
}

type SystemReport struct {
	System     model.SystemID   `json:"system"`
	SystemName string           `json:"systemName"`
	SystemData []DataComparison `json:"systemData"`
	Emulators  []EmulatorReport `json:"emulators"`
}

type EmulatorReport struct {
	Emulator     model.EmulatorID `json:"emulator"`
	EmulatorName string           `json:"emulatorName"`
	EmulatorData []DataComparison `json:"emulatorData"`
}

type FrontendReport struct {
	Frontend     model.FrontendID `json:"frontend"`
	FrontendName string           `json:"frontendName"`
	FrontendData []DataComparison `json:"frontendData"`
}

type DataComparison struct {
	DataType DataType   `json:"dataType"`
	Source   FolderInfo `json:"source"`
	Kyaraben FolderInfo `json:"kyaraben"`
	Diff     DiffInfo   `json:"diff"`
	Notes    []string   `json:"notes,omitempty"`
}

type FolderInfo struct {
	Path      string       `json:"path"`
	FileCount int          `json:"fileCount"`
	TotalSize int64        `json:"totalSize"`
	Symlink   *SymlinkInfo `json:"symlink,omitempty"`
	Exists    bool         `json:"exists"`
	IsFlat    bool         `json:"isFlat,omitempty"`
}

type SymlinkInfo struct {
	Target string `json:"target"`
	Intact bool   `json:"intact"`
}

type DiffInfo struct {
	OnlyInSource   []FileInfo `json:"onlyInSource,omitempty"`
	OnlyInKyaraben []FileInfo `json:"onlyInKyaraben,omitempty"`
	SourceDelta    int64      `json:"sourceDelta"`
	KyarabenDelta  int64      `json:"kyarabenDelta"`
}

type FileInfo struct {
	RelPath string `json:"relPath"`
	Size    int64  `json:"size"`
}

type DiffSummary struct {
	TotalOnlyInSource   int64 `json:"totalOnlyInSource"`
	TotalOnlyInKyaraben int64 `json:"totalOnlyInKyaraben"`
}
