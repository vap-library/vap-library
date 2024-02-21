package testutils

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"

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

func CreateFromFile(filename string, t *testing.T) {
	out, err := runCommand("kubectl", "apply", "-f", filename)
	if err != nil {
		t.Fatalf("failed to create kubernetes resources: %v", out)
	}
	fmt.Println(out)
}

func RecreateNamespace(name string, t *testing.T) {
	out, err := runCommand("kubectl", "delete", "namespace", name)
	fmt.Println(out)

	out, err = runCommand("kubectl", "create", "namespace", name)
	if err != nil {
		t.Fatalf("failed to create namespace: %v", out)
	}
	fmt.Println(out)
}

func Destroy(id string) {
	fmt.Println("Destroying", id)
}

func CreationShouldSucceed(t *testing.T, resourceDef string) {
	out, err := runCommandWithInput(resourceDef, "kubectl", "apply", "-f", "-")
	if err != nil {
		t.Fatalf("Creation failed: %v", out)
	}
}

func CreationShouldFail(t *testing.T, resourceDef string) string {
	out, err := runCommandWithInput(resourceDef, "kubectl", "apply", "-f", "-")
	if err == nil {
		t.Fatalf("Creation should have failed, but it succeeded: %v", out)
	}
	return out
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
