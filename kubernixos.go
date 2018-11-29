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
	// needed to enable oidc authentication
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

var (
	doApply = false
	doPrune = false
	doDump  = false
	nixArgs = make([]string, 0)
)

func main() {

	passthroughArgs := parseCmdline(os.Args[1:])

	deployFile, err := ioutil.TempFile("", "kubernixos")
	fail("init", err)
	defer os.Remove(deployFile.Name())

	config, err := eval(deployFile)
	fail("eval", err)

	err = apply(deployFile, config, passthroughArgs)
	fail("apply", err)

	restConfig, err := kubeclient.GetKubeConfig(config.Server)
	fail("kube-config", err)

	clients, err := kubeclient.GetKubeClient(restConfig)
	fail("kube-client", err)

	var types []kubeclient.ResourceType
	types, err = kubeclient.GetResourceTypes(clients)
	fail("resource-types", err)

	objects, err := kubeclient.GetAllResources(restConfig, config, types)
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

		if doPrune {
			fmt.Println()
			err := kubeclient.DeleteObject(restConfig, o)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to prune object: %s with error: %s\n",
					o.Metadata.SelfLink,
					err.Error())
			}
		} else {
			fmt.Println(" (dry-run)")
		}
	}
}
