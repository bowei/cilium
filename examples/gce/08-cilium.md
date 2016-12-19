# Setting up cilium daemon-set

Before we start any container we need to add cilium network plugin. We provide a
kubernetes daemon set that will run cilium on each gce node.

```
kubectl create -f https://raw.githubusercontent.com/cilium/cilium/master/examples/gce/daemon-set.yaml
```

Wait until the daemon set is deployed and check its status.

```
kubectl get daemonset cilium-net-controller
```

```
NAME                    DESIRED   CURRENT   READY     NODE-SELECTOR   AGE
cilium-net-controller   3         3         3         <none>          7m
```

## Network policies

Since cilium is already running on all workers, any pod that is started won't be able to
talk with any other pods nor the exterior. We need to create some kubernetes network
policy for that purpose.

Create the given 3 kubernetes network policies

```
kubectl create -f https://raw.githubusercontent.com/cilium/cilium/master/examples/gce/network-policy/admin-policy.json
kubectl create -f https://raw.githubusercontent.com/cilium/cilium/master/examples/gce/network-policy/lizards-policy-db.json
kubectl create -f https://raw.githubusercontent.com/cilium/cilium/master/examples/gce/network-policy/lizards-policy-web.json
```

Check if they were successfully created

```
kubectl get networkpolicy
```
NAME                 POD-SELECTOR   AGE
admin                kube-system=   9s
lizards-policy-db    db=            9s
lizards-policy-web   web=           9s
```

Add the cilium policy to all pods.

```
while read line; do
cat <<EOF | kubectl exec -i ${line} --  cilium -D policy import -
{
        "name": "io.cilium",
        "rules": [{
                "coverage": ["reserved:host"],
                "allow": ["reserved:all"]
        }]
}
EOF
done < <(kubectl get pods --output=jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}')
```

You can also check how the cilium policy tree looks like.

```
kubectl exec cilium-net-controller-4rj0l cilium policy dump
```
> The output is relatively large so we have omitted part of it.
```
{
  "name": "io.cilium",
  "rules": [
    {
      "coverage": [
        {
          "key": "host",
          "source": "reserved"
        }
      ],
      "allow": [
        {
          "action": "accept",
          "label": {
            "key": "all",
            "source": "reserved"
          }
        }
      ]
    }
  ],
  "children": {
...
```

From this moment on all pods created will be able to receive packets only if they have
the same set of labels set on the policy tree.
