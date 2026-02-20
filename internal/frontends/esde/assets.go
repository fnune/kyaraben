package esde

import (
	_ "embed"

	"github.com/fnune/kyaraben/internal/model"
)

//go:embed assets/grid.jpg
var gridJPG []byte

//go:embed assets/hero.jpg
var heroJPG []byte

//go:embed assets/logo.png
var logoPNG []byte

//go:embed assets/capsule.jpg
var capsuleJPG []byte

var esdeGridAssets = model.SteamGridAssets{
	Grid:    gridJPG,
	Hero:    heroJPG,
	Logo:    logoPNG,
	Capsule: capsuleJPG,
}
