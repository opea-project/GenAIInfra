# Authentication and authorization

Access to enterprise AI workloads requires robust authentication and authorization to ensure data security and protect sensitive information. This is critical for enterprises because it prevents unauthorized access, which could lead to data breaches, intellectual property theft, and misuse of AI resources. By implementing strict access controls, enterprises can maintain compliance with regulatory standards, safeguard their competitive edge, and ensure that only authorized personnel can manage and utilize AI systems. This not only protects the organization’s assets but also builds trust with clients and stakeholders, reinforcing the enterprise’s commitment to security and integrity.

Here we provide different options for user to implement authentication and authorization according to their needs:

## Istio based implementation for cloud native environments

Utilize Istio, a well-known and widely adopted cloud-native tool, to enhance authentication and authorization processes. This ensures secure, efficient, and reliable enterprise operations by managing access controls effectively. This solution can integrate directly with OIDC providers like Keycloak or utilize an authentication proxy such as oauth2-proxy, enhancing its flexibility and scalability. We provide support for both helm chart and GMC based GenAI workload deployment to handle the authentication and authorization with either strict or flexible policies, adapting to various security requirements and operational needs. Please refer the documentation [here](./auth-istio/README.md) for more information.
