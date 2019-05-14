package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dbcdk/kubernixos/kubeclient"
	"github.com/dbcdk/kubernixos/kubectl"
	"github.com/dbcdk/kubernixos/nix"
	"io/ioutil"
	"k8s.io/client-go/rest"
	"os"
	"path"
	"regexp"
	"strings"

	// needed to enable oidc authentication
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	godiff "github.com/sourcegraph/go-diff/diff"
)

var (
	doDiff = false
	doApply = false
	doPrune = false
	doDump  = false
	nixArgs = make([]string, 0)
	ignoreDiffOn = map[string]bool{
		"kubernixos": true,
		"generation": true,
	}
)

func main() {

	passthroughArgs := parseCmdline(os.Args[1:])

	deployFile, err := ioutil.TempFile("", "kubernixos")
	fail("init", err)
	defer os.Remove(deployFile.Name())

	config, err := eval(deployFile)
	fail("eval", err)

	// The diff command cannot currently be combined with apply
	if doDiff {
		err = diff(deployFile, config, passthroughArgs)
		fail("diff", err)
	} else {
		err = apply(deployFile, config, passthroughArgs)
		fail("apply", err)
	}

	restConfig, err := kubeclient.GetKubeConfig(config.Server)
	fail("kube-config", err)

	clients, err := kubeclient.GetKubeClient(restConfig)
	fail("kube-client", err)

	var types []kubeclient.ResourceType
	types, err = kubeclient.GetResourceTypes(clients)
	fail("resource-types", err)

	objects, err := kubeclient.GetResourcesToPrune(restConfig, config, types)
	fail("all-resources", err)

	prune(objects, restConfig)
}

func fail(stage string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error occurred during stage \"%s\": %s", stage, err.Error())
		os.Exit(1)
	}
}

func parseCmdline(args []string) (passthroughArgs []string) {
	for _, a := range args {
		if !parseArg(a) {
			passthroughArgs = append(passthroughArgs, a)
		}
	}
	return
}

func parseArg(arg string) bool {
	switch arg {
	case "diff":
		doDiff = true
		return true
	case "apply":
		doApply = true
		return true
	case "prune":
		doPrune = true
		return true
	case "dump":
		doDump = true
		return true
	case "--show-trace":
		nixArgs = append(nixArgs, arg)
		return true
	case "--prune":
		fmt.Fprintf(os.Stderr, "Usage of `kubectl apply --prune` is disabled in kubernixos\n")
		os.Exit(1)
	}

	return false
}

func diff(inFile *os.File, config *nix.Config, args []string) error {

	difs, err := kubectl.Diff(inFile, config.Server, args)
	if err != nil {
		return err
	}

	rx, err := regexp.Compile("^(\\+|\\-)\\s*([a-zA-Z0-9]+):")
	if err != nil {
		return err
	}
	for _, d := range difs {
		if hasChanged(d, rx) {
			parts := strings.Split(path.Base(d.OrigName), ".")
			name, parts := parts[len(parts)-1], parts[:len(parts)-1]
			namespace, parts := parts[len(parts)-1], parts[:len(parts)-1]
			kind := strings.Join(parts, ".")
			if namespace == "" {
				fmt.Printf("%s : %s\n", kind, name)
			} else {
				fmt.Printf("%s : %s : %s\n", namespace, kind, name)
			}
		}
	}

	return nil
}

func hasChanged(d *godiff.FileDiff, rx *regexp.Regexp) bool {
	for _, h := range d.Hunks {
		lines := strings.Split(string(h.Body), "\n")
		for _, l := range lines {
			field := rx.FindStringSubmatch(l)
			if len(field) == 3 {
				if _, ok := ignoreDiffOn[field[2]]; ok {
					continue
				} else {
					return true
				}
			}
		}
	}
	return false
}

func apply(inFile *os.File, config *nix.Config, args []string) error {
	if !doApply {
		return nil
	}
	return kubectl.Apply(inFile, config.Server, args)
}

func eval(outFile *os.File) (config *nix.Config, err error) {
	var raw *bytes.Buffer
	var byteArr []byte
	var data map[string]map[string]interface{}

	raw, err = nix.Eval("kubernixos", nixArgs)
	if err != nil {
		return
	}
	err = json.Unmarshal(raw.Bytes(), &data)
	if err != nil {
		return
	}

	byteArr, err = json.Marshal(data["manifests"])
	config = &nix.Config{
		Server:   data["config"]["server"].(string),
		Checksum: data["config"]["checksum"].(string),
	}
	ioutil.WriteFile(outFile.Name(), byteArr, 0755)
	if doDump {
		var out bytes.Buffer
		json.Indent(&out, byteArr, "", "\t")
		out.WriteTo(os.Stdout)
		fmt.Println()
	}
	return
}

func prune(objects map[string]kubeclient.Object, restConfig *rest.Config) {
	for _, o := range objects {
		fmt.Print("Pruning: ")
		fmt.Print(o.Metadata.SelfLink)
		fmt.Print(", checksum: ")
		fmt.Print(o.Metadata.Labels["kubernixos"])
		fmt.Println(" (dry-run)")
	}

	count := len(objects)
	if count > 0 && doPrune {
		fmt.Fprintf(os.Stderr, "You are about to delete %d objects, please confirm with 'yes' or 'no': ", count)
		if askForConfirmation() {
			for _, o := range objects {
				fmt.Print("Pruning: ")
				fmt.Print(o.Metadata.SelfLink)
				fmt.Print(", checksum: ")
				fmt.Print(o.Metadata.Labels["kubernixos"])
				fmt.Println()

				err := kubeclient.DeleteObject(restConfig, o)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to prune object: %s with error: %s\n",
						o.Metadata.SelfLink,
						err.Error())
				}
			}
		}
	}
}

func askForConfirmation() bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false
	}
	response = strings.ToLower(response)
	if response == "yes" {
		return true
	} else if response == "no" {
		return false
	} else {
		fmt.Print("Please type yes or no and then press enter: ")
		return askForConfirmation()
	}
}
