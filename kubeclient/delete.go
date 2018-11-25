package kubeclient

import "k8s.io/client-go/rest"

func DeleteObject(restConfig *rest.Config, o Object) error {

	client, err := rest.UnversionedRESTClientFor(restConfig)
	if err != nil {
		return err
	}

	req := client.Delete().RequestURI(o.Metadata.SelfLink)
	_, err = req.DoRaw()
	return err
}
