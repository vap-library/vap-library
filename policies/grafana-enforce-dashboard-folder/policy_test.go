package httproute_enforce_hostnames

import (
	"context"
	"fmt"
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
	var err error
	var namespaceLabels = map[string]string{"test": "true"}
	testEnv, err = testutils.CreateTestEnv("", false, namespaceLabels)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to create Kind cluster for test. Error msg: %s", err))
	}

	os.Exit(testEnv.Run(m))
}

func TestVapGrafanaEnforceDashboardFolder(t *testing.T) {

	// Create a feature
	feature := features.New("Testing applying resources").
		Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			//r, err := resources.New(config.Client().RESTConfig())
			//if err != nil {
			//	t.Fatal(err)
			//}
			//err = decoder.ApplyWithManifestDir(ctx, r, "./", "*.yaml", []resources.CreateOption{})
			//if err != nil {
			//	t.Fatal(err)
			//}
			return ctx
		}).Assess("Nginx pod can call github api", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		return ctx
	}).Feature()

	_ = testEnv.Test(t, feature)
}
