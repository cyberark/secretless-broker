package v1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConfigurationDeepCopy(t *testing.T) {
	orig := &Configuration{
		TypeMeta: metav1.TypeMeta{Kind: "Configuration", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name: "conf1",
		},
		Spec: ConfigurationSpec{
			Listeners: []Listener{
				{
					TypeMeta: metav1.TypeMeta{Kind: "Listener"},
					ObjectMeta: metav1.ObjectMeta{
						Name: "listener1",
					},
					CACertFiles: []string{"certA", "certB"},
				},
			},
			Handlers: []Handler{
				{
					TypeMeta: metav1.TypeMeta{Kind: "Handler"},
					ObjectMeta: metav1.ObjectMeta{
						Name: "handler1",
					},
					Match: []string{"m1", "m2"},
					Credentials: []Variable{
						{
							TypeMeta: metav1.TypeMeta{Kind: "Variable"},
							ObjectMeta: metav1.ObjectMeta{
								Name: "var1",
							},
						},
					},
				},
			},
		},
		Status: ConfigurationStatus{},
	}

	cpy := orig.DeepCopy()
	if !reflect.DeepEqual(cpy, orig) {
		t.Fatalf("DeepCopy result not equal to original\norig=%#v\ncpy=%#v", orig, cpy)
	}
	orig.Spec.Listeners[0].CACertFiles[0] = "modified"
	if cpy.Spec.Listeners[0].CACertFiles[0] != "certA" {
		t.Fatalf("DeepCopy is not deep: copy changed after original mutated")
	}

	var into Configuration
	orig.DeepCopyInto(&into)
	if !reflect.DeepEqual(&into, orig) {
		t.Fatalf("DeepCopyInto result not equal to original\norig=%#v\ninto=%#v", orig, &into)
	}
	obj := orig.DeepCopyObject()
	if obj == nil {
		t.Fatalf("DeepCopyObject returned nil")
	}
	if got, ok := obj.(*Configuration); !ok || !reflect.DeepEqual(got, orig) {
		t.Fatalf("DeepCopyObject result wrong type or not equal\nobj=%#v", obj)
	}
}

func TestConfigurationListDeepCopy(t *testing.T) {
	item := Configuration{
		TypeMeta: metav1.TypeMeta{Kind: "Configuration"},
		ObjectMeta: metav1.ObjectMeta{
			Name: "item1",
		},
	}
	orig := &ConfigurationList{
		TypeMeta: metav1.TypeMeta{Kind: "ConfigurationList"},
		Items:    []Configuration{item},
	}

	cpy := orig.DeepCopy()
	if !reflect.DeepEqual(cpy, orig) {
		t.Fatalf("ConfigurationList DeepCopy not equal to original")
	}
	orig.Items[0].ObjectMeta.Name = "changed"
	if cpy.Items[0].ObjectMeta.Name != "item1" {
		t.Fatalf("ConfigurationList DeepCopy is not deep")
	}

	obj := orig.DeepCopyObject()
	if obj == nil {
		t.Fatalf("ConfigurationList DeepCopyObject returned nil")
	}
	if got, ok := obj.(*ConfigurationList); !ok || !reflect.DeepEqual(got, orig) {
		t.Fatalf("ConfigurationList DeepCopyObject wrong")
	}
}

func TestHandlerListenerVariableDeepCopy(t *testing.T) {
	h := &Handler{
		TypeMeta: metav1.TypeMeta{Kind: "Handler"},
		ObjectMeta: metav1.ObjectMeta{
			Name: "h1",
		},
		Match: []string{"a", "b"},
		Credentials: []Variable{
			{
				TypeMeta: metav1.TypeMeta{Kind: "Variable"},
				ObjectMeta: metav1.ObjectMeta{
					Name: "cred1",
				},
			},
		},
	}
	hCpy := h.DeepCopy()
	if !reflect.DeepEqual(hCpy, h) {
		t.Fatalf("Handler DeepCopy not equal")
	}
	h.Match[0] = "changed"
	if hCpy.Match[0] != "a" {
		t.Fatalf("Handler DeepCopy is not deep")
	}

	l := &Listener{
		TypeMeta: metav1.TypeMeta{Kind: "Listener"},
		ObjectMeta: metav1.ObjectMeta{
			Name: "l1",
		},
		CACertFiles: []string{"x", "y"},
	}
	lCpy := l.DeepCopy()
	if !reflect.DeepEqual(lCpy, l) {
		t.Fatalf("Listener DeepCopy not equal")
	}
	l.CACertFiles[0] = "z"
	if lCpy.CACertFiles[0] != "x" {
		t.Fatalf("Listener DeepCopy is not deep")
	}

	v := &Variable{
		TypeMeta: metav1.TypeMeta{Kind: "Variable"},
		ObjectMeta: metav1.ObjectMeta{
			Name: "v1",
		},
	}
	vCpy := v.DeepCopy()
	if !reflect.DeepEqual(vCpy, v) {
		t.Fatalf("Variable DeepCopy not equal")
	}
}

func TestNilReceivers(t *testing.T) {
	var c *Configuration
	if c.DeepCopy() != nil {
		t.Fatalf("expected nil DeepCopy for nil Configuration")
	}
	if c.DeepCopyObject() != nil {
		t.Fatalf("expected nil DeepCopyObject for nil Configuration")
	}

	var cl *ConfigurationList
	if cl.DeepCopy() != nil {
		t.Fatalf("expected nil DeepCopy for nil ConfigurationList")
	}
	if cl.DeepCopyObject() != nil {
		t.Fatalf("expected nil DeepCopyObject for nil ConfigurationList")
	}

	var spec *ConfigurationSpec
	if spec.DeepCopy() != nil {
		t.Fatalf("expected nil DeepCopy for nil ConfigurationSpec")
	}
	if spec.DeepCopyObject() != nil {
		t.Fatalf("expected nil DeepCopyObject for nil ConfigurationSpec")
	}
}
