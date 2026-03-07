package shaders

import "github.com/fnune/kyaraben/internal/model"

type Type string

const (
	TypeCRT        Type = "crt"
	TypeSameboyLCD Type = "sameboy-lcd"
	TypeGameboy    Type = "gameboy"
)

func TypeForSystem(systemID model.SystemID) Type {
	switch systemID {
	case model.SystemIDGB:
		return TypeGameboy
	case model.SystemIDGBC, model.SystemIDGBA, model.SystemIDNGP, model.SystemIDGameGear:
		return TypeSameboyLCD
	default:
		return TypeCRT
	}
}

func PresetContent(t Type) string {
	switch t {
	case TypeSameboyLCD:
		return SameboyLCD.RetroArchPreset()
	case TypeCRT:
		return CRTLottes.RetroArchPreset()
	case TypeGameboy:
		return GameboyRetroArchPreset()
	}
	return ""
}

const (
	slangShadersCommit  = "b8a7e9e8eaf1a40210b62e6f967d25d81b37d0c8"
	SlangShadersBaseURL = "https://raw.githubusercontent.com/libretro/slang-shaders/" + slangShadersCommit
)

type DownloadFile struct {
	Path   string
	SHA256 string
}

func DownloadFiles(t Type) []DownloadFile {
	switch t {
	case TypeSameboyLCD:
		return []DownloadFile{
			{"handheld/shaders/sameboy-lcd.slang", "9cb12c233a13b2a6ca4d87c7e828859c5737577409844dc95c4ea8cc63826142"},
		}
	case TypeCRT:
		return []DownloadFile{
			{"crt/shaders/crt-lottes.slang", "e382eec413c2acf5d03cce65d447347553032acf2e550826d0142a9c51d559ef"},
		}
	case TypeGameboy:
		return []DownloadFile{
			{"handheld/shaders/gameboy/shader-files/gb-pass0.slang", "435e8d3c2f14c52564c449915992a08e3611a38992ad1fda5c65635ee4f09e39"},
			{"handheld/shaders/gameboy/shader-files/gb-pass1.slang", "796abc0db46c99760e4d9384a79ed0352c291ae40794cc9f827777c4376d1c85"},
			{"handheld/shaders/gameboy/shader-files/gb-pass2.slang", "83217e4825209d2b99f1d0d99476713af7e0ed96229db4abc5fa3de1321f3cc9"},
			{"handheld/shaders/gameboy/shader-files/gb-pass3.slang", "1b2c960473fa445aa0512b24da7c643e6ff3d2dfeeab50e7da06e2bf0c03f5d5"},
			{"handheld/shaders/gameboy/shader-files/gb-pass4.slang", "626fe5669363bfa2744779773af6288a1bdd7ae35d60ad93528f84eaa8738dfc"},
			{"handheld/shaders/gameboy/resources/palette.png", "9865e419eafb484e1b3cfdf99241d71a183e5f23871295fa1822337a22aca9fc"},
			{"handheld/shaders/gameboy/resources/background.png", "1509dbddab4779cf12a4279323524fe81ee7448cb9041b9a8d2dd98e4b8630b0"},
		}
	}
	return nil
}
