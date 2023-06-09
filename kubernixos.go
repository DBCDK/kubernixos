package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dbcdk/kubernixos/kubeclient"
	"github.com/dbcdk/kubernixos/kubectl"
	"github.com/dbcdk/kubernixos/nix"
	"k8s.io/client-go/rest"
	"os"
	"path/filepath"
	"strings"

	// needed to enable oidc authentication
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

var (
	doBuild = false
    doTest = false
	doApply = false
	doPrune = false
	doDump  = false
	nixArgs = make([]string, 0)
)

func main() {

	passthroughArgs := parseCmdline(os.Args[1:])
	var config *nix.Config
	var manifests *map[string]interface{}
	var err error
	var inFile *os.File

    if doTest {




    }

	if doDump {
		config, manifests, err = readManifests()
		fail("eval", err)
	} else {
		config, err = readConfig()
	}


	if doDump {
		byteArr, err := json.Marshal(manifests)
		fail("dump", err)
		var out bytes.Buffer
		json.Indent(&out, byteArr, "", "\t")
		out.WriteTo(os.Stdout)
		fmt.Println()
	} else {
		// Print the checksum only, if dump isn't requested
		config, err := readConfig()
		fail("config", err)
		fmt.Println(config.Checksum)
	}

	if doBuild || doApply || doPrune {
		build, err := nix.Build("build", nixArgs)
		fail("build", err)
		fmt.Println(build) // print outpath to stdout
		inFile, err = os.Open(filepath.Join(build, "kubernixos.json"))
		fail("validate", err)
	}

	// non of the below steps should be taken if we're not in either apply or prune mode
	if doApply || doPrune {
		err = apply(inFile, config, passthroughArgs)
		fail("apply", err)

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
    case "test":
        doTest = true 
        return true
	case "build":
		doBuild = true
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

func apply(inFile *os.File, config *nix.Config, args []string) error {
	if !doApply {
		return nil
	}
	return kubectl.Apply(inFile, config.Server, args)
}

func read(attr string) (data map[string]map[string]interface{}, err error) {
	var raw *bytes.Buffer
	raw, err = nix.Eval(attr, nixArgs)
	if err != nil {
		return
	}
	err = json.Unmarshal(raw.Bytes(), &data)
	if err != nil {
		return
	}

	return
}

func subread(attr string) (data map[string]interface{}, err error) {
	var raw *bytes.Buffer
	raw, err = nix.Eval(attr, nixArgs)
	if err != nil {
		return
	}
	err = json.Unmarshal(raw.Bytes(), &data)
	if err != nil {
		return
	}

	return
}

func readConfig() (*nix.Config, error) {
	data, err := subread("eval.config")
	if err != nil {
		return nil, err
	}
	return configFromData(&data), nil
}

func readManifests() (*nix.Config, *map[string]interface{}, error) {
	data, err := read("eval")
	if err != nil {
		return nil, nil, err
	}
	mData := data["manifests"]
	cData := data["config"]
	return configFromData(&cData), &mData, nil
}

func configFromData(data *map[string]interface{}) *nix.Config {
	return &nix.Config{
		Server:   (*data)["server"].(string),
		Checksum: (*data)["checksum"].(string),
	}
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
