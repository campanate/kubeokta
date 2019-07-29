# Kubeokta

A cli that helps you connect your kubernetes cluster with your Okta user. Requirements:
1. An Okta OIDC application with password as grant type.
2. Your cluster kubernetes with OIDC configured.
3. A kubeconfig file with certificate-authority-data configured.

```
kubeokta [OPTIONS]

Application Options:
      --cluster=       Kubernetes cluster for okta authentication. [$K8S_CLUSTER]
      --okta-user=     Okta user for authentication. [$OKTA_USER]
      --issuer-url=    Issuer URL of your okta authorization server. [$ISSUER_URL]
      --client-id=     Client ID of your OIDC Okta application. [$CLIENT_ID]
      --client-secret= CLient Secret of your OIDC Okta application. [$CLIENT_SECRET]

Help Options:
  -h, --help           Show this help message
```