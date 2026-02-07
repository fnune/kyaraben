package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/nix"
)

type ValidateFlakeCmd struct{}

func (cmd *ValidateFlakeCmd) Run(ctx *Context) error {
	fmt.Println("Validating Nix flake for all emulators...")

	registry := ctx.NewRegistry()
	flakeGenerator := nix.NewFlakeGenerator(registry, registry)

	nixClient, err := ctx.NewNixClient()
	if err != nil {
		return fmt.Errorf("creating nix client: %w", err)
	}

	if !nixClient.IsAvailable() {
		return fmt.Errorf("nix is not available")
	}

	tmpDir, err := os.MkdirTemp("", "kyaraben-flake-validate-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	allEmulators := registry.AllEmulators()
	emulatorIDs := make([]model.EmulatorID, len(allEmulators))
	for i, emu := range allEmulators {
		emulatorIDs[i] = emu.ID
	}

	allFrontends := registry.AllFrontends()
	frontendIDs := make([]model.FrontendID, len(allFrontends))
	for i, fe := range allFrontends {
		frontendIDs[i] = fe.ID
	}

	fmt.Printf("Generating flake for %d emulators and %d frontends...\n", len(emulatorIDs), len(frontendIDs))
	genResult, err := flakeGenerator.Generate(tmpDir, emulatorIDs, frontendIDs)
	if err != nil {
		return fmt.Errorf("generating flake: %w", err)
	}

	for _, skipped := range genResult.SkippedEmulators {
		fmt.Printf("Warning: emulator '%s' was skipped (unknown)\n", skipped)
	}
	for _, skipped := range genResult.SkippedFrontends {
		fmt.Printf("Warning: frontend '%s' was skipped (unknown)\n", skipped)
	}

	fmt.Println("Evaluating flake (this checks syntax and attribute existence)...")

	evalCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := nixClient.FlakeCheck(evalCtx, string(genResult.Path)); err != nil {
		return fmt.Errorf("flake validation failed: %w", err)
	}

	fmt.Println("Flake is valid!")
	return nil
}
