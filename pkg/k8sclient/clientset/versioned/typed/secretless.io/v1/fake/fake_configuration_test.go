package fake

import (
	"context"
	"errors"
	_ "reflect"
	testing2 "testing"

	secretlessiov1 "github.com/cyberark/secretless-broker/pkg/apis/secretless.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/testing"
)

func newFakeConfigurations() *FakeConfigurations {
	return &FakeConfigurations{
		Fake: &FakeSecretlessV1{Fake: &testing.Fake{}},
		ns:   "default",
	}
}

func TestFakeConfigurations_Get(t *testing2.T) {
	fc := newFakeConfigurations()
	expected := &secretlessiov1.Configuration{}
	fc.Fake.AddReactor("get", "configurations", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
		return true, expected, nil
	})
	got, err := fc.Get(context.TODO(), "test", metav1.GetOptions{})
	if err != nil || got != expected {
		t.Errorf("Get() = %v, %v; want %v, nil", got, err, expected)
	}
	// test nil obj
	fc.Fake = &FakeSecretlessV1{Fake: &testing.Fake{}}
	fc.Fake.AddReactor("get", "configurations", func(action testing.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("err")
	})
	got, err = fc.Get(context.TODO(), "test", metav1.GetOptions{})
	if err == nil || got != nil {
		t.Errorf("Get() with nil obj = %v, %v; want nil, err", got, err)
	}
}

func TestFakeConfigurations_List(t *testing2.T) {
	fc := newFakeConfigurations()
	item := secretlessiov1.Configuration{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"a": "b"}}}
	list := &secretlessiov1.ConfigurationList{Items: []secretlessiov1.Configuration{item}}
	fc.Fake.AddReactor("list", "configurations", func(action testing.Action) (bool, runtime.Object, error) {
		return true, list, nil
	})
	got, err := fc.List(context.TODO(), metav1.ListOptions{})
	if err != nil || len(got.Items) != 1 {
		t.Errorf("List() = %v, %v; want 1 item, nil", got, err)
	}
	// test nil obj
	fc.Fake = &FakeSecretlessV1{Fake: &testing.Fake{}}
	fc.Fake.AddReactor("list", "configurations", func(action testing.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("err")
	})
	got, err = fc.List(context.TODO(), metav1.ListOptions{})
	if err == nil || got != nil {
		t.Errorf("List() with nil obj = %v, %v; want nil, err", got, err)
	}
}

func TestFakeConfigurations_Watch(t *testing2.T) {
	fc := newFakeConfigurations()
	fc.Fake.AddWatchReactor("configurations", testing.DefaultWatchReactor(watch.NewFake(), nil))
	_, err := fc.Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Errorf("Watch() error = %v; want nil", err)
	}
}

func TestFakeConfigurations_Create(t *testing2.T) {
	fc := newFakeConfigurations()
	expected := &secretlessiov1.Configuration{}
	fc.Fake.AddReactor("create", "configurations", func(action testing.Action) (bool, runtime.Object, error) {
		return true, expected, nil
	})
	got, err := fc.Create(context.TODO(), expected, metav1.CreateOptions{})
	if err != nil || got != expected {
		t.Errorf("Create() = %v, %v; want %v, nil", got, err, expected)
	}
	// test nil obj
	fc.Fake = &FakeSecretlessV1{Fake: &testing.Fake{}}
	fc.Fake.AddReactor("create", "configurations", func(action testing.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("err")
	})
	got, err = fc.Create(context.TODO(), expected, metav1.CreateOptions{})
	if err == nil || got != nil {
		t.Errorf("Create() with nil obj = %v, %v; want nil, err", got, err)
	}
}

func TestFakeConfigurations_Update(t *testing2.T) {
	fc := newFakeConfigurations()
	expected := &secretlessiov1.Configuration{}
	fc.Fake.AddReactor("update", "configurations", func(action testing.Action) (bool, runtime.Object, error) {
		return true, expected, nil
	})
	got, err := fc.Update(context.TODO(), expected, metav1.UpdateOptions{})
	if err != nil || got != expected {
		t.Errorf("Update() = %v, %v; want %v, nil", got, err, expected)
	}
	// test nil obj
	fc.Fake = &FakeSecretlessV1{Fake: &testing.Fake{}}
	fc.Fake.AddReactor("update", "configurations", func(action testing.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("err")
	})
	got, err = fc.Update(context.TODO(), expected, metav1.UpdateOptions{})
	if err == nil || got != nil {
		t.Errorf("Update() with nil obj = %v, %v; want nil, err", got, err)
	}
}

func TestFakeConfigurations_Delete(t *testing2.T) {
	fc := newFakeConfigurations()
	fc.Fake.AddReactor("delete", "configurations", func(action testing.Action) (bool, runtime.Object, error) {
		return true, nil, nil
	})
	err := fc.Delete(context.TODO(), "test", metav1.DeleteOptions{})
	if err != nil {
		t.Errorf("Delete() error = %v; want nil", err)
	}
}

func TestFakeConfigurations_DeleteCollection(t *testing2.T) {
	fc := newFakeConfigurations()
	fc.Fake.AddReactor("delete-collection", "configurations", func(action testing.Action) (bool, runtime.Object, error) {
		return true, &secretlessiov1.ConfigurationList{}, nil
	})
	err := fc.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{})
	if err != nil {
		t.Errorf("DeleteCollection() error = %v; want nil", err)
	}
}

func TestFakeConfigurations_Patch(t *testing2.T) {
	fc := newFakeConfigurations()
	expected := &secretlessiov1.Configuration{}
	fc.Fake.AddReactor("patch", "configurations", func(action testing.Action) (bool, runtime.Object, error) {
		return true, expected, nil
	})
	got, err := fc.Patch(context.TODO(), "test", types.MergePatchType, []byte("{}"), metav1.PatchOptions{})
	if err != nil || got != expected {
		t.Errorf("Patch() = %v, %v; want %v, nil", got, err, expected)
	}
	// test nil obj
	fc.Fake = &FakeSecretlessV1{Fake: &testing.Fake{}}
	fc.Fake.AddReactor("patch", "configurations", func(action testing.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("err")
	})
	got, err = fc.Patch(context.TODO(), "test", types.MergePatchType, []byte("{}"), metav1.PatchOptions{})
	if err == nil || got != nil {
		t.Errorf("Patch() with nil obj = %v, %v; want nil, err", got, err)
	}
}
