//go:build e2e

/*
 * Copyright The Kubernetes Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Serial prevents collisions with tests that install the same DeviceClass.
var _ = Describe("Default driver names", Serial, func() {
	for _, profile := range []string{"gpu", "cpu", "net"} {
		It("should allocate devices across a rolling update with the "+profile+" default name", func(ctx SpecContext) {
			By("installing the driver with its default name")
			drv := installDriver(ctx, DriverConfig{
				UseDefaultDriverName: true,
				ExtraValues: map[string]string{
					"deviceProfile":                     profile,
					"kubeletPlugin.cpu.cpusPerNUMANode": "1",
				},
			})
			Expect(drv.DriverName).To(Equal(profile + ".dra-example-driver.sigs.k8s.io"))

			By("allocating a device before the rolling update")
			claim := createDeviceClaim(ctx, drv, "before-update")
			pod := startPodWithClaim(ctx, drv, "before-update", claim.Name, "")
			beforeUpdate, err := clientset.ResourceV1().ResourceClaims(drv.Namespace).Get(ctx, claim.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(beforeUpdate.Status.Allocation).NotTo(BeNil())
			Expect(beforeUpdate.Status.Allocation.Devices.Results).NotTo(BeEmpty())
			for _, result := range beforeUpdate.Status.Allocation.Devices.Results {
				Expect(result.Driver).To(Equal(drv.DriverName))
			}

			By("rolling the driver while the workload holds its claim")
			rollDriver(ctx, drv)

			// Node-allocatable CPU claims cannot be shared between Pods.
			// GPU and network claims exercise prepare after checkpoint recovery.
			if profile != "cpu" {
				By("preparing the existing claim on the same node through the new driver Pod")
				startPodWithClaim(ctx, drv, "recovered", claim.Name, pod.Spec.NodeName)
			}

			By("checking that the workload and its allocation survive the update")
			afterUpdate, err := clientset.ResourceV1().ResourceClaims(drv.Namespace).Get(ctx, claim.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(afterUpdate.Status.Allocation).To(Equal(beforeUpdate.Status.Allocation))
			checkPodsReadyAndRunning(ctx, drv.Namespace, []string{pod.Name})

			By("allocating a new device after the update")
			newClaim := createDeviceClaim(ctx, drv, "after-update")
			startPodWithClaim(ctx, drv, "after-update", newClaim.Name, "")

			By("checking driver health across multiple liveness probes")
			verifyDriverRemainsHealthy(ctx, drv)
		})
	}
})
