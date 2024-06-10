package service_type

import (
	"context"
	"fmt"
	"log"
	"os"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"testing"
	"time"
	"vap-library/testutils"
)

var testParameterYAML string = `
apiVersion: vap-library.com/v1beta1
kind: VAPLibServiceTypeParam
metadata:
  name: service-type.vap-library.com
  namespace: %s
spec:
  allowedTypes:
  - ClusterIP
  - NodePort
`

var testParameterNoCIPYAML string = `
apiVersion: vap-library.com/v1beta1
kind: VAPLibServiceTypeParam
metadata:
  name: service-type.vap-library.com
  namespace: %s
spec:
  allowedTypes:
  - NodePort
`

var serviceClusterIPYAML string = `
apiVersion: v1
kind: Service
metadata:
  name: test-clusterip
  namespace: %s
spec:
  ports:
  - appProtocol: http
    port: 8080
    targetPort: 8080
  selector:
    app: myapp
  type: ClusterIP
`

var serviceNoTypeYAML string = `
apiVersion: v1
kind: Service
metadata:
  name: test-notype
  namespace: %s
spec:
  ports:
  - appProtocol: http
    port: 8080
    targetPort: 8080
  selector:
    app: myapp
`

var serviceLBYAML string = `
apiVersion: v1
kind: Service
metadata:
  name: test-clusterip
  namespace: %s
spec:
  ports:
  - appProtocol: http
    port: 8080
    targetPort: 8080
  selector:
    app: myapp
  type: LoadBalancer
`

var testEnv env.Environment

func TestMain(m *testing.M) {
	var namespaceLabels = map[string]string{"vap-library.com/service-type": "deny"}

	var err error
	testEnv, err = testutils.CreateTestEnv("", false, namespaceLabels, nil)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to create Kind cluster for test. Error msg: %s", err))
	}

	// wait for the cluster to be ready
	time.Sleep(2 * time.Second)

	os.Exit(testEnv.Run(m))
}

func TestWithFullParameter(t *testing.T) {

	f := features.New("Service with full parameter").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// apply parameter first
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(testParameterYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			// wait for the parameter to be registered properly
			time.Sleep(10 * time.Second)

			return ctx
		}).
		Assess("A valid Service is accepted", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(serviceClusterIPYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("A Service with invalid type is rejected", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(serviceLBYAML, namespace))
			if err == nil {
				t.Fatal("A Service with invalid type was accepted")
			}

			return ctx
		}).
		Assess("A Service which does not contain any type is accepted when ClusterIP is allowed", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should PASS!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(serviceNoTypeYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}

func TestWithNoCIPParameter(t *testing.T) {

	f := features.New("Service with parameter without ClusterIP").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// apply parameter first
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(testParameterNoCIPYAML, namespace))
			if err != nil {
				t.Fatal(err)
			}

			// wait for the parameter to be registered properly
			time.Sleep(10 * time.Second)

			return ctx
		}).
		Assess("A Service with invalid type is rejected", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(serviceLBYAML, namespace))
			if err == nil {
				t.Fatal("A Service with invalid type was accepted")
			}

			return ctx
		}).
		Assess("A Service which does not contain any type is rejected when ClusterIP is allowed", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(serviceNoTypeYAML, namespace))
			if err == nil {
				t.Fatal("A Service without type was accepted")
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}

func TestWithoutParameter(t *testing.T) {

	f := features.New("Service without VAP parameter").
		Assess("Without the VAP parameter, Services are rejected", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// get namespace
			namespace := ctx.Value(testutils.GetNamespaceKey(t)).(string)

			// this should FAIL as we do not have a parameter for VAP!
			err := testutils.ApplyK8sResourceFromYAML(ctx, cfg, fmt.Sprintf(serviceClusterIPYAML, namespace))
			if err == nil {
				t.Fatal("A Service without the VAP parameter was accepted")
			}

			return ctx
		})

	_ = testEnv.Test(t, f.Feature())

}
