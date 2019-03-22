package nix

import (
	"bytes"
	"errors"
	"github.com/dbcdk/kubernixos/assets"
	"os"
	"os/exec"
	"path/filepath"
)

type Config struct {
	Server   string
	Checksum string
}

func Eval(attribute string, args []string) (buffer *bytes.Buffer, err error) {
	root, _ := assets.Setup()
	defer assets.Teardown(root)
	kubernixosNix := filepath.Join(root, "kubernixos.nix")

	modules := os.Getenv("MODULES")
	if modules == "" {
		return nil, errors.New("MODULES must be set in environment")
	}
	packages := os.Getenv("PACKAGES")

	nixArgs := make([]string, 0)
	nixArgs = append(nixArgs, "eval")
	if packages != "" {
		nixArgs = append(nixArgs, []string{"--arg", "packages", packages}...)
	}
	nixArgs = append(nixArgs, []string{"--arg", "modules", modules}...)
	nixArgs = append(nixArgs, args...)
	nixArgs = append(nixArgs, []string{"-f", kubernixosNix, "--json"}...)
	nixArgs = append(nixArgs, attribute)
	cmd := exec.Command("nix", nixArgs...)

	buffer = &bytes.Buffer{}
	cmd.Stdout = buffer
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	return
}