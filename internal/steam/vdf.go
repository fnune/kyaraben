package steam

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
)

const (
	typeMap    byte = 0x00
	typeString byte = 0x01
	typeInt32  byte = 0x02
	typeEnd    byte = 0x08
)

type Shortcut struct {
	AppID               uint32
	AppName             string
	Exe                 string
	StartDir            string
	Icon                string
	ShortcutPath        string
	LaunchOptions       string
	IsHidden            uint32
	AllowDesktopConfig  uint32
	AllowOverlay        uint32
	OpenVR              uint32
	Devkit              uint32
	DevkitGameID        string
	DevkitOverrideAppID uint32
	LastPlayTime        uint32
	FlatpakAppID        string
	Tags                []string
}

func ParseShortcuts(r io.Reader) ([]Shortcut, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading data: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	p := &parser{data: data}
	return p.parseRoot()
}

func WriteShortcuts(w io.Writer, shortcuts []Shortcut) error {
	buf := &bytes.Buffer{}

	if err := buf.WriteByte(typeMap); err != nil {
		return err
	}
	if err := writeString(buf, "shortcuts"); err != nil {
		return err
	}

	for i, s := range shortcuts {
		if err := buf.WriteByte(typeMap); err != nil {
			return err
		}
		if err := writeString(buf, fmt.Sprintf("%d", i)); err != nil {
			return err
		}

		if err := writeInt32Field(buf, "appid", s.AppID); err != nil {
			return err
		}
		if err := writeStringField(buf, "AppName", s.AppName); err != nil {
			return err
		}
		if err := writeStringField(buf, "Exe", quoteExe(s.Exe)); err != nil {
			return err
		}
		if err := writeStringField(buf, "StartDir", s.StartDir); err != nil {
			return err
		}
		if err := writeStringField(buf, "icon", s.Icon); err != nil {
			return err
		}
		if err := writeStringField(buf, "ShortcutPath", s.ShortcutPath); err != nil {
			return err
		}
		if err := writeStringField(buf, "LaunchOptions", s.LaunchOptions); err != nil {
			return err
		}
		if err := writeInt32Field(buf, "IsHidden", s.IsHidden); err != nil {
			return err
		}
		if err := writeInt32Field(buf, "AllowDesktopConfig", s.AllowDesktopConfig); err != nil {
			return err
		}
		if err := writeInt32Field(buf, "AllowOverlay", s.AllowOverlay); err != nil {
			return err
		}
		if err := writeInt32Field(buf, "OpenVR", s.OpenVR); err != nil {
			return err
		}
		if err := writeInt32Field(buf, "Devkit", s.Devkit); err != nil {
			return err
		}
		if err := writeStringField(buf, "DevkitGameID", s.DevkitGameID); err != nil {
			return err
		}
		if err := writeInt32Field(buf, "DevkitOverrideAppID", s.DevkitOverrideAppID); err != nil {
			return err
		}
		if err := writeInt32Field(buf, "LastPlayTime", s.LastPlayTime); err != nil {
			return err
		}
		if err := writeStringField(buf, "FlatpakAppID", s.FlatpakAppID); err != nil {
			return err
		}
		if err := writeStringField(buf, "sortas", ""); err != nil {
			return err
		}

		if err := buf.WriteByte(typeMap); err != nil {
			return err
		}
		if err := writeString(buf, "tags"); err != nil {
			return err
		}
		for j, tag := range s.Tags {
			if err := writeStringField(buf, fmt.Sprintf("%d", j), tag); err != nil {
				return err
			}
		}
		if err := buf.WriteByte(typeEnd); err != nil {
			return err
		}

		if err := buf.WriteByte(typeEnd); err != nil {
			return err
		}
	}

	if err := buf.WriteByte(typeEnd); err != nil {
		return err
	}
	if err := buf.WriteByte(typeEnd); err != nil {
		return err
	}

	_, err := w.Write(buf.Bytes())
	return err
}

func GenerateAppID(exe, appName string) uint32 {
	input := exe + appName
	crc := crc32.ChecksumIEEE([]byte(input))
	return crc | 0x80000000
}

type parser struct {
	data []byte
	pos  int
}

func (p *parser) parseRoot() ([]Shortcut, error) {
	if p.pos >= len(p.data) {
		return nil, nil
	}

	typ := p.data[p.pos]
	p.pos++

	if typ != typeMap {
		return nil, fmt.Errorf("expected map at root, got type %02x", typ)
	}

	name, err := p.readString()
	if err != nil {
		return nil, fmt.Errorf("reading root map name: %w", err)
	}

	if name != "shortcuts" {
		return nil, fmt.Errorf("expected 'shortcuts' root, got %q", name)
	}

	return p.parseShortcutsList()
}

func (p *parser) parseShortcutsList() ([]Shortcut, error) {
	var shortcuts []Shortcut

	for {
		if p.pos >= len(p.data) {
			return nil, errors.New("unexpected end of data")
		}

		typ := p.data[p.pos]
		if typ == typeEnd {
			p.pos++
			break
		}

		if typ != typeMap {
			return nil, fmt.Errorf("expected map for shortcut entry, got type %02x", typ)
		}
		p.pos++

		_, err := p.readString()
		if err != nil {
			return nil, fmt.Errorf("reading shortcut index: %w", err)
		}

		shortcut, err := p.parseShortcut()
		if err != nil {
			return nil, fmt.Errorf("parsing shortcut: %w", err)
		}

		shortcuts = append(shortcuts, shortcut)
	}

	return shortcuts, nil
}

func (p *parser) parseShortcut() (Shortcut, error) {
	var s Shortcut
	s.AllowDesktopConfig = 1
	s.AllowOverlay = 1

	for {
		if p.pos >= len(p.data) {
			return s, errors.New("unexpected end of data in shortcut")
		}

		typ := p.data[p.pos]
		if typ == typeEnd {
			p.pos++
			break
		}
		p.pos++

		key, err := p.readString()
		if err != nil {
			return s, fmt.Errorf("reading field key: %w", err)
		}

		switch typ {
		case typeString:
			val, err := p.readString()
			if err != nil {
				return s, fmt.Errorf("reading string value for %q: %w", key, err)
			}
			switch key {
			case "AppName", "appname":
				s.AppName = val
			case "Exe", "exe":
				s.Exe = unquoteExe(val)
			case "StartDir", "startdir":
				s.StartDir = val
			case "icon":
				s.Icon = val
			case "ShortcutPath", "shortcutpath":
				s.ShortcutPath = val
			case "LaunchOptions", "launchoptions":
				s.LaunchOptions = val
			case "DevkitGameID":
				s.DevkitGameID = val
			case "FlatpakAppID":
				s.FlatpakAppID = val
			}

		case typeInt32:
			val, err := p.readInt32()
			if err != nil {
				return s, fmt.Errorf("reading int32 value for %q: %w", key, err)
			}
			switch key {
			case "appid":
				s.AppID = val
			case "IsHidden", "ishidden":
				s.IsHidden = val
			case "AllowDesktopConfig", "allowdesktopconfig":
				s.AllowDesktopConfig = val
			case "AllowOverlay", "allowoverlay":
				s.AllowOverlay = val
			case "OpenVR", "openvr":
				s.OpenVR = val
			case "Devkit", "devkit":
				s.Devkit = val
			case "DevkitOverrideAppID":
				s.DevkitOverrideAppID = val
			case "LastPlayTime", "lastplaytime":
				s.LastPlayTime = val
			}

		case typeMap:
			if key == "tags" {
				tags, err := p.parseTags()
				if err != nil {
					return s, fmt.Errorf("parsing tags: %w", err)
				}
				s.Tags = tags
			} else {
				if err := p.skipMap(); err != nil {
					return s, fmt.Errorf("skipping unknown map %q: %w", key, err)
				}
			}

		default:
			return s, fmt.Errorf("unknown type %02x for key %q", typ, key)
		}
	}

	return s, nil
}

func (p *parser) parseTags() ([]string, error) {
	var tags []string

	for {
		if p.pos >= len(p.data) {
			return nil, errors.New("unexpected end of data in tags")
		}

		typ := p.data[p.pos]
		if typ == typeEnd {
			p.pos++
			break
		}
		p.pos++

		_, err := p.readString()
		if err != nil {
			return nil, fmt.Errorf("reading tag index: %w", err)
		}

		switch typ {
		case typeString:
			val, err := p.readString()
			if err != nil {
				return nil, fmt.Errorf("reading tag value: %w", err)
			}
			tags = append(tags, val)
		default:
			return nil, fmt.Errorf("unexpected type %02x in tags", typ)
		}
	}

	return tags, nil
}

func (p *parser) skipMap() error {
	for {
		if p.pos >= len(p.data) {
			return errors.New("unexpected end of data while skipping map")
		}

		typ := p.data[p.pos]
		if typ == typeEnd {
			p.pos++
			return nil
		}
		p.pos++

		if _, err := p.readString(); err != nil {
			return err
		}

		switch typ {
		case typeString:
			if _, err := p.readString(); err != nil {
				return err
			}
		case typeInt32:
			if _, err := p.readInt32(); err != nil {
				return err
			}
		case typeMap:
			if err := p.skipMap(); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown type %02x while skipping", typ)
		}
	}
}

func (p *parser) readString() (string, error) {
	start := p.pos
	for p.pos < len(p.data) {
		if p.data[p.pos] == 0 {
			s := string(p.data[start:p.pos])
			p.pos++
			return s, nil
		}
		p.pos++
	}
	return "", errors.New("unterminated string")
}

func (p *parser) readInt32() (uint32, error) {
	if p.pos+4 > len(p.data) {
		return 0, errors.New("not enough data for int32")
	}
	val := binary.LittleEndian.Uint32(p.data[p.pos : p.pos+4])
	p.pos += 4
	return val, nil
}

func writeString(w *bytes.Buffer, s string) error {
	if _, err := w.WriteString(s); err != nil {
		return err
	}
	return w.WriteByte(0)
}

func writeStringField(w *bytes.Buffer, key, val string) error {
	if err := w.WriteByte(typeString); err != nil {
		return err
	}
	if err := writeString(w, key); err != nil {
		return err
	}
	return writeString(w, val)
}

func writeInt32Field(w *bytes.Buffer, key string, val uint32) error {
	if err := w.WriteByte(typeInt32); err != nil {
		return err
	}
	if err := writeString(w, key); err != nil {
		return err
	}
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], val)
	_, err := w.Write(buf[:])
	return err
}

func quoteExe(exe string) string {
	if len(exe) == 0 {
		return exe
	}
	if exe[0] == '"' {
		return exe
	}
	return `"` + exe + `"`
}

func unquoteExe(exe string) string {
	if len(exe) >= 2 && exe[0] == '"' && exe[len(exe)-1] == '"' {
		return exe[1 : len(exe)-1]
	}
	return exe
}
