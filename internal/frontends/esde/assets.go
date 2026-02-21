package esde

import (
	_ "embed"

	"github.com/fnune/kyaraben/internal/model"
)

//go:embed assets/grid.png
var gridPNG []byte

//go:embed assets/hero.png
var heroPNG []byte

//go:embed assets/logo.png
var logoPNG []byte

//go:embed assets/capsule.png
var capsulePNG []byte

var esdeGridAssets = model.SteamGridAssets{
	Grid:    gridPNG,
	Hero:    heroPNG,
	Logo:    logoPNG,
	Capsule: capsulePNG,
}
