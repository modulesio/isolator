package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"runtime"

	"github.com/go-errors/errors"
	"github.com/itchio/butler/mansion"
	"github.com/itchio/butler/manager"
	"github.com/itchio/butler/runner"
)

var args = struct {
	dir     *string
	command *[]string
}{}

func Register(ctx *mansion.Context) {
	cmd := ctx.App.Command("runner", "Runs a command").Hidden()
	args.dir = cmd.Flag("dir", "The working directory for the command").Hidden().String()
	args.command = cmd.Arg("command", "A command to run, with arguments").Strings()
	ctx.Register(cmd, do)
}

func do(ctx *mansion.Context) {
	ctx.Must(Do())
}

func Do() error {
	/* command := *args.command
	dir := *args.dir */

  dirPath := "/tmp/app"
  fullTargetPath := "/tmp/app/node"
  installPath := "/tmp/app/install"
  var args []string = []string{}
  envBlock := os.Environ()
  prereqsDir := "/tmp/prereqs"
	localRuntime := manager.CurrentRuntime()

  runParams := &runner.RunnerParams{
		// Consumer: consumer,
		// Conn:     conn,
		// Ctx:      ctx,

		Sandbox: true,

		FullTargetPath: fullTargetPath,

		Name:   dirPath,
		Dir:    dirPath,
		Args:   args,
		Env:    envBlock,
		// Stdout: stdout,
		// Stderr: stderr,

		PrereqsDir:    prereqsDir,
		// Credentials:   params.Credentials,
		InstallFolder: installPath,
		Runtime:       localRuntime,
	}

  run, err := runner.GetRunner(runParams)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	err = run.Prepare()
	if err != nil {
		return errors.Wrap(err, 0)
	}

  exitCode, err := interpretRunError(run.Run())
  if err != nil {
    return errors.Wrap(err, 0)
  }

  if exitCode != 0 {
    var signedExitCode = int64(exitCode)
    if runtime.GOOS == "windows" {
      // Windows uses 32-bit unsigned integers as exit codes, although the
      // command interpreter treats them as signed. If a process fails
      // initialization, a Windows system error code may be returned.
      signedExitCode = int64(int32(signedExitCode))

      // The line above turns `4294967295` into -1
    }

    exeName := filepath.Base(runParams.FullTargetPath)
    msg := fmt.Sprintf("Exit code 0x%x (%d) for (%s)", uint32(exitCode), signedExitCode, exeName)
    fmt.Printf(msg)

    /* if runDuration.Seconds() > 10 {
      fmt.Printf("That's after running for %s, ignoring non-zero exit code", runDuration)
    } else { */
      return errors.New(msg)
    // }
  }

  /* launcherParams := &LauncherParams{
		// Conn:     conn,
		Ctx:      ctx,
		// Consumer: consumer,

		FullTargetPath: "/tmp",
		// Candidate:      candidate,
		// AppManifest:    appManifest,
		// Action:         manifestAction,
		Sandbox:        true,
		Args:           [],
		Env:            env,

		PrereqsDir:    "/tmp/prereqs",
		// Credentials:   params.Credentials,
		InstallFolder: "/tmp/install",
		Runtime:       manager.CurrentRuntime(),
	}

	err = launcher.Do(launcherParams)
	if err != nil {
		return errors.Wrap(err, 0)
	} */

	/* cmd := exec.Command(command[0], command[1:]...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		return errors.Wrap(err, 0)
	} */

	return nil
}

func interpretRunError(err error) (int, error) {
	if err != nil {
		if exitError, ok := AsExitError(err); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), nil
			}
		}

		return 127, err
	}

	return 0, nil
}

func AsExitError(err error) (*exec.ExitError, bool) {
	if err == nil {
		return nil, false
	}

	if se, ok := err.(*errors.Error); ok {
		return AsExitError(se.Err)
	}

	if ee, ok := err.(*exec.ExitError); ok {
		return ee, true
	}

	return nil, false
}
