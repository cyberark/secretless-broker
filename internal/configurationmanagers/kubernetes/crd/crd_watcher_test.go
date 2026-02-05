package crd

import (
	"context"
	"testing"
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	k8stesting "k8s.io/client-go/testing"

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

	// Add a custom watch reactor that sends bookmark events to satisfy newer k8s client-go expectations
	fakeClient.PrependWatchReactor("configurations", func(action k8stesting.Action) (handled bool, ret watch.Interface, err error) {
		gvr := action.GetResource()
		ns := action.GetNamespace()
		w, err := fakeClient.Tracker().Watch(gvr, ns)
		if err != nil {
			return false, nil, err
		}
		// Wrap the watch to send an initial bookmark event
		fw := watch.NewFake()
		go func() {
			defer fw.Stop()
			defer w.Stop()

			// Send a Bookmark event with the required annotation to signal end of initial list
			bookmarkEvent := &api_v1.Configuration{
				ObjectMeta: meta_v1.ObjectMeta{
					ResourceVersion: "1",
					Annotations: map[string]string{
						"k8s.io/initial-events-end": "true",
					},
				},
			}
			fw.Action(watch.Bookmark, bookmarkEvent)

			// Forward all real events from the tracker
			for event := range w.ResultChan() {
				fw.Action(event.Type, event.Object)
			}
		}()
		return true, fw, nil
	})

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

	// Give informer time to start and process the initial bookmark
	time.Sleep(500 * time.Millisecond)

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
