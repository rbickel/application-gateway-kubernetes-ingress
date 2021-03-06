// -------------------------------------------------------------------------------------------
// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.
// --------------------------------------------------------------------------------------------

package k8scontext

import (
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/version"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

var (
	runtimeScheme = k8sruntime.NewScheme()

	// IsNetworkingV1PackageSupported is flag that indicates whether networking/v1beta ingress should be used instead.
	IsNetworkingV1PackageSupported bool
)

func init() {
	extensionsv1beta1.AddToScheme(runtimeScheme)
	networkingv1.AddToScheme(runtimeScheme)
}

func fromExtensions(old *extensionsv1beta1.Ingress) (*networkingv1.Ingress, error) {
	networkingIngress := &networkingv1.Ingress{}

	err := runtimeScheme.Convert(old, networkingIngress, nil)
	if err != nil {
		return nil, err
	}

	return networkingIngress, nil
}

func fromV1beta1(old *networkingv1beta1.Ingress) (*networkingv1.Ingress, error) {
	networkingIngress := &networkingv1.Ingress{}

	err := runtimeScheme.Convert(old, networkingIngress, nil)
	if err != nil {
		return nil, err
	}

	return networkingIngress, nil
}

func toIngress(obj interface{}) (*networkingv1.Ingress, bool) {
	oldVersion, inExtension := obj.(*extensionsv1beta1.Ingress)
	if inExtension {
		ing, err := fromExtensions(oldVersion)
		if err != nil {
			klog.Errorf("unexpected error converting Ingress from extensions package: %v", err)
			return nil, false
		}

		return ing, true
	}

	v1beta1Version, inV1beta1 := obj.(*networkingv1beta1.Ingress)
	if inV1beta1 {
		ing, err := fromV1beta1(v1beta1Version)
		if err != nil {
			klog.Errorf("unexpected error converting Ingress from v1beta1 version: %v", err)
			return nil, false
		}

		return ing, true
	}

	if ing, ok := obj.(*networkingv1.Ingress); ok {
		return ing, true
	}

	return nil, false
}

// SupportsNetworkingPackage checks if the package "k8s.io/api/networking/v1"
// is available or not and if Ingress V1 is supported (k8s >= v1.18.0)
func SupportsNetworkingPackage(client clientset.Interface) (bool, bool) {
	// check kubernetes version to use new ingress package or not
	version114, _ := version.ParseGeneric("v1.14.0")
	version118, _ := version.ParseGeneric("v1.18.0")

	serverVersion, err := client.Discovery().ServerVersion()
	if err != nil {
		return false, false
	}

	runningVersion, err := version.ParseGeneric(serverVersion.String())
	if err != nil {
		klog.Errorf("unexpected error parsing running Kubernetes version: %v", err)
		return false, false
	}

	return runningVersion.AtLeast(version114), runningVersion.AtLeast(version118)
}
