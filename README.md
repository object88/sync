# Restarter

Restarter is a synchronizing structure that acts as a retriggering gate for another method.  The method will be started with Restarter's `Invoke` function, and must accept a cancellable `context.Context`.  If a second call to `Invoke` is made while a previous is still running, the first one will be cancelled, and then the second will be started.

### Example

The canonical example of this might be a build system that watches your local file system for source code changes.  If some file changes are detected, a series of build commands is started.  However, if the files change again while the build is still running, then that build should be stopped, and a new one started.

### Usage

``` Go
import (
  "context"
  "io/ioutil"
  "os"
  "os/exec"

  "github.com/object88/bbreloader/sync"
)

// Build builds a thing that needs building, of course
type Build struct {
  Args      []string
  restarter *sync.Restarter
}

// NewBuild creates a new Build
func NewBuild(args []string) *Build {
  r := sync.NewRestarter()
  return &Build{args, r}
}

// Run executes the step with an interruptable context
func (b *Build) Run() {
  b.restarter.Invoke(b.work)
}

func (b *Build) work(ctx context.Context) {
  tempDir, _ = ioutil.TempDir("", "")
  tempFileName := tempDir + "/a.tmp"

  // Spawn the go build command
  select {
  case <-ctx.Done():
    return nil
  default:
    completeArgs := make([]string, len(b.Args) + 3)
    completeArgs[0] = "build"
    completeArgs[1] = "-o"
    completeArgs[2] = tempFileName
    for k, v := range b.Args {
      completeArgs[k + 3] = v
    }
    cmd := exec.CommandContext(ctx, "go", completeArgs...)
    err := cmd.Run()
    if err != nil {
      return
    }
  }

  select {
  case <-ctx.Done():
  default:
    os.Rename(tempFileName, "./MyNewApplication")
  }
}
```
