package assets

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func Setup() (assetRoot string, err error) {
	assetRoot, err = ioutil.TempDir("", "kubernixos-")
	if err != nil {
		return "", err
	}

	kubernixos, err := Asset("lib/kubernixos.nix")
	if err != nil {
		return "", err
	}
	kubernixosPath := filepath.Join(assetRoot, "kubernixos.nix")
	ioutil.WriteFile(kubernixosPath, kubernixos, 0644)

	return
}

func Teardown(assetRoot string) (err error) {
	err = os.Remove(filepath.Join(assetRoot, "kubernixos.nix"))
	if err != nil {
		return err
	}

	err = os.Remove(assetRoot)
	if err != nil {
		return err
	}

	return nil
}
