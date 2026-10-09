# Domain migration

These notes apply to installations that use the previous `example.com` names and API groups.
New installations can follow the [README](../README.md).

## Name changes

This change replaces the example domain with `dra-example-driver.sigs.k8s.io`.
In the table, `<profile>` is `gpu`, `cpu`, or `net`.

| Item | Previous name | New name |
| --- | --- | --- |
| Default driver and DeviceClass | `<profile>.example.com` | `<profile>.dra-example-driver.sigs.k8s.io` |
| GPU configuration API | `gpu.resource.example.com/v1alpha1` | `gpu.resource.dra-example-driver.sigs.k8s.io/v1alpha1` |
| Network configuration API | `net.resource.example.com/v1alpha1` | `net.resource.dra-example-driver.sigs.k8s.io/v1alpha1` |
| Checkpoint API | `checkpoint.internal.example.com/v1` | `checkpoint.internal.dra-example-driver.sigs.k8s.io/v1` |
| Health annotation prefix | `health.example.com/` | `health.dra-example-driver.sigs.k8s.io/` |
| Validating webhook | `dra.example.com` | `dra.dra-example-driver.sigs.k8s.io` |
| Example extended resource | `example.com/gpu` | `dra-example-driver.sigs.k8s.io/gpu` |
| Go configuration API packages | `api/example.com/` | `api/dra-example-driver.sigs.k8s.io/` |

Driver developers must update their Go imports.

## Upgrade notes

Existing claim allocations do not transfer to the new driver name.
The driver does not accept the previous configuration API groups or checkpoint group.
The previous health annotation prefix no longer controls device health.
Setting `driverName` to the previous name does not preserve API or checkpoint compatibility.

### Cleanup and reinstall

1. Stop the workloads that use the old driver. Keep the driver running until their
   ResourceClaims are deleted and devices are released. Check all namespaces.
   Do not force claim deletion or remove finalizers.
2. Uninstall the old driver. Remove only its DeviceClasses and ResourceSlices.
   Keep namespaces and resources used by other workloads or drivers.
3. After the old driver stops, remove its checkpoint on each node:
   `/var/lib/kubelet/plugins/<old-driver-name>/checkpoint.json`.
   Use the configured plugin directory if it differs from the default.
   The new default name uses a different directory, so the old file does not affect
   the new installation. Removing it prevents reuse of old state if you restore
   the old name, including during a downgrade.
4. Update manifests and Helm values to use the new names. Install the new driver,
   then recreate the claims and workloads. Check that Pods are ready and claims
   are allocated to the new driver.

Use the same cleanup order before a downgrade across this API group change.
Install the target version's image, chart, and manifests together.
