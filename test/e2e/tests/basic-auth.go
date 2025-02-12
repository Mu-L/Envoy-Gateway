// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func init() {
	ConformanceTests = append(ConformanceTests, BasicAuthTest)
}

var BasicAuthTest = suite.ConformanceTest{
	ShortName:   "BasicAuth",
	Description: "Resource with BasicAuth enabled",
	Manifests:   []string{"testdata/basic-auth.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("valid username password", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-basic-auth-1", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "basic-auth-1", Namespace: ns})
			// TODO: We should wait for the `programmed` condition to be true before sending traffic.
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/basic-auth-1",
					Headers: map[string]string{
						"Authorization": "Basic dXNlcjE6dGVzdDE=", // user1:test1
					},
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}
		})

		t.Run("without Authorization header", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-basic-auth-1", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "basic-auth-1", Namespace: ns})
			// TODO: We should wait for the `programmed` condition to be true before sending traffic.
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/basic-auth-1",
				},
				Response: http.Response{
					StatusCode: 401,
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}
		})

		t.Run("invalid username password", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-basic-auth-1", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "basic-auth-1", Namespace: ns})
			// TODO: We should wait for the `programmed` condition to be true before sending traffic.
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/basic-auth-1",
					Headers: map[string]string{
						"Authorization": "Basic dXNlcjE6dGVzdDI=", // user1:test2
					},
				},
				Response: http.Response{
					StatusCode: 401,
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}
		})

		t.Run("per route configuration second route", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-basic-auth-2", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "basic-auth-2", Namespace: ns})
			// TODO: We should wait for the `programmed` condition to be true before sending traffic.
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/basic-auth-2",
					Headers: map[string]string{
						"Authorization": "Basic dXNlcjQ6dGVzdDQ=", // user4:test4
					},
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}
		})
	},
}

// SecurityPolicyMustBeAccepted waits for the specified SecurityPolicy to be accepted.
func SecurityPolicyMustBeAccepted(
	t *testing.T,
	client client.Client,
	securityPolicyName types.NamespacedName) {
	t.Helper()

	waitErr := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
		securityPolicy := &egv1a1.SecurityPolicy{}
		err := client.Get(ctx, securityPolicyName, securityPolicy)
		if err != nil {
			return false, fmt.Errorf("error fetching SecurityPolicy: %w", err)
		}

		for _, condition := range securityPolicy.Status.Conditions {
			if condition.Type == string(gwv1a2.PolicyConditionAccepted) && condition.Status == metav1.ConditionTrue {
				return true, nil
			}
		}
		t.Logf("SecurityPolicy not yet accepted: %v", securityPolicy)
		return false, nil
	})
	require.NoErrorf(t, waitErr, "error waiting for HTTPRoute to have parents matching expectations")
}
