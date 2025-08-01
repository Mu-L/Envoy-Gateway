---
title: "Backend TLS: Skip TLS Verification"
---

This task demonstrates how to skip TLS verification for a backend service in Envoy Gateway.

By default, you must configure a [BackendTLSPolicy][] to validate the TLS certificate of a backend service when it uses TLS.

However, in certain scenarios—such as development or testing—you might want to skip TLS verification for a backend service.
To do this, you can use the [Backend][] API and set the `tls.insecureSkipVerify` field to true in the [Backend][] resource.

Warning: Skipping TLS verification disables certificate validation, which can expose your connection to man-in-the-middle
attacks. This setting is typically used only for development or testing and is not recommended in production environments.

## Prerequisites

- OpenSSL to generate TLS assets.

## Installation

{{< boilerplate prerequisites >}}

## Enable Backend

The [Backend][] API is disabled by default in Envoy Gateway. To enable it, follow the steps outlined in [Backend Routing][] to configure the [EnvoyGateway][] startup settings accordingly.

## TLS Certificates

Generate the certificates and keys used by the backend to terminate TLS connections from the Gateways.

Create a root certificate and private key to sign certificates:

```shell
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=example Inc./CN=example.com' -keyout ca.key -out ca.crt
```

Create a certificate and a private key for `www.example.com`.

First, create an openssl configuration file:

```shell
cat > openssl.conf  <<EOF
[req]
req_extensions = v3_req
prompt = no

[v3_req]
keyUsage = keyEncipherment, digitalSignature
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = www.example.com
EOF
```

Then create a certificate using this openssl configuration file:

```shell
openssl req -out www.example.com.csr -newkey rsa:2048 -nodes -keyout www.example.com.key -subj "/CN=www.example.com/O=example organization"
openssl x509 -req -days 365 -CA ca.crt -CAkey ca.key -set_serial 0 -in www.example.com.csr -out www.example.com.crt -extfile openssl.conf -extensions v3_req
```

Note that the certificate must contain a DNS SAN for the relevant domain.

Store the cert/key in a Secret:

```shell
kubectl create secret tls example-cert --key=www.example.com.key --cert=www.example.com.crt
```

Store the CA Cert in another Secret:

```shell
kubectl create configmap example-ca --from-file=ca.crt
```

## Setup TLS on the backend

Patch the existing quickstart backend to enable TLS. The patch will mount the TLS certificate secret into the backend as volume.

```shell
kubectl patch deployment backend --type=json --patch '
  - op: add
    path: /spec/template/spec/containers/0/volumeMounts
    value:
    - name: secret-volume
      mountPath: /etc/secret-volume
  - op: add
    path: /spec/template/spec/volumes
    value:
    - name: secret-volume
      secret:
        secretName: example-cert
        items:
        - key: tls.crt
          path: crt
        - key: tls.key
          path: key
  - op: add
    path: /spec/template/spec/containers/0/env/-
    value:
      name: TLS_SERVER_CERT
      value: /etc/secret-volume/crt
  - op: add
    path: /spec/template/spec/containers/0/env/-
    value:
      name: TLS_SERVER_PRIVKEY
      value: /etc/secret-volume/key
  '
```

Create a service that exposes port 443 on the backend service.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  labels:
    app: backend
    service: backend
  name: tls-backend
  namespace: default
spec:
  selector:
    app: backend
  ports:
  - name: https
    port: 443
    protocol: TCP
    targetPort: 8443
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: backend
    service: backend
  name: tls-backend
  namespace: default
spec:
  selector:
    app: backend
  ports:
  - name: https
    port: 443
    protocol: TCP
    targetPort: 8443
```

{{% /tab %}}
{{< /tabpane >}}

## Skip TLS Verification

In the following example, we will create a [Backend][] resource that routes to the `tls-backend` service, and an [HTTPRoute][]
that references the [Backend][] resource.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - backendRefs:
        - group: gateway.envoyproxy.io
          kind: Backend
          name: tls-backend
      matches:
        - path:
            type: PathPrefix
            value: /
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: tls-backend
  namespace: default
spec:
  endpoints:
    - fqdn:
        hostname: tls-backend.default.svc.cluster.local
        port: 443
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resources to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - backendRefs:
        - group: gateway.envoyproxy.io
          kind: Backend
          name: tls-backend
      matches:
        - path:
            type: PathPrefix
            value: /
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: tls-backend
  namespace: default
spec:
  endpoints:
    - fqdn:
        hostname: tls-backend.default.svc.cluster.local
        port: 443
```

{{% /tab %}}
{{< /tabpane >}}

Send a request to the backend service:

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:80:${GATEWAY_HOST}" \
http://www.example.com:80/get
```

Because the backend service is using TLS, and no [BackendTLSPolicy][] is configured to validate the TLS certificate,
the request will fail with a `400 Bad Request` error:

```bash
* Connected to www.example.com (172.18.0.200) port 80
* using HTTP/1.x
> GET /get HTTP/1.1
> Host:www.example.com
> User-Agent: curl/8.14.1
> Accept: */*
>
* Request completely sent off
< HTTP/1.1 400 Bad Request
< date: Thu, 31 Jul 2025 06:09:51 GMT
< transfer-encoding: chunked
```

Disabling TLS verification by setting the `InsecureSkipVerify` field to `true` allows the request to succeed:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: tls-backend
  namespace: default
spec:
  endpoints:
    - fqdn:
        hostname: tls-backend.default.svc.cluster.local
        port: 443
  tls:
    insecureSkipVerify: true
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: tls-backend
  namespace: default
spec:
  endpoints:
    - fqdn:
        hostname: tls-backend.default.svc.cluster.local
        port: 443
  tls:
    insecureSkipVerify: true
```

{{% /tab %}}
{{< /tabpane >}}

Send the request again:

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:80:${GATEWAY_HOST}" \
http://www.example.com:80/get
```

You should now see a successful response from the backend service. Since TLS verification has been skipped, the request
proceeds without validating the backend’s TLS certificate. The response will include TLS details such as the protocol
version and cipher suite used for the connection.


```shell
< HTTP/1.1 200 OK
[...]
 "tls": {
  "version": "TLSv1.2",
  "serverName": "",
  "negotiatedProtocol": "http/1.1",
  "cipherSuite": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
 }
```

[Backend Routing]: ../traffic/backend/#enable-backend
[Backend]: ../../../api/extension_types#backend
[EnvoyGateway]: ../../../api/extension_types#envoygateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
[BackendTLSPolicy]: https://gateway-api.sigs.k8s.io/api-types/backendtlspolicy/
