// +build acceptance

package provider

import (
	"testing"
	"time"

	tfstatehelper "github.com/hashicorp/terraform-provider-kubernetes-alpha/test/helper/state"
)

func TestKubernetesManifest_WaitForFields_Pod(t *testing.T) {
	name := randName()
	namespace := randName()

	tf := tfhelper.RequireNewWorkingDir(t)
	tf.SetReattachInfo(reattachInfo)
	defer func() {
		tf.RequireDestroy(t)
		tf.Close()
		k8shelper.AssertNamespacedResourceDoesNotExist(t, "v1", "pods", namespace, name)
	}()

	k8shelper.CreateNamespace(t, namespace)
	defer k8shelper.DeleteNamespace(t, namespace)

	tfvars := TFVARS{
		"namespace": namespace,
		"name":      name,
	}
	tfconfig := loadTerraformConfig(t, "wait_for_fields_pod.tf", tfvars)
	tf.RequireSetConfig(t, tfconfig)
	tf.RequireInit(t)

	startTime := time.Now()
	t.Log("Running terraform apply. This test should wait around 10 seconds.")
	tf.RequireApply(t)

	// NOTE We set a readinessProbe in the fixture with a delay of 10s
	// so the apply should take at least 10 seconds to complete.
	minDuration := time.Duration(10) * time.Second
	applyDuration := time.Since(startTime)
	if applyDuration < minDuration {
		t.Fatalf("the apply should have taken at least %s", minDuration)
	}

	k8shelper.AssertNamespacedResourceExists(t, "v1", "pods", namespace, name)

	tfstate := tfstatehelper.NewHelper(tf.RequireState(t))
	tfstate.AssertAttributeValues(t, tfstatehelper.AttributeValues{
		"kubernetes_manifest.test.wait_for": map[string]interface{}{
			"fields": map[string]interface{}{
				"metadata.annotations[\"test.terraform.io\"]": "test",

				"status.containerStatuses[0].ready":        "true",
				"status.containerStatuses[0].restartCount": "0",

				"status.podIP": "^(\\d+(\\.|$)){4}",
				"status.phase": "Running",
			},
		},
	})
}
