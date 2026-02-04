package nix

import (
	"context"
	"io"
	"os/exec"
)

type RunOpts struct {
	Env    []string
	Dir    string
	Stdout io.Writer
	Stderr io.Writer
}

type CommandRunner interface {
	Run(ctx context.Context, name string, args []string, opts RunOpts) error
	Output(ctx context.Context, name string, args []string, opts RunOpts) ([]byte, error)
}

type ExecRunner struct{}

func (r *ExecRunner) Run(ctx context.Context, name string, args []string, opts RunOpts) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = opts.Env
	cmd.Dir = opts.Dir
	cmd.Stdout = opts.Stdout
	cmd.Stderr = opts.Stderr
	return cmd.Run()
}

func (r *ExecRunner) Output(ctx context.Context, name string, args []string, opts RunOpts) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = opts.Env
	cmd.Dir = opts.Dir
	cmd.Stderr = opts.Stderr
	return cmd.Output()
}

type FakeRunner struct {
	Calls      []FakeRunnerCall
	RunErr     error
	OutputData []byte
	OutputErr  error
}

type FakeRunnerCall struct {
	Name string
	Args []string
	Opts RunOpts
}

func (r *FakeRunner) Run(ctx context.Context, name string, args []string, opts RunOpts) error {
	r.Calls = append(r.Calls, FakeRunnerCall{Name: name, Args: args, Opts: opts})
	return r.RunErr
}

func (r *FakeRunner) Output(ctx context.Context, name string, args []string, opts RunOpts) ([]byte, error) {
	r.Calls = append(r.Calls, FakeRunnerCall{Name: name, Args: args, Opts: opts})
	return r.OutputData, r.OutputErr
}
