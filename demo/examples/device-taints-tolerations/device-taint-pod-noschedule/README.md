# Device Taint Pod NoSchedule Example

`NoSchedule` prevents new allocations of devices with an untolerated taint.
Pods already using those devices keep running.

## Requirements

- Kubernetes 1.37 or later.
- Example driver deployed with the `gpu` profile and at least two available devices.

## How to Run

Run this example on its own, from this directory.

1. Create the ResourceClaimTemplate and first pod, then wait for the pod to be ready:

   Waiting ensures the first pod is already running when the taint is applied,
   so you can observe that `NoSchedule` does not evict it.

   ```bash
   kubectl apply -f 1-basic-resourceclaimtemplate.yaml
   kubectl apply -f 2-pod-keeps-running.yaml
   kubectl wait -n basic-resourceclaimtemplate \
     --for=condition=Ready pod/pod-keeps-running --timeout=120s
   ```

2. Apply the rule, which taints all devices managed by `gpu.example.com`, then create the second pod:

   ```bash
   kubectl apply -f 3-device-taint-rule.yaml
   kubectl apply -f 4-pod-not-scheduled.yaml
   ```

3. Observe the pods and scheduling events:

   ```bash
   kubectl get pods -n basic-resourceclaimtemplate
   kubectl get events -n basic-resourceclaimtemplate
   ```

## Expected Behavior

- `pod-keeps-running` stays Running because `NoSchedule` does not evict pods.
- `pod-no-schedule` stays Pending because it has no toleration for the device taint.

The second result assumes that the scheduler has read the rule before device allocation.
Otherwise, the second pod can also start and keep running.

```mermaid
graph TD
    subgraph Initial["Before Device Taint"]
        P1[Pod: pod-keeps-running<br/>Status: Running]
        G1[GPU Devices]
        P1 -.->|Using| G1
    end

    subgraph AfterTaint["After NoSchedule Taint"]
        P1R[Pod: pod-keeps-running<br/>Status: Running]
        P2[Pod: pod-no-schedule<br/>Status: Pending]
        G1T[GPU Devices<br/>Tainted: NoSchedule]
        P1R -.->|Keeps using| G1T
        P2 -.->|Allocation blocked| G1T
    end

    Initial --> AfterTaint

    style P1 fill:#d4edda,color:#000,stroke:#28a745,stroke-width:2px
    style P1R fill:#d4edda,color:#000,stroke:#28a745,stroke-width:2px
    style P2 fill:#f8d7da,color:#000,stroke:#dc3545,stroke-width:2px
    style G1 fill:#9b59b6,color:#fff,stroke:#8e44ad,stroke-width:3px
    style G1T fill:#e74c3c,color:#fff,stroke:#c0392b,stroke-width:3px
```

## Cleanup

```bash
kubectl delete -f 4-pod-not-scheduled.yaml
kubectl delete -f 3-device-taint-rule.yaml
kubectl delete -f 2-pod-keeps-running.yaml
kubectl delete -f 1-basic-resourceclaimtemplate.yaml
```
