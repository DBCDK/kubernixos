package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/DBCDK/kingpin"
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
	app       = kingpin.New("Kubernixos", "Simple kubernetes manifest reconciler configured by NixOS modules")
	dump      = parseDump(app.Command("dump", "Dump all kubernetes resources in manifest to stdout."))
	apply     = parseApply(app.Command("apply", "Builds the manifest, applies resources to the cluster, then prunes old dpeloyments"))
	build     = parseBuild(app.Command("build", "Returns a store-path with the kubernetes resources from the manifest."))
	test      = parseBuild(app.Command("test", "test."))
	manifest  string
	doPrune   bool
	showTrace bool
	nixArgs   []string
)

func parseManifestAndGlobalFlags(cmd *kingpin.CmdClause) {
	parseShowTraceFlag(cmd)
	cmd.Arg("manifest", "File containing the kubernixos manifest").Required().StringVar(&manifest)
}

func parseDump(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	parseManifestAndGlobalFlags(cmd)
	return cmd
}

func parseApply(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	parseManifestAndGlobalFlags(cmd)
	parsePruneFlag(cmd)
	return cmd
}

func parsePruneFlag(cmd *kingpin.CmdClause) {
	cmd.Flag("prune", "Prune deployments with noncurrent label").Default("False").BoolVar(&doPrune)
}

func parseShowTraceFlag(cmd *kingpin.CmdClause) {
	cmd.Flag("show-trace", "Whether to ask interactively for remote sudo password when needed").Default("False").BoolVar(&showTrace)
}

func parseBuild(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	parseManifestAndGlobalFlags(cmd)
	return cmd
}

func addFlagsToNixArgs() {
	if showTrace {
		nixArgs = append(nixArgs, "--show-trace")
	}
}

func fail(stage string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error occurred during stage \"%s\": %s", stage, err.Error())
		os.Exit(1)
	}
}

func main() {
	parser := kingpin.MustParse(app.Parse(os.Args[1:]))
	addFlagsToNixArgs()

	switch parser {
	case dump.FullCommand():
		_, manifests, err := readManifests()
		fail("eval", err)
		parseAndDump(manifests)
	case apply.FullCommand():
		config, err := readConfig()
		fail("eval", err)
		inFile := buildResources()
		applyResources(inFile, config)
	case build.FullCommand():
		_ = buildResources()
	}
}

func parseAndDump(manifests *map[string]interface{}) {
	var buffer bytes.Buffer
	byteArr, err := json.Marshal(manifests)
	fail("parse-manifests", err)
	json.Indent(&buffer, byteArr, "", "\t")
	buffer.WriteTo(os.Stdout)
	fmt.Println()
}

func buildResources() (inFile *os.File) {
	build, err := nix.Build("build", nixArgs)
	fail("build", err)
	fmt.Println(build) // print outpath to stdout
	inFile, err = os.Open(filepath.Join(build, "kubernixos.json"))
	fail("validate", err)
	return
}

func applyResources(inFile *os.File, config *nix.Config) {
	err := applyKubernixos(inFile, config, nixArgs)
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

	pruneCluster(objects, restConfig)
}

func applyKubernixos(inFile *os.File, config *nix.Config, args []string) error {
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

func pruneCluster(objects map[string]kubeclient.Object, restConfig *rest.Config) {
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
