package yarn

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/apex/log"
	"ttp.sh/frost/internal/manager"
)

type handler struct {
}

func init() {
	manager.AddHandler(&handler{})
}

func (h *handler) Name() string {
	return "yarn"
}

func (h *handler) Detect(root string) bool {
	_, err := os.Stat(path.Join(root, "yarn.lock"))
	return err == nil
}
func (h *handler) Exec(ctx context.Context, root string, jobs chan manager.Job) (err error) {

	return nil
}

func (h *handler) PostExec(ctx context.Context, root string) error {
	log.Info("yarn")
	cmd := exec.Command("yarn")
	cmd.Dir = root
	var out bytes.Buffer
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return err
	}
	fmt.Printf(out.String())
	return nil
}
