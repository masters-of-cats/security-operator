This is a K8S operator that automatically binds a subject to a pod security policy via a `PodSecurityPolicyBinding` custom resource

# How to use it
1. Setup a dedicated `operator` namespace and configure the operator into it:

```
kubectl -n operator apply -f deploy/setup-operator.yaml
```

2. Create the "deployer" namespace and setup the deployer service account (in our example the namespace is called `foo`)
```
kubectl -n foo apply -f deploy/deployer_service_account.yaml
```

4. Create the `PodSecurityPolicyBinding` custom resource into the `operator` namespace (so that the operator configures the `deployer` user with the pod security policy)
```
kubectl -n operator apply -f deploy/crds/map_deployer_to_podsecurity_cr.yaml
```

5. Create a pod as `deployer`
```
kubectl -n foo --as system:serviceaccount:foo:deployer apply -f deploy/pod.yml
```

6. Success
The pod has both seccomp and apparomour profiles enabled
