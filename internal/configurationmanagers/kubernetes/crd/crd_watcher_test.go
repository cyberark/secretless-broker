package crd

import (
	"context"
	"testing"
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	api_v1 "github.com/cyberark/secretless-broker/pkg/apis/secretless.io/v1"
	secretlessClientset "github.com/cyberark/secretless-broker/pkg/k8sclient/clientset/versioned"
	secretlessClientsetfake "github.com/cyberark/secretless-broker/pkg/k8sclient/clientset/versioned/fake"
)

type chanHandler struct {
	addCh    chan *api_v1.Configuration
	updateCh chan [2]*api_v1.Configuration
	delCh    chan *api_v1.Configuration
}

func newChanHandler() *chanHandler {
	return &chanHandler{
		addCh:    make(chan *api_v1.Configuration, 1),
		updateCh: make(chan [2]*api_v1.Configuration, 1),
		delCh:    make(chan *api_v1.Configuration, 1),
	}
}

func (h *chanHandler) CRDAdded(cfg *api_v1.Configuration) {
	h.addCh <- cfg
}

func (h *chanHandler) CRDDeleted(cfg *api_v1.Configuration) {
	h.delCh <- cfg
}

func (h *chanHandler) CRDUpdated(oldCfg, newCfg *api_v1.Configuration) {
	h.updateCh <- [2]*api_v1.Configuration{oldCfg, newCfg}
}

func waitFor[T any](t *testing.T, ch <-chan T, timeout time.Duration) T {
	select {
	case v := <-ch:
		return v
	case <-time.After(timeout):
		t.Fatalf("timed out waiting for event")
	}
	panic("unreachable")
}

func Test_RegisterCRDListener_AddUpdateDelete(t *testing.T) {
	// save/restore hooks
	oldNewKube := newKubernetesConfig
	oldNewClient := newSecretlessClientForConfig
	defer func() {
		newKubernetesConfig = oldNewKube
		newSecretlessClientForConfig = oldNewClient
	}()

	newKubernetesConfig = func() (*rest.Config, error) {
		return &rest.Config{}, nil
	}

	// create fake client and wire the factory to return it
	fakeClient := secretlessClientsetfake.NewSimpleClientset()
	newSecretlessClientForConfig = func(cfg *rest.Config) (secretlessClientset.Interface, error) {
		return fakeClient, nil
	}

	namespace := "default"
	configName := "my-config-spec"
	handler := newChanHandler()

	// start the informer
	if err := RegisterCRDListener(namespace, configName, handler); err != nil {
		t.Fatalf("RegisterCRDListener returned error: %v", err)
	}

	// allow informer to start
	time.Sleep(100 * time.Millisecond)

	created, err := fakeClient.SecretlessV1().Configurations(namespace).Create(context.TODO(),
		&api_v1.Configuration{
			ObjectMeta: meta_v1.ObjectMeta{Name: configName},
		},
		meta_v1.CreateOptions{},
	)
	if err != nil {
		t.Fatalf("failed to create configuration: %v", err)
	}
	gotAdd := waitFor(t, handler.addCh, 2*time.Second)
	if gotAdd.ObjectMeta.Name != created.ObjectMeta.Name {
		t.Fatalf("Add event name mismatch: got %s want %s", gotAdd.ObjectMeta.Name, created.ObjectMeta.Name)
	}

	created.Spec = api_v1.ConfigurationSpec{} // modify to create an update
	updated, err := fakeClient.SecretlessV1().Configurations(namespace).Update(context.TODO(), created, meta_v1.UpdateOptions{})
	if err != nil {
		t.Fatalf("failed to update configuration: %v", err)
	}
	gotUpdate := waitFor(t, handler.updateCh, 2*time.Second)
	if gotUpdate[1].ObjectMeta.Name != updated.ObjectMeta.Name {
		t.Fatalf("Update event name mismatch: got %s want %s", gotUpdate[1].ObjectMeta.Name, updated.ObjectMeta.Name)
	}

	if err := fakeClient.SecretlessV1().Configurations(namespace).Delete(context.TODO(), configName, meta_v1.DeleteOptions{}); err != nil {
		t.Fatalf("failed to delete configuration: %v", err)
	}
	gotDel := waitFor(t, handler.delCh, 2*time.Second)
	if gotDel.ObjectMeta.Name != configName {
		t.Fatalf("Delete event name mismatch: got %s want %s", gotDel.ObjectMeta.Name, configName)
	}
}
