package kubeclient

import "fmt"

type ResourceType struct {
	Name        string
	APIPath     string
	APIGroup    string
	APIVersion  string
	QueryString string
	Namespaced  bool
}

type ObjectList struct {
	Items []Object `json:"items"`
}

type Object struct {
	APIVersion string `json:"apiVersion"`
	Kind string `json:"kind"`
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	Name	 string			   `json:"name"`
	Namespace string		   `json:"namespace"`
	Labels   map[string]string `json:"labels"`
	SelfLink string            `json:"selfLink"`
	UID      string            `json:"uid"`
}

type ObjectSpec struct {
	APIVersion string
	Kind string
	Namespace string
	Name string
}

func (spec ObjectSpec) String() string {
	if spec.Namespace == "" {
		return fmt.Sprintf("%s.%s : %s\n", spec.APIVersion, spec.Kind, spec.Name)
	} else {
		return fmt.Sprintf("%s : %s.%s : %s\n", spec.Namespace, spec.APIVersion, spec.Kind, spec.Name)
	}
}

func (object *Object) MakeSpec() ObjectSpec {
	return ObjectSpec{
		Kind: object.Kind,
		Namespace: object.Metadata.Namespace,
		Name: object.Metadata.Name,
	}
}