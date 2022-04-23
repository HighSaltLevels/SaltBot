""" Interface to k8s configmaps from within k8s """
import os

import kubernetes
from kubernetes.client.models.v1_config_map import V1ConfigMap
from kubernetes.client.models.v1_object_meta import V1ObjectMeta

SA_TOKEN_PATH = "/var/run/secrets/kubernetes.io/serviceaccount/token"
CA_CERT_PATH = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

NAMESPACE = "saltbot"


class ConfigMap:
    """K8s configmap representation class"""

    def __init__(self):
        with open(SA_TOKEN_PATH) as _file:
            token = _file.read().strip()

        host = os.getenv("KUBERNETES_SERVICE_HOST")
        port = os.getenv("KUBERNETES_SERVICE_PORT")
        url = f"https://{host}:{port}"
        configuration = kubernetes.client.Configuration(url)
        configuration.ssl_ca_cert = CA_CERT_PATH
        configuration.api_key["authorization"] = token
        configuration.api_key_prefix["authorization"] = "Bearer"
        with kubernetes.client.ApiClient(configuration) as api_client:
            self._api = kubernetes.client.CoreV1Api(api_client)

    def list(self, label_selector=None):
        """List configmaps by selector label"""
        resp = self._api.list_namespaced_config_map(
            NAMESPACE, label_selector=label_selector
        )
        return [item.data for item in resp.items]

    def create(self, name, labels, data):
        """Create a new config map"""
        metadata = V1ObjectMeta(name=name, labels=labels)
        config_map = V1ConfigMap(data=data, metadata=metadata)
        self._api.create_namespaced_config_map(NAMESPACE, config_map)

    def delete(self, name):
        """Delete a config map by name"""
        self._api.delete_namespaced_config_map(name, NAMESPACE)

    def get(self, name):
        """Get a config map by name"""
        resp = self._api.read_namespaced_config_map(name, NAMESPACE)
        return resp.data

    def patch(self, name, data):
        """Patch an existing config map by name"""
        body = {"data": data}
        self._api.patch_namespaced_config_map(name, NAMESPACE, body)
