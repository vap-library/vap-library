package main

import (
	"os"
	"strings"
	"testing"
	"vap-library/testutils"

	"github.com/lithammer/dedent"
	"sigs.k8s.io/e2e-framework/pkg/env"
)

var testenv env.Environment

func TestMain(m *testing.M) {
	testenv = env.New()
	os.Exit(testenv.Run(m))
}

func TestVapGrafanaEnforceDashboardFolder(t *testing.T) {
	testutils.CreateFromFile("policy.yaml", t)
	testutils.RecreateNamespace("sample-app", t)
	testutils.CreateFromFile("binding.yaml", t)

	t.Run("dashboard with folder corresponding to namespace should be allowed", func(t *testing.T) {
		errorMessage := testutils.CreationShouldFail(t, dedent.Dedent(`
		apiVersion: apps/v1
		kind: Deployment
		metadata:
		  labels:
		    app: nginx
		  name: nginx
		  namespace: sample-app
		spec:
		  replicas: 6
		  selector:
		    matchLabels:
		      app: nginx
		  template:
		    metadata:
		      labels:
		        app: nginx
		    spec:
		      containers:
		      - image: nginx
		        name: nginx`))

		if !strings.HasSuffix(errorMessage, "object.spec.replicas <= 3\n") {
			t.Errorf("Unexpected error message: %s", errorMessage)
		}
	})
}