package kubectl

import (
	"os"
	"os/exec"
)

func Apply(inFile *os.File, server string, args []string) (err error) {
	kubectlArgs := make([]string, 0)
	kubectlArgs = append(kubectlArgs, "apply")
	kubectlArgs = append(kubectlArgs, []string{"-s", server, "-f", inFile.Name()}...)
	kubectlArgs = append(kubectlArgs, []string{"-l", "!kubernixos-ignore"}...)
	kubectlArgs = append(kubectlArgs, args...)

	cmd := exec.Command("kubectl", kubectlArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	return
}
