// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package k8s

import (
	"sigs.k8s.io/network-policy-api/apis/v1alpha1"
)

func (p *policyWatcher) addK8sAdminNewtorkPolicy(k8sANP *v1alpha1.AdminNetworkPolicy) error {
	p.log.Infof("XXX/bowei add_K8sAdminNewtorkPolicy()")

	return nil
}

func (p *policyWatcher) deleteK8sAdminNewtorkPolicy() error {
	p.log.Infof("XXX/bowei delete_K8sAdminNewtorkPolicy()")

	return nil
}
