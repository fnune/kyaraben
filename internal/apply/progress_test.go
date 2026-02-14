package apply

import (
	"testing"
	"time"

	"github.com/fnune/kyaraben/internal/packages"
)

func TestProgressAggregatorEmitsSpeed(t *testing.T) {
	var last Progress
	onProgress := func(p Progress) {
		last = p
	}

	downloadTotals := map[string]int64{
		"retroarch": 1000,
	}
	packageTotals := map[string]int64{
		"retroarch": 1000,
	}
	archiveTypes := map[string]string{
		"retroarch": "7z",
	}

	agg := newProgressAggregator(packageTotals, downloadTotals, archiveTypes, 1000, onProgress)
	agg.onPackageProgress(packages.InstallProgress{
		PackageName:     "retroarch",
		Phase:           "downloading",
		BytesDownloaded: 200,
		BytesTotal:      1000,
	})
	time.Sleep(120 * time.Millisecond)
	agg.onPackageProgress(packages.InstallProgress{
		PackageName:     "retroarch",
		Phase:           "downloading",
		BytesDownloaded: 400,
		BytesTotal:      1000,
	})

	if last.BytesPerSecond <= 0 {
		t.Fatalf("expected positive bytes per second, got %d", last.BytesPerSecond)
	}
	if last.BytesDownloaded == 0 {
		t.Fatal("expected bytes downloaded to be set")
	}
}
