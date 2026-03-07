package shaders

import "fmt"

type CRTLottesParams struct {
	ShadowMask  float64
	WarpX       float64
	WarpY       float64
	BloomAmount float64
	BrightBoost float64
}

var CRTLottes = CRTLottesParams{
	ShadowMask:  1.0,
	WarpX:       0.01,
	WarpY:       0.01,
	BloomAmount: 0.05,
	BrightBoost: 1.15,
}

func (p CRTLottesParams) RetroArchPreset() string {
	return fmt.Sprintf(`shaders = 1

shader0 = crt-lottes.slang
filter_linear0 = false
scale_type0 = viewport

shadowMask = "%.1f"
warpX = "%.2f"
warpY = "%.2f"
bloomAmount = "%.2f"
brightBoost = "%.2f"
`, p.ShadowMask, p.WarpX, p.WarpY, p.BloomAmount, p.BrightBoost)
}
