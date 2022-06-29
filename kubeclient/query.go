package kubeclient

import (
	"context"
	"encoding/json"
	"github.com/dbcdk/kubernixos/nix"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"strings"
)

const labelName = "kubernixos"

func GetResourcesToPrune(restConfig *rest.Config, config *nix.Config, types []ResourceType) (map[string]Object, error) {
	resources := make(map[string]Object, 0)
	for _, t := range types {

		restConfig.APIPath = t.APIPath
		restConfig.GroupVersion = &schema.GroupVersion{
			Group:   t.APIGroup,
			Version: t.APIVersion,
		}

		client, err := rest.RESTClientFor(restConfig)
		if err != nil {
			return nil, err
		}

		req := client.Get().
			Resource(t.Name).
			Param("labelSelector", labelName)

		ctx := context.Background()
		raw, err := req.DoRaw(ctx)
		if err != nil {
			if !errors.IsNotFound(err) {
				return nil, err
			}
			continue
		}

		var target ObjectList
		err = json.Unmarshal(raw, &target)
		if err != nil {
			return nil, err
		}
		if len(target.Items) > 0 {
			for _, i := range target.Items {
				if config.Checksum != i.Metadata.Labels["kubernixos"] {
					resources[i.Metadata.UID] = i
				}
			}
		}
	}

	return resources, nil
}

func GetResourceTypes(clients *kubernetes.Clientset) (resources []ResourceType, err error) {
	resources = make([]ResourceType, 0)

	var resourceList []*v1.APIResourceList
	resourceList, err = getApiResources(clients)
	for _, rl := range resourceList {
		for _, r := range rl.APIResources {

			if hasVerbs([]string{"GET", "LIST", "UPDATE", "PATCH", "DELETE"}, r.Verbs) {
				groupVersionParts := strings.Split(rl.GroupVersion, "/")
				queryStringParts := make([]string, 0)
				queryStringParts = append(queryStringParts, r.Name)
				var apiGroup = ""
				var apiVersion = ""
				if len(groupVersionParts) > 1 {
					queryStringParts = append(queryStringParts, groupVersionParts[1])
					queryStringParts = append(queryStringParts, groupVersionParts[0])
					apiGroup = groupVersionParts[0]
					apiVersion = groupVersionParts[1]
				} else {
					apiVersion = groupVersionParts[0]
				}
				var apiPath = "/api"
				if apiGroup != "" {
					apiPath = "/apis"
				}
				resources = append(resources, ResourceType{
					Name:        r.Name,
					APIGroup:    apiGroup,
					APIVersion:  apiVersion,
					APIPath:     apiPath,
					QueryString: strings.Join(queryStringParts, "."),
					Namespaced:  r.Namespaced,
				})
			}
		}
	}
	return
}

func getApiResources(clients *kubernetes.Clientset) (resourceList []*v1.APIResourceList, err error) {
	resourceList, err = clients.Discovery().ServerPreferredResources()
  return
}

func hasVerbs(needles []string, haystack []string) bool {
	var needed int
	var found int
	needed = len(needles)
	found = 0
	for _, n := range needles {
		n = strings.ToLower(n)
		var progress = found
		for _, h := range haystack {
			if strings.ToLower(h) == n {
				found++
				break
			}
		}
		if found <= progress {
			return false
		}
	}
	return found == needed
}
