/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
)

var testenv env.Environment

func TestMain(m *testing.M) {
	testenv = env.New()
	kindClusterName := envconf.RandomName("vap-library-", 16)
	testenv.Setup(
		envfuncs.CreateCluster(kind.NewProvider(), kindClusterName),
	)
	fmt.Println("Setting up tests")

	testenv.Finish(
		envfuncs.DestroyCluster(kindClusterName),
	)

	os.Exit(testenv.Run(m))
}
