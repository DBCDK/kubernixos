package kubectl

import (
	"bytes"
	"github.com/sourcegraph/go-diff/diff"
	"os"
	"os/exec"
)


func Diff(inFile *os.File, server string, args []string) (fileDiff []*diff.FileDiff, err error) {
	kubectlArgs := make([]string, 0)
	kubectlArgs = append(kubectlArgs, "diff")
	kubectlArgs = append(kubectlArgs, []string{"-s", server, "-f", inFile.Name()}...)
	kubectlArgs = append(kubectlArgs, args...)

	var buf bytes.Buffer
	cmd := exec.Command("kubectl", kubectlArgs...)
	cmd.Stdout = &buf

	err = cmd.Run()

	return diff.ParseMultiFileDiff(buf.Bytes())
}

func Apply(inFile *os.File, server string, args []string) (err error) {
	kubectlArgs := make([]string, 0)
	kubectlArgs = append(kubectlArgs, "apply")
	kubectlArgs = append(kubectlArgs, []string{"-s", server, "-f", inFile.Name()}...)
	kubectlArgs = append(kubectlArgs, args...)

	cmd := exec.Command("kubectl", kubectlArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	return
}
