package v1_test

import (
	"context"
	"testing"

	apiv1 "github.com/cyberark/secretless-broker/pkg/apis/secretless.io/v1"
	fakev1 "github.com/cyberark/secretless-broker/pkg/k8sclient/clientset/versioned/typed/secretless.io/v1/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8stesting "k8s.io/client-go/testing"
)

func TestConfigurations_Get(t *testing.T) {
	client := &fakev1.FakeSecretlessV1{Fake: &k8stesting.Fake{}}
	c := client.Configurations("ns")
	_, err := c.Get(context.TODO(), "test", metav1.GetOptions{})
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
}

func TestConfigurations_List(t *testing.T) {
	client := &fakev1.FakeSecretlessV1{Fake: &k8stesting.Fake{}}
	c := client.Configurations("ns")
	_, err := c.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
}

func TestConfigurations_Watch(t *testing.T) {
	client := &fakev1.FakeSecretlessV1{Fake: &k8stesting.Fake{}}
	// Add a watch reactor to handle the watch action
	client.Fake.PrependWatchReactor("configurations", k8stesting.DefaultWatchReactor(nil, nil))

	c := client.Configurations("ns")
	_, err := c.Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Errorf("Watch() error = %v", err)
	}
}

func TestConfigurations_Create(t *testing.T) {
	client := &fakev1.FakeSecretlessV1{Fake: &k8stesting.Fake{}}
	c := client.Configurations("ns")
	_, err := c.Create(context.TODO(), &apiv1.Configuration{}, metav1.CreateOptions{})
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
}

func TestConfigurations_Update(t *testing.T) {
	client := &fakev1.FakeSecretlessV1{Fake: &k8stesting.Fake{}}
	c := client.Configurations("ns")
	_, err := c.Update(context.TODO(), &apiv1.Configuration{ObjectMeta: metav1.ObjectMeta{Name: "test"}}, metav1.UpdateOptions{})
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}
}

func TestConfigurations_Delete(t *testing.T) {
	client := &fakev1.FakeSecretlessV1{Fake: &k8stesting.Fake{}}
	c := client.Configurations("ns")
	err := c.Delete(context.TODO(), "test", metav1.DeleteOptions{})
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}
}

func TestConfigurations_DeleteCollection(t *testing.T) {
	client := &fakev1.FakeSecretlessV1{Fake: &k8stesting.Fake{}}
	c := client.Configurations("ns")
	err := c.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{})
	if err != nil {
		t.Errorf("DeleteCollection() error = %v", err)
	}
}

func TestConfigurations_Patch(t *testing.T) {
	client := &fakev1.FakeSecretlessV1{Fake: &k8stesting.Fake{}}
	c := client.Configurations("ns")
	_, err := c.Patch(context.TODO(), "test", types.MergePatchType, []byte("{}"), metav1.PatchOptions{})
	if err != nil {
		t.Errorf("Patch() error = %v", err)
	}
}
