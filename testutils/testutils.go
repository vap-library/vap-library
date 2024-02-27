package testutils

import (
	"context"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
	"time"
)

const (
	defaultKindVersion = "v1.29.2"
	kindNamePrefix     = "vaplibtest"
	testNamespace      = "vap-testing"
)

func CreateTestEnv(kindVersion string, keepLogs bool, namespaceLabels map[string]string) (env.Environment, error) {

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
			r, err := resources.New(cfg.Client().RESTConfig())
			if err != nil {
				return ctx, err
			}
			err = decoder.ApplyWithManifestDir(ctx, r, "./", "*.yaml", []resources.CreateOption{})
			if err != nil {
				return ctx, err
			}

			// Sleep 2 sec to make sure the API has the VAP and binding properly "registered"
			time.Sleep(2 * time.Second)

			return ctx, nil
		},
	)

	// Create a namespace for the test with the labels
	setupFuncs = append(
		setupFuncs,
		envfuncs.CreateNamespace(testNamespace, envfuncs.WithLabels(namespaceLabels)),
	)

	// Apply the CRD for the parameter if we got one
	//crdFileName := "./crd-parameter.yaml"
	//setupFuncs = append(
	//	setupFuncs,
	//	func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
	//		if p := utils.RunCommand(fmt.Sprintf("kubectl apply --server-side -f %s", crdFileName)); p.Err() != nil {
	//			return ctx, p.Err()
	//		}
	//		return ctx, nil
	//	},
	//)

	testEnv.Setup(setupFuncs...)

	// Remove the test namespace
	finishFuncs = append(finishFuncs, envfuncs.DeleteNamespace(testNamespace))

	// Keep the logs if the flag is set
	if keepLogs {
		finishFuncs = append(finishFuncs, envfuncs.ExportClusterLogs(kindClusterName, "./test-logs"))
	}

	// Destroy the cluster
	finishFuncs = append(finishFuncs, envfuncs.DestroyCluster(kindClusterName))

	testEnv.Finish(finishFuncs...)

	return testEnv, nil
}
