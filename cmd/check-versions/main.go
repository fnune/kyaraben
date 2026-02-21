package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/fnune/kyaraben/internal/versions"
)

func main() {
	if err := versions.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load version data: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	fmt.Println("Checking for package updates...")
	fmt.Println()

	var checks []versions.VersionCheck
	for check := range versions.CheckAllVersions(ctx) {
		checks = append(checks, check)
	}

	sort.Slice(checks, func(i, j int) bool {
		return checks[i].Package < checks[j].Package
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "PACKAGE\tCURRENT\tSTABLE\tPRERELEASE\tSTATUS")

	hasUpdates := false
	for _, c := range checks {
		status := "up to date"
		if c.Error != nil {
			status = fmt.Sprintf("error: %v", c.Error)
		} else if c.HasUpdate {
			status = "UPDATE AVAILABLE"
			hasUpdates = true
		}

		stable := c.LatestStable
		if stable == "" {
			stable = "-"
		}
		pre := c.LatestPre
		if pre == "" || pre == stable {
			pre = "-"
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			c.Package, c.Current, stable, pre, status)
	}

	_ = w.Flush()
	fmt.Println()

	if hasUpdates {
		fmt.Println("Some packages have updates available.")
	} else {
		fmt.Println("All packages are up to date.")
	}
}
