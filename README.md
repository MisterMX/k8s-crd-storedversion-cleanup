# K8s CRD Stored Version Cleanup

Small utility that removes CRD versions from `status.storedVersion`. It essentially cuts

```yaml
status:
    storedVersions:
        - v1alpha1
        - v1beta1
        - v1
```

down to 

```yaml
status:
    storedVersions:
        - v1
```

This allows updating the CRD to an newer version where `v1alpha1` has been removed.

Otherwise `kubectl` might respond with something like

```
The CustomResourceDefinition is invalid: status.storedVersions[0]: Invalid value: "v1alpha1": must appear in spec.versions
```

## Usage

1. Clone this repo
2. Run `go run cmd/k8s-crd-storedversion-cleanup/main.go`

Optional parameters:

* `--group <group>` the `spec.group` that should be filtered. Matches to the suffix so `api.io` goes for `a.api.io` and `b.api.io`
