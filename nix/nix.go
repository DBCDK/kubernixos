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
	kubernixosNix := filepath.Join(root, "eval.nix")

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

func Build(attribute string, args []string) (path string, err error) {
	root, _ := assets.Setup()
	defer assets.Teardown(root)
	kubernixosNix := filepath.Join(root, "eval.nix")

	modules := os.Getenv("MODULES")
	if modules == "" {
		return "", errors.New("MODULES must be set in environment")
	}
	packages := os.Getenv("PACKAGES")

	nixArgs := make([]string, 0)
	nixArgs = append(nixArgs, []string{"build", "-v"}...)
	if packages != "" {
		nixArgs = append(nixArgs, []string{"--arg", "packages", packages}...)
	}
	nixArgs = append(nixArgs, []string{"--arg", "modules", modules}...)
	nixArgs = append(nixArgs, args...)
	nixArgs = append(nixArgs, []string{"-o", "result-kubernixos"}...)
	nixArgs = append(nixArgs, []string{"-f", kubernixosNix}...)
	nixArgs = append(nixArgs, []string{attribute}...)
	cmd := exec.Command("nix", nixArgs...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return "", err
	}
	return os.Readlink("result-kubernixos")
}
