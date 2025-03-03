package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-provider-kubernetes-alpha/openapi"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"

	// this is how client-go expects auth plugins to be loaded
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// keys into the global state storage
const (
	OAPIFoundry string = "OPENAPIFOUNDRY"
)

// getDynamicClient returns a configured unstructured (dynamic) client instance
func (ps *RawProviderServer) getDynamicClient() (dynamic.Interface, error) {
	if ps.dynamicClient != nil {
		return ps.dynamicClient, nil
	}
	if ps.clientConfig == nil {
		return nil, fmt.Errorf("cannot create dynamic client: no client config")
	}
	dynClient, err := dynamic.NewForConfig(ps.clientConfig)
	if err != nil {
		return nil, err
	}
	ps.dynamicClient = dynClient
	return dynClient, nil
}

// getDiscoveryClient returns a configured discovery client instance.
func (ps *RawProviderServer) getDiscoveryClient() (discovery.DiscoveryInterface, error) {
	if ps.discoveryClient != nil {
		return ps.discoveryClient, nil
	}
	if ps.clientConfig == nil {
		return nil, fmt.Errorf("cannot create discovery client: no client config")
	}
	discoClient, err := discovery.NewDiscoveryClientForConfig(ps.clientConfig)
	if err != nil {
		return nil, err
	}
	ps.discoveryClient = discoClient
	return discoClient, nil
}

// getRestMapper returns a RESTMapper client instance
func (ps *RawProviderServer) getRestMapper() (meta.RESTMapper, error) {
	if ps.restMapper != nil {
		return ps.restMapper, nil
	}
	dc, err := ps.getDiscoveryClient()
	if err != nil {
		return nil, err
	}
	cacher := memory.NewMemCacheClient(dc)
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(cacher)
	ps.restMapper = mapper
	return mapper, nil
}

// getRestClient returns a raw REST client instance
func (ps *RawProviderServer) getRestClient() (rest.Interface, error) {
	if ps.restClient != nil {
		return ps.restClient, nil
	}
	if ps.clientConfig == nil {
		return nil, fmt.Errorf("cannot create REST client: no client config")
	}
	restClient, err := rest.UnversionedRESTClientFor(ps.clientConfig)
	if err != nil {
		return nil, err
	}
	ps.restClient = restClient
	return restClient, nil
}

// getOAPIFoundry returns an interface to request cty types from an OpenAPI spec
func (ps *RawProviderServer) getOAPIFoundry() (openapi.Foundry, error) {
	if ps.OAPIFoundry != nil {
		return ps.OAPIFoundry, nil
	}

	rc, err := ps.getRestClient()
	if err != nil {
		return nil, fmt.Errorf("failed get OpenAPI spec: %s", err)
	}

	rq := rc.Verb("GET").Timeout(30*time.Second).AbsPath("openapi", "v2")
	rs, err := rq.DoRaw(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed get OpenAPI spec: %s", err)
	}

	oapif, err := openapi.NewFoundryFromSpecV2(rs)
	if err != nil {
		return nil, fmt.Errorf("failed construct OpenAPI foundry: %s", err)
	}

	ps.OAPIFoundry = oapif

	return oapif, nil
}

func loggingTransport(rt http.RoundTripper) http.RoundTripper {
	return &loggingRountTripper{
		ot: rt,
		lt: logging.NewTransport("Kubernetes API", rt),
	}
}

type loggingRountTripper struct {
	ot http.RoundTripper
	lt http.RoundTripper
}

func (t *loggingRountTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Path == "/openapi/v2" {
		// don't trace-log the OpenAPI spec document, it's really big
		return t.ot.RoundTrip(req)
	}
	return t.lt.RoundTrip(req)
}
