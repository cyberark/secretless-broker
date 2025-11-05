package crd

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	apiextensionsclientsetfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
)

func Test_createCRD_Succeeds(t *testing.T) {
	assert := assert.New(t)

	fakeClient := apiextensionsclientsetfake.NewSimpleClientset()

	assert.NoError(createCRD(fakeClient))

	// ensure CRD exists in the fake client's tracker
	got, err := fakeClient.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), CRDFQDNName, meta_v1.GetOptions{})
	assert.NoError(err)
	assert.NotNil(got)
}

func Test_createCRD_AlreadyExists(t *testing.T) {
	assert := assert.New(t)

	fakeClient := apiextensionsclientsetfake.NewSimpleClientset()

	// First create should succeed
	assert.NoError(createCRD(fakeClient))

	// Second create should hit the AlreadyExists path and return nil
	assert.NoError(createCRD(fakeClient))
}

func Test_createCRD_OtherError(t *testing.T) {
	assert := assert.New(t)

	fakeClient := apiextensionsclientsetfake.NewSimpleClientset()

	// Prepend reactor to force a non-AlreadyExists error on create
	fakeClient.PrependReactor("create", "customresourcedefinitions",
		func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, apierrors.NewInternalError(fmt.Errorf("boom"))
		})

	err := createCRD(fakeClient)
	assert.Error(err)
	assert.Contains(err.Error(), "Could not create Secretless CRD")
}
