package configformat

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/beevik/etree"

	"github.com/fnune/kyaraben/internal/model"
)

type xmlHandler struct{}

func (h *xmlHandler) Read(path string) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	result[""] = make(map[string]string)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return nil, err
	}

	root := doc.Root()
	if root == nil {
		return result, nil
	}

	h.readXMLElement(root, nil, result)
	return result, nil
}

func (h *xmlHandler) readXMLElement(elem *etree.Element, path []string, result map[string]map[string]string) {
	currentPath := append(path, elem.Tag)

	children := elem.ChildElements()
	if len(children) == 0 {
		text := elem.Text()
		section := SectionKey(currentPath[:len(currentPath)-1])
		if result[section] == nil {
			result[section] = make(map[string]string)
		}
		result[section][elem.Tag] = text
		return
	}

	for _, child := range children {
		h.readXMLElement(child, currentPath, result)
	}
}

func (h *xmlHandler) Apply(path string, entries []model.ConfigEntry) (ApplyResult, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return ApplyResult{}, fmt.Errorf("creating config directory: %w", err)
	}

	doc := etree.NewDocument()
	if data, err := os.ReadFile(path); err == nil {
		if err := doc.ReadFromBytes(data); err != nil {
			doc = etree.NewDocument()
		}
	}

	for _, entry := range entries {
		if entry.Unmanaged && hasXMLValue(doc, entry.Path) {
			continue
		}
		setXMLValue(doc, entry.Path, entry.Value)
	}

	doc.Indent(2)
	if err := doc.WriteToFile(path); err != nil {
		return ApplyResult{}, fmt.Errorf("writing XML file: %w", err)
	}

	hash, err := hashFile(path)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("hashing config file: %w", err)
	}

	return ApplyResult{Path: path, BaselineHash: hash}, nil
}

func setXMLValue(doc *etree.Document, path []string, value string) {
	if len(path) == 0 {
		return
	}

	root := doc.Root()
	if root == nil {
		root = doc.CreateElement(path[0])
		path = path[1:]
	} else if root.Tag != path[0] {
		root = doc.CreateElement(path[0])
		path = path[1:]
	} else {
		path = path[1:]
	}

	elem := root
	for i, key := range path {
		isLast := i == len(path)-1
		child := elem.SelectElement(key)
		if child == nil {
			child = elem.CreateElement(key)
		}
		if isLast {
			child.SetText(value)
		}
		elem = child
	}
}

func hasXMLValue(doc *etree.Document, path []string) bool {
	if len(path) == 0 {
		return false
	}

	root := doc.Root()
	if root == nil || root.Tag != path[0] {
		return false
	}

	elem := root
	for _, key := range path[1:] {
		child := elem.SelectElement(key)
		if child == nil {
			return false
		}
		elem = child
	}
	return true
}
