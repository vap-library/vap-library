package testutils

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/lithammer/dedent"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func CreateVapFromFile3(filename string, ctx context.Context, c *envconf.Config) (k8s.Object, error) {
	fmt.Println("Creating VAP from file", filename)

	content, _ := os.ReadFile(filename)
	object, _, err := scheme.Codecs.UniversalDeserializer().Decode(content, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decode VAP: %v", err)
	}

	client, err := c.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	o, _ := object.(k8s.Object)

	err = client.Resources().Create(ctx, o)
	if err != nil {
		return nil, fmt.Errorf("failed to create VAP: %v", err)
	}

	return o, nil
}

// Creates kubernetes resources from a YAML file.
func CreateFromFile(filename string, t *testing.T) {
	out, err := runCommand("kubectl", "apply", "-f", filename)
	if err != nil {
		t.Fatalf("failed to create kubernetes resources: %v", out)
	}
	t.Log(out)
}

// Deletes and recreates a namespace. If the namespace does not exist, it is just created.
func RecreateNamespace(name string, t *testing.T) {
	DeleteNamespace(name, t)

	out, err := runCommand("kubectl", "create", "namespace", name)
	if err != nil {
		t.Fatalf("failed to create namespace: %v", out)
	}
	t.Log(out)
}

// Deletes a namespace. Fails silently if the namespace does not exist.
func DeleteNamespace(name string, t *testing.T) {
	runCommand("kubectl", "delete", "namespace", name)
	// ignore errors if namespace is not there
}

// Deletes the kubernetes resource specified in the YAML file
func DeleteFromFile(id string) {
	fmt.Println("Destroying", id)
}

// Creates a kubernetes resource from a YAML definition. Fail the test if the creation fails.
func CreationShouldSucceed(t *testing.T, resourceDef string) {
	out, err := runCommandWithInput(resourceDef, "kubectl", "apply", "-f", "-")
	if err != nil {
		t.Fatalf("Creation failed: %v", out)
	}
}

// Creates a kubernetes resource from a YAML definition. Fail the test if the creation succeeds.
// If the creation fails, return the error message.
func CreationShouldFail(t *testing.T, resourceDef string) string {
	out, err := runCommandWithInput(resourceDef, "kubectl", "apply", "-f", "-")
	if err == nil {
		t.Fatalf("Creation should have failed, but it succeeded: %v", out)
	}
	return out
}

// De-indents a string and replaces tabs with spaces.
func Dedent(text string) string {
	return strings.ReplaceAll(dedent.Dedent(text), "\t", "    ")
}

func runCommandWithInput(input string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = bytes.NewBufferString(input)
	buf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.Stdout = buf
	cmd.Stderr = errBuf
	err := cmd.Run()
	if err != nil {
		return errBuf.String(), err
	}

	return buf.String(), nil
}

func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	buf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.Stdout = buf
	cmd.Stderr = errBuf
	err := cmd.Run()
	if err != nil {
		return errBuf.String(), err
	}

	return buf.String(), nil
}
