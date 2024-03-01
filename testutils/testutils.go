package testutils

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
	"strings"
	"testing"
	"time"
)

const (
	defaultKindVersion = "v1.29.2"
	kindNamePrefix     = "vaplibtest"
	testNamespace      = "vap-testing"
)

type (
	NamespaceCtxKey string
	ClusterCtxKey   string
)

func CreateTestEnv(kindVersion string, keepLogs bool, namespaceLabels map[string]string, extraResourcesFromDir map[string]string) (env.Environment, error) {

	// Specifying a run ID so that multiple runs wouldn't collide.
	runID := envconf.RandomName(testNamespace, 14)

	// Use the default Kind version if none is provided
	if kindVersion == "" {
		kindVersion = defaultKindVersion
	}

	// Create a new environment from the flags
	var testEnv env.Environment
	testEnv, _ = env.NewFromFlags()

	// Define an empty slice of EnvFunc type for Env setup and finish
	var setupFuncs []env.Func
	var finishFuncs []env.Func

	// Create cluster
	kindClusterName := envconf.RandomName(kindNamePrefix, 16)
	setupFuncs = append(setupFuncs, envfuncs.CreateClusterWithConfig(kind.NewProvider(), kindClusterName, "../../testutils/kind-config.yaml", kind.WithImage("kindest/node:"+kindVersion)))

	// Apply all yaml from the policy directory
	setupFuncs = append(
		setupFuncs,
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			return applyResourcesFromDir(ctx, cfg, "./", "*.yaml", 2)
		},
	)

	// Apply extra resources
	for dir, pattern := range extraResourcesFromDir {
		setupFuncs = append(
			setupFuncs,
			func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
				return applyResourcesFromDir(ctx, cfg, dir, pattern, 1)
			},
		)
	}

	testEnv.Setup(setupFuncs...)

	// Remove the applied resources
	finishFuncs = append(
		finishFuncs,
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			return deleteResourcesFromDir(ctx, cfg, "./", "*.yaml")
		},
	)

	// Keep the logs if the flag is set
	if keepLogs {
		finishFuncs = append(finishFuncs, envfuncs.ExportClusterLogs(kindClusterName, "./test-logs"))
	}

	// Destroy the cluster
	finishFuncs = append(finishFuncs, envfuncs.DestroyCluster(kindClusterName))

	testEnv.Finish(finishFuncs...)

	// Set the BeforeEachTest and AfterEachTest functions that creates and deletes a namespace for each test
	testEnv.BeforeEachTest(func(ctx context.Context, cfg *envconf.Config, t *testing.T) (context.Context, error) {
		return createNSForTest(ctx, cfg, t, runID, namespaceLabels)
	})
	testEnv.AfterEachTest(func(ctx context.Context, cfg *envconf.Config, t *testing.T) (context.Context, error) {
		return deleteNSForTest(ctx, cfg, t, runID)
	})

	return testEnv, nil
}

// applyResourcesFromDir applies all the resources from the given directory
func applyResourcesFromDir(ctx context.Context, cfg *envconf.Config, dir string, pattern string, waitSec time.Duration) (context.Context, error) {
	r, err := resources.New(cfg.Client().RESTConfig())
	if err != nil {
		return ctx, err
	}
	err = decoder.ApplyWithManifestDir(ctx, r, dir, pattern, []resources.CreateOption{})
	if err != nil {
		return ctx, err
	}

	// Wait for the resources to be registered properly
	time.Sleep(waitSec * time.Second)

	return ctx, nil
}

// deleteResourcesFromDir removes all the resources from the given directory
func deleteResourcesFromDir(ctx context.Context, cfg *envconf.Config, dir string, pattern string) (context.Context, error) {
	r, err := resources.New(cfg.Client().RESTConfig())
	if err != nil {
		return ctx, err
	}
	err = decoder.DeleteWithManifestDir(ctx, r, dir, pattern, []resources.DeleteOption{})
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

// createNSForTest creates a random namespace with the runID as a prefix. It is stored in the context
// so that the deleteNSForTest routine can look it up and delete it.
func createNSForTest(ctx context.Context, cfg *envconf.Config, t *testing.T, runID string, namespaceLabels map[string]string) (context.Context, error) {
	ns := envconf.RandomName(runID, 20)
	ctx = context.WithValue(ctx, GetNamespaceKey(t), ns)

	t.Logf("Creating NS %v for test %v", ns, t.Name())
	nsObj := v1.Namespace{}
	nsObj.Name = ns
	nsObj.Labels = namespaceLabels
	return ctx, cfg.Client().Resources().Create(ctx, &nsObj)
}

// deleteNSForTest looks up the namespace corresponding to the given test and deletes it.
func deleteNSForTest(ctx context.Context, cfg *envconf.Config, t *testing.T, _ string) (context.Context, error) {
	ns := fmt.Sprint(ctx.Value(GetNamespaceKey(t)))
	t.Logf("Deleting NS %v for test %v", ns, t.Name())

	nsObj := v1.Namespace{}
	nsObj.Name = ns
	return ctx, cfg.Client().Resources().Delete(ctx, &nsObj)
}

// GetNamespaceKey returns the context key for a given test
func GetNamespaceKey(t *testing.T) NamespaceCtxKey {
	// When we pass t.Name() from inside an `assess` step, the name is in the form TestName/Features/Assess
	if strings.Contains(t.Name(), "/") {
		return NamespaceCtxKey(strings.Split(t.Name(), "/")[0])
	}

	// When pass t.Name() from inside a `testenv.BeforeEachTest` function, the name is just TestName
	return NamespaceCtxKey(t.Name())
}

// ApplyK8sResourceFromYAML applies a k8s resource from a yaml string
func ApplyK8sResourceFromYAML(ctx context.Context, cfg *envconf.Config, yaml string) error {
	r, err := resources.New(cfg.Client().RESTConfig())
	if err != nil {
		return err
	}

	obj, err := decoder.DecodeAny(strings.NewReader(yaml))
	if err != nil {
		return err
	}
	handler := decoder.CreateHandler(r)
	return handler(ctx, obj)
}
