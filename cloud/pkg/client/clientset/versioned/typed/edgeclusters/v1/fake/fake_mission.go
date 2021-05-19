/*
Copyright The KubeEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	edgeclustersv1 "github.com/kubeedge/kubeedge/cloud/pkg/apis/edgeclusters/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeMissions implements MissionInterface
type FakeMissions struct {
	Fake *FakeEdgeclustersV1
}

var missionsResource = schema.GroupVersionResource{Group: "edgeclusters.kubeedge.io", Version: "v1", Resource: "missions"}

var missionsKind = schema.GroupVersionKind{Group: "edgeclusters.kubeedge.io", Version: "v1", Kind: "Mission"}

// Get takes name of the mission, and returns the corresponding mission object, and an error if there is any.
func (c *FakeMissions) Get(ctx context.Context, name string, options v1.GetOptions) (result *edgeclustersv1.Mission, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(missionsResource, name), &edgeclustersv1.Mission{})
	if obj == nil {
		return nil, err
	}
	return obj.(*edgeclustersv1.Mission), err
}

// List takes label and field selectors, and returns the list of Missions that match those selectors.
func (c *FakeMissions) List(ctx context.Context, opts v1.ListOptions) (result *edgeclustersv1.MissionList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(missionsResource, missionsKind, opts), &edgeclustersv1.MissionList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &edgeclustersv1.MissionList{ListMeta: obj.(*edgeclustersv1.MissionList).ListMeta}
	for _, item := range obj.(*edgeclustersv1.MissionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested missions.
func (c *FakeMissions) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(missionsResource, opts))
}

// Create takes the representation of a mission and creates it.  Returns the server's representation of the mission, and an error, if there is any.
func (c *FakeMissions) Create(ctx context.Context, mission *edgeclustersv1.Mission, opts v1.CreateOptions) (result *edgeclustersv1.Mission, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(missionsResource, mission), &edgeclustersv1.Mission{})
	if obj == nil {
		return nil, err
	}
	return obj.(*edgeclustersv1.Mission), err
}

// Update takes the representation of a mission and updates it. Returns the server's representation of the mission, and an error, if there is any.
func (c *FakeMissions) Update(ctx context.Context, mission *edgeclustersv1.Mission, opts v1.UpdateOptions) (result *edgeclustersv1.Mission, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(missionsResource, mission), &edgeclustersv1.Mission{})
	if obj == nil {
		return nil, err
	}
	return obj.(*edgeclustersv1.Mission), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMissions) UpdateStatus(ctx context.Context, mission *edgeclustersv1.Mission, opts v1.UpdateOptions) (*edgeclustersv1.Mission, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(missionsResource, "status", mission), &edgeclustersv1.Mission{})
	if obj == nil {
		return nil, err
	}
	return obj.(*edgeclustersv1.Mission), err
}

// Delete takes name of the mission and deletes it. Returns an error if one occurs.
func (c *FakeMissions) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(missionsResource, name), &edgeclustersv1.Mission{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMissions) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(missionsResource, listOpts)

	_, err := c.Fake.Invokes(action, &edgeclustersv1.MissionList{})
	return err
}

// Patch applies the patch and returns the patched mission.
func (c *FakeMissions) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *edgeclustersv1.Mission, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(missionsResource, name, pt, data, subresources...), &edgeclustersv1.Mission{})
	if obj == nil {
		return nil, err
	}
	return obj.(*edgeclustersv1.Mission), err
}
