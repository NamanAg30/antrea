// Copyright 2022 Antrea Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package monitor

import (
	"testing"
	"time"

	fakeclientset "antrea.io/antrea/pkg/client/clientset/versioned/fake"
	crdinformers "antrea.io/antrea/pkg/client/informers/externalversions"
	"github.com/stretchr/testify/assert"
)

const (
	informerDefaultResync = 12 * time.Hour
)

func TestSyncExternalNode(t *testing.T) {

	tc := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "Invalid key format",
			key:      "namespace/name/error",
			expected: "unexpected key format: \"namespace/name/error\"",
		},
		{
			name:     "Key does not exist",
			key:      "ns1/vm2-e8be5",
			expected: "",
		},
	}

	crdClient := fakeclientset.NewSimpleClientset()
	crdInformerFactory := crdinformers.NewSharedInformerFactory(crdClient, informerDefaultResync)
	externalNodeInformer := crdInformerFactory.Crd().V1alpha1().ExternalNodes()

	externalNodeLister := externalNodeInformer.Lister()
	controller := &controllerMonitor{
		client:             crdClient,
		externalNodeLister: externalNodeLister,
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			generated := controller.syncExternalNode(tt.key)
			if tt.expected == "" {
				assert.NoError(t, generated)
			} else {
				assert.Equal(t, tt.expected, generated.Error())
			}
		})
	}
}

func TestDeleteAgentCRD(t *testing.T) {

	tc := []struct {
		test_name string
		name      string
		expected  string
	}{
		{
			test_name: "CRD Agent with given name does not exist",
			name:      "vm2-e8be5",
			expected:  "",
		},
	}

	crdClient := fakeclientset.NewSimpleClientset()
	crdInformerFactory := crdinformers.NewSharedInformerFactory(crdClient, informerDefaultResync)
	externalNodeInformer := crdInformerFactory.Crd().V1alpha1().ExternalNodes()
	externalNodeLister := externalNodeInformer.Lister()
	controller := &controllerMonitor{
		client:             crdClient,
		externalNodeLister: externalNodeLister,
	}

	for _, tt := range tc {
		t.Run(tt.test_name, func(t *testing.T) {
			generated := controller.deleteAgentCRD(tt.name)
			if tt.expected == "" {
				assert.NoError(t, generated)
			} else {
				assert.Equal(t, tt.expected, generated.Error())
			}
		})
	}

}
