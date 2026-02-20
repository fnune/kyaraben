package configformat

import (
	"fmt"
	"path/filepath"

	"github.com/beevik/etree"
	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/model"
)

type xmlAttrHandler struct {
	fs vfs.FS
}

func (h *xmlAttrHandler) Read(path string) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	result[""] = make(map[string]string)

	data, err := h.fs.ReadFile(path)
	if err != nil {
		return result, nil
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return result, nil
	}

	for _, elem := range doc.ChildElements() {
		nameAttr := elem.SelectAttr("name")
		valueAttr := elem.SelectAttr("value")
		if nameAttr != nil && valueAttr != nil {
			result[""][nameAttr.Value] = valueAttr.Value
		}
	}

	return result, nil
}

func (h *xmlAttrHandler) Apply(path string, entries []model.ConfigEntry, _ []model.ManagedRegion) (ApplyResult, error) {
	if err := vfs.MkdirAll(h.fs, filepath.Dir(path), 0755); err != nil {
		return ApplyResult{}, fmt.Errorf("creating config directory: %w", err)
	}

	doc := etree.NewDocument()
	existing := make(map[string]*etree.Element)

	if data, err := h.fs.ReadFile(path); err == nil {
		if err := doc.ReadFromBytes(data); err == nil {
			for _, elem := range doc.ChildElements() {
				if nameAttr := elem.SelectAttr("name"); nameAttr != nil {
					existing[nameAttr.Value] = elem
				}
			}
		}
	}

	if doc.Root() == nil {
		doc.CreateProcInst("xml", `version="1.0"`)
	}

	for _, entry := range entries {
		if len(entry.Path) == 0 {
			continue
		}
		name := entry.Path[len(entry.Path)-1]

		if entry.DefaultOnly {
			if _, exists := existing[name]; exists {
				continue
			}
		}

		if elem, exists := existing[name]; exists {
			if attr := elem.SelectAttr("value"); attr != nil {
				attr.Value = entry.Value
			} else {
				elem.CreateAttr("value", entry.Value)
			}
		} else {
			elem := doc.CreateElement("string")
			elem.CreateAttr("name", name)
			elem.CreateAttr("value", entry.Value)
			existing[name] = elem
		}
	}

	doc.Indent(2)
	xmlBytes, err := doc.WriteToBytes()
	if err != nil {
		return ApplyResult{}, fmt.Errorf("encoding XML: %w", err)
	}

	if err := h.fs.WriteFile(path, xmlBytes, 0644); err != nil {
		return ApplyResult{}, fmt.Errorf("writing XML file: %w", err)
	}

	hash, err := hashFileWithFS(h.fs, path)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("hashing config file: %w", err)
	}

	return ApplyResult{Path: path, BaselineHash: hash, PatchHash: ComputePatchHash(entries)}, nil
}
