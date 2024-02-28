package httproute_enforce_hostnames

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"log"
	"os"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"testing"
	"vap-library/testutils"

	"sigs.k8s.io/e2e-framework/pkg/env"
)

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/httproute-enforce-hostnames": "deny"}

	var err error
	testEnv, err = testutils.CreateTestEnv("", true, namespaceLabels)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to create Kind cluster for test. Error msg: %s", err))
	}

	os.Exit(testEnv.Run(m))
}

func TestVAPHTTPRouteEnforceHostnames(t *testing.T) {

	f := features.New("pod list").
		Assess("pods from namespace", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			var pods v1.PodList
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)
			err := cfg.Client().Resources(namespace).List(context.TODO(), &pods)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("found %d pods in namespace %s", len(pods.Items), namespace)
			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

	// Create a feature
	//feature := features.New("Testing applying resources").
	//	Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
	//		//r, err := resources.New(config.Client().RESTConfig())
	//		//if err != nil {
	//		//	t.Fatal(err)
	//		//}
	//		//err = decoder.ApplyWithManifestDir(ctx, r, "./", "*.yaml", []resources.CreateOption{})
	//		//if err != nil {
	//		//	t.Fatal(err)
	//		//}
	//		return ctx
	//	}).Assess("Nginx pod can call github api", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
	//	return ctx
	//}).Feature()

	//_ = testEnv.Test(t, feature)
}
