// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package test

import (
	appsv1 "k8s.io/api/apps/v1"
	certificatesv1b1 "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	mcsapiv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

type GroupKindNamespacedName struct {
	Group     gwapiv1.Group
	Kind      gwapiv1.Kind
	Namespace gwapiv1.Namespace
	Name      gwapiv1.ObjectName
}

// GetEnvoyProxy returns an EnvoyProxy object with the provided ns/name.
func GetEnvoyProxy(nsName types.NamespacedName, mergeGateways bool) *egv1a1.EnvoyProxy {
	return &egv1a1.EnvoyProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsName.Name,
			Namespace: nsName.Namespace,
		},
		Spec: egv1a1.EnvoyProxySpec{
			MergeGateways: &mergeGateways,
		},
	}
}

// GetGatewayClass returns a sample GatewayClass.
func GetGatewayClass(name string, controller gwapiv1.GatewayController, envoyProxy *GroupKindNamespacedName) *gwapiv1.GatewayClass {
	gwc := &gwapiv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: gwapiv1.GatewayClassSpec{
			ControllerName: controller,
		},
	}

	if envoyProxy != nil {
		gwc.Spec.ParametersRef = &gwapiv1.ParametersReference{
			Group:     envoyProxy.Group,
			Kind:      envoyProxy.Kind,
			Name:      string(envoyProxy.Name),
			Namespace: &envoyProxy.Namespace,
		}
	}

	return gwc
}

// GetGateway returns a sample Gateway with single listener.
func GetGateway(nsName types.NamespacedName, gwclass string, listenerPort int32) *gwapiv1.Gateway {
	return &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsName.Namespace,
			Name:      nsName.Name,
		},
		Spec: gwapiv1.GatewaySpec{
			GatewayClassName: gwapiv1.ObjectName(gwclass),
			Listeners: []gwapiv1.Listener{
				{
					Name:     "test",
					Port:     gwapiv1.PortNumber(listenerPort),
					Protocol: gwapiv1.HTTPProtocolType,
				},
			},
		},
	}
}

// GetSecureGateway returns a sample Gateway with single TLS listener.
func GetSecureGateway(nsName types.NamespacedName, gwclass string, secretKindNSName GroupKindNamespacedName) *gwapiv1.Gateway {
	secureGateway := GetGateway(nsName, gwclass, 8080)
	secureGateway.Spec.Listeners[0].TLS = &gwapiv1.GatewayTLSConfig{
		Mode: ptr.To(gwapiv1.TLSModeTerminate),
		CertificateRefs: []gwapiv1.SecretObjectReference{{
			Kind:      &secretKindNSName.Kind,
			Namespace: &secretKindNSName.Namespace,
			Name:      secretKindNSName.Name,
		}},
	}

	return secureGateway
}

// GetSecret returns a sample Secret object.
func GetSecret(nsName types.NamespacedName) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsName.Namespace,
			Name:      nsName.Name,
		},
	}
}

func GetServiceBackendRef(name types.NamespacedName, port int32) gwapiv1.BackendObjectReference {
	return gwapiv1.BackendObjectReference{
		Name: gwapiv1.ObjectName(name.Name),
		Port: ptr.To(gwapiv1.PortNumber(port)),
	}
}

func GetServiceImportBackendRef(name types.NamespacedName, port int32) gwapiv1.BackendObjectReference {
	return gwapiv1.BackendObjectReference{
		Name:  gwapiv1.ObjectName(name.Name),
		Port:  ptr.To(gwapiv1.PortNumber(port)),
		Kind:  gatewayapi.KindPtr(resource.KindServiceImport),
		Group: gatewayapi.GroupPtr(mcsapiv1a1.GroupName),
	}
}

// GetHTTPRoute returns a sample HTTPRoute with a parent reference.
func GetHTTPRoute(nsName types.NamespacedName, parent string, backendRef gwapiv1.BackendObjectReference, httpRouteFilterName string) *gwapiv1.HTTPRoute {
	httpRoute := &gwapiv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsName.Namespace,
			Name:      nsName.Name,
		},
		Spec: gwapiv1.HTTPRouteSpec{
			CommonRouteSpec: gwapiv1.CommonRouteSpec{
				ParentRefs: []gwapiv1.ParentReference{
					{Name: gwapiv1.ObjectName(parent)},
				},
			},
			Rules: []gwapiv1.HTTPRouteRule{
				{
					BackendRefs: []gwapiv1.HTTPBackendRef{
						{
							BackendRef: gwapiv1.BackendRef{
								BackendObjectReference: backendRef,
							},
						},
					},
				},
			},
		},
	}

	if httpRouteFilterName != "" {
		httpRoute.Spec.Rules[0].Filters = []gwapiv1.HTTPRouteFilter{
			{
				Type: gwapiv1.HTTPRouteFilterExtensionRef,
				ExtensionRef: &gwapiv1.LocalObjectReference{
					Group: egv1a1.GroupName,
					Kind:  egv1a1.KindHTTPRouteFilter,
					Name:  gwapiv1.ObjectName(httpRouteFilterName),
				},
			},
		}
	}

	return httpRoute
}

// GetGRPCRoute returns a sample GRPCRoute with a parent reference.
func GetGRPCRoute(nsName types.NamespacedName, parent string, serviceName types.NamespacedName, port int32) *gwapiv1.GRPCRoute {
	return &gwapiv1.GRPCRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsName.Namespace,
			Name:      nsName.Name,
		},
		Spec: gwapiv1.GRPCRouteSpec{
			CommonRouteSpec: gwapiv1.CommonRouteSpec{
				ParentRefs: []gwapiv1.ParentReference{
					{Name: gwapiv1.ObjectName(parent)},
				},
			},
			Rules: []gwapiv1.GRPCRouteRule{
				{
					BackendRefs: []gwapiv1.GRPCBackendRef{
						{
							BackendRef: gwapiv1.BackendRef{
								BackendObjectReference: gwapiv1.BackendObjectReference{
									Name: gwapiv1.ObjectName(serviceName.Name),
									Port: ptr.To(gwapiv1.PortNumber(port)),
								},
							},
						},
					},
				},
			},
		},
	}
}

// GetTLSRoute returns a sample TLSRoute with a parent reference.
func GetTLSRoute(nsName types.NamespacedName, parent string, serviceName types.NamespacedName, port int32) *gwapiv1a2.TLSRoute {
	return &gwapiv1a2.TLSRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsName.Namespace,
			Name:      nsName.Name,
		},
		Spec: gwapiv1a2.TLSRouteSpec{
			CommonRouteSpec: gwapiv1a2.CommonRouteSpec{
				ParentRefs: []gwapiv1a2.ParentReference{
					{Name: gwapiv1a2.ObjectName(parent)},
				},
			},
			Rules: []gwapiv1a2.TLSRouteRule{
				{
					BackendRefs: []gwapiv1a2.BackendRef{
						{
							BackendObjectReference: gwapiv1a2.BackendObjectReference{
								Name: gwapiv1a2.ObjectName(serviceName.Name),
								Port: ptr.To(gwapiv1.PortNumber(port)),
							},
						},
					},
				},
			},
		},
	}
}

// GetTCPRoute returns a sample TCPRoute with a parent reference.
func GetTCPRoute(nsName types.NamespacedName, parent string, serviceName types.NamespacedName, port int32) *gwapiv1a2.TCPRoute {
	return &gwapiv1a2.TCPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsName.Namespace,
			Name:      nsName.Name,
		},
		Spec: gwapiv1a2.TCPRouteSpec{
			CommonRouteSpec: gwapiv1a2.CommonRouteSpec{
				ParentRefs: []gwapiv1a2.ParentReference{
					{Name: gwapiv1a2.ObjectName(parent)},
				},
			},
			Rules: []gwapiv1a2.TCPRouteRule{
				{
					BackendRefs: []gwapiv1a2.BackendRef{
						{
							BackendObjectReference: gwapiv1a2.BackendObjectReference{
								Name: gwapiv1a2.ObjectName(serviceName.Name),
								Port: ptr.To(gwapiv1a2.PortNumber(port)),
							},
						},
					},
				},
			},
		},
	}
}

// GetUDPRoute returns a sample UDPRoute with a parent reference.
func GetUDPRoute(nsName types.NamespacedName, parent string, serviceName types.NamespacedName, port int32) *gwapiv1a2.UDPRoute {
	return &gwapiv1a2.UDPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsName.Namespace,
			Name:      nsName.Name,
		},
		Spec: gwapiv1a2.UDPRouteSpec{
			CommonRouteSpec: gwapiv1a2.CommonRouteSpec{
				ParentRefs: []gwapiv1a2.ParentReference{
					{Name: gwapiv1a2.ObjectName(parent)},
				},
			},
			Rules: []gwapiv1a2.UDPRouteRule{
				{
					BackendRefs: []gwapiv1a2.BackendRef{
						{
							BackendObjectReference: gwapiv1a2.BackendObjectReference{
								Name: gwapiv1a2.ObjectName(serviceName.Name),
								Port: ptr.To(gwapiv1a2.PortNumber(port)),
							},
						},
					},
				},
			},
		},
	}
}

// GetGatewayDeployment returns a sample Deployment for a Gateway object.
func GetGatewayDeployment(nsName types.NamespacedName, labels map[string]string) client.Object {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsName.Namespace,
			Name:      nsName.Name,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "dummy",
						Image: "dummy",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
						}},
					}},
				},
			},
		},
	}
}

// GetGatewayDaemonSet returns a sample DaemonSet for a Gateway object.
func GetGatewayDaemonSet(nsName types.NamespacedName, labels map[string]string) client.Object {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsName.Namespace,
			Name:      nsName.Name,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "dummy",
						Image: "dummy",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
						}},
					}},
				},
			},
		},
	}
}

// GetService returns a sample Service with labels and ports.
func GetService(nsName types.NamespacedName, labels map[string]string, ports map[string]int32) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsName.Name,
			Namespace: nsName.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{},
		},
	}
	for name, port := range ports {
		service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{
			Name: name,
			Port: port,
		})
	}
	return service
}

// GetConfigMap returns a sample ConfigMap with labels and data
func GetConfigMap(nsName types.NamespacedName, labels, data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsName.Name,
			Namespace: nsName.Namespace,
			Labels:    labels,
		},
		Data: data,
	}
}

// GetEndpointSlice returns a sample EndpointSlice.
func GetEndpointSlice(nsName types.NamespacedName, svcName string, isServiceImport bool) *discoveryv1.EndpointSlice {
	var labels map[string]string
	if isServiceImport {
		labels = map[string]string{mcsapiv1a1.LabelServiceName: svcName}
	} else {
		labels = map[string]string{discoveryv1.LabelServiceName: svcName}
	}

	return &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsName.Name,
			Namespace: nsName.Namespace,
			Labels:    labels,
		},
		Endpoints: []discoveryv1.Endpoint{
			{
				Addresses: []string{"10.0.0.1"},
				Conditions: discoveryv1.EndpointConditions{
					Ready: ptr.To(true),
				},
			},
		},
		Ports: []discoveryv1.EndpointPort{
			{
				Name:     &[]string{"dummy"}[0],
				Port:     &[]int32{8080}[0],
				Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
			},
		},
	}
}

// GetHTTPRouteFilter returns a sample Service with labels and ports.
func GetHTTPRouteFilter(nsName types.NamespacedName) *egv1a1.HTTPRouteFilter {
	return &egv1a1.HTTPRouteFilter{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsName.Name,
			Namespace: nsName.Namespace,
		},
		Spec: egv1a1.HTTPRouteFilterSpec{
			URLRewrite: &egv1a1.HTTPURLRewriteFilter{
				Path: &egv1a1.HTTPPathModifier{
					Type: egv1a1.RegexHTTPPathModifier,
					ReplaceRegexMatch: &egv1a1.ReplaceRegexMatch{
						Pattern:      "foo",
						Substitution: "bar",
					},
				},
			},
		},
	}
}

func GetClusterTrustBundle(name string) *certificatesv1b1.ClusterTrustBundle {
	return &certificatesv1b1.ClusterTrustBundle{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: certificatesv1b1.ClusterTrustBundleSpec{
			TrustBundle: "fake-trust-bundle",
		},
	}
}

func GetClientTrafficPolicy(nn types.NamespacedName, tls *egv1a1.ClientTLSSettings) *egv1a1.ClientTrafficPolicy {
	return &egv1a1.ClientTrafficPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nn.Name,
			Namespace: nn.Namespace,
		},
		Spec: egv1a1.ClientTrafficPolicySpec{
			TLS: tls,
		},
	}
}
