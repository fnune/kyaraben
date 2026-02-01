package cli

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"text/tabwriter"
	"time"

	"github.com/fnune/kyaraben/internal/versions"
)

type ValidateURLsCmd struct {
	ShowSizes bool `help:"Show download sizes for each URL" default:"true"`
}

func (cmd *ValidateURLsCmd) Run(ctx *Context) error {
	fmt.Println("Validating download URLs for all emulators...")

	fetchCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	results := versions.FetchAllSizes(fetchCtx)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "EMULATOR\tVERSION\tTARGET\tSIZE\tSTATUS")

	var failed atomic.Int32
	var succeeded atomic.Int32
	var mu sync.Mutex

	for info := range results {
		mu.Lock()
		if info.Error != nil {
			failed.Add(1)
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t-\tFAILED: %v\n",
				info.Name, info.Version, info.Target, info.Error)
		} else {
			succeeded.Add(1)
			sizeStr := "-"
			if cmd.ShowSizes {
				sizeStr = versions.FormatSize(info.Size)
			}
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\tOK\n",
				info.Name, info.Version, info.Target, sizeStr)
		}
		mu.Unlock()
	}

	_ = w.Flush()
	fmt.Println()
	fmt.Printf("Results: %d OK, %d failed\n", succeeded.Load(), failed.Load())

	if failed.Load() > 0 {
		return fmt.Errorf("%d URLs failed validation", failed.Load())
	}

	return nil
}
