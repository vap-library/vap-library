package testutils

import (
	"context"
	"fmt"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
	"sigs.k8s.io/e2e-framework/support/utils"
)

func CreateTestEnv(kindVersion string) (env.Environment, error) {

	// Hard code a kind version as default
	if kindVersion == "" {
		kindVersion = "v1.29.2"
	}

	var testEnv env.Environment
	testEnv, _ = env.NewFromFlags()

	// Define an empty slice of EnvFunc type for Env setup and finish
	var setupFuncs []env.Func
	var finishFuncs []env.Func

	// Create cluster
	kindClusterName := envconf.RandomName("vaplibtest-", 16)
	setupFuncs = append(setupFuncs, envfuncs.CreateClusterWithConfig(kind.NewProvider(), kindClusterName, "../../testutils/kind-config.yaml", kind.WithImage("kindest/node:"+kindVersion)))

	// Apply the CRD for the parameter if we got one
	crdFileName := "./crd-parameter.yaml"
	setupFuncs = append(
		setupFuncs,
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			if p := utils.RunCommand(fmt.Sprintf("kubectl apply --server-side -f %s", crdFileName)); p.Err() != nil {
				return ctx, p.Err()
			}
			return ctx, nil
		},
	)

	//if _, err := os.Stat(crdFileName); err == nil {
	//	setupFuncs = append(
	//		setupFuncs,
	//		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
	//			// Deploy the controller components
	//			if p := utils.RunCommand(fmt.Sprintf("kubectl apply --server-side -f %s", crdFileName)); p.Err() != nil {
	//				return ctx, p.Err()
	//			}
	//			return ctx, nil
	//		},
	//	)
	//
	//	setupFuncs = append(
	//		finishFuncs,
	//		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
	//			utils.RunCommand(fmt.Sprintf("kubectl delete -f %s", crdFileName))
	//			return ctx, nil
	//		},
	//	)
	//} else {
	//	return nil, err
	//}

	testEnv.Setup(setupFuncs...)

	finishFuncs = append(finishFuncs, envfuncs.ExportClusterLogs(kindClusterName, "./test-logs"))
	finishFuncs = append(finishFuncs, envfuncs.DestroyCluster(kindClusterName))

	testEnv.Finish(finishFuncs...)

	return testEnv, nil
}
