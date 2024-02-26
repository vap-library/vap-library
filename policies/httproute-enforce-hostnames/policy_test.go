package httproute_enforce_hostnames

import (
	"context"
	"fmt"
	"os"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"testing"
	"vap-library/testutils"

	"sigs.k8s.io/e2e-framework/pkg/env"
)

var testEnv env.Environment

func TestMain(m *testing.M) {
	var err error
	testEnv, err = testutils.CreateTestEnv("")
	if err != nil {
		fmt.Println("haho")
		fmt.Print(err)
		os.Exit(1)
	} else {
		fmt.Println("minden ok")
	}

	os.Exit(testEnv.Run(m))
}

func TestKindCluster(t *testing.T) {
	feature := features.New("Testing applying resources").
		Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			r, err := resources.New(config.Client().RESTConfig())
			if err != nil {
				t.Fatal(err)
			}
			err = decoder.ApplyWithManifestDir(ctx, r, "./", "*.yaml", []resources.CreateOption{})
			if err != nil {
				t.Fatal(err)
			}
			return ctx
		}).Assess("Nginx pod can call github api", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		return ctx
	}).Feature()

	_ = testEnv.Test(t, feature)
}
