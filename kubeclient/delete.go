package kubeclient

import "context"
import "k8s.io/client-go/rest"

func DeleteObject(restConfig *rest.Config, o Object) error {

	client, err := rest.UnversionedRESTClientFor(restConfig)
	if err != nil {
		return err
	}

	req := client.Delete().RequestURI(o.Metadata.SelfLink)
	ctx := context.Background()
	_, err = req.DoRaw(ctx)
	return err
}
