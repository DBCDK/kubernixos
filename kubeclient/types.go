package kubeclient

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
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	Labels   map[string]string `json:"labels"`
	SelfLink string            `json:"selfLink"`
	UID      string            `json:"uid"`
}
