package fake

import (
	v1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeShoots implements ShootInterface
type FakeShoots struct {
	Fake *FakeGardenV1beta1
	ns   string
}

var shootsResource = schema.GroupVersionResource{Group: "garden.sapcloud.io", Version: "v1beta1", Resource: "shoots"}

var shootsKind = schema.GroupVersionKind{Group: "garden.sapcloud.io", Version: "v1beta1", Kind: "Shoot"}

// Get takes name of the shoot, and returns the corresponding shoot object, and an error if there is any.
func (c *FakeShoots) Get(name string, options v1.GetOptions) (result *v1beta1.Shoot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(shootsResource, c.ns, name), &v1beta1.Shoot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Shoot), err
}

// List takes label and field selectors, and returns the list of Shoots that match those selectors.
func (c *FakeShoots) List(opts v1.ListOptions) (result *v1beta1.ShootList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(shootsResource, shootsKind, c.ns, opts), &v1beta1.ShootList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.ShootList{}
	for _, item := range obj.(*v1beta1.ShootList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested shoots.
func (c *FakeShoots) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(shootsResource, c.ns, opts))

}

// Create takes the representation of a shoot and creates it.  Returns the server's representation of the shoot, and an error, if there is any.
func (c *FakeShoots) Create(shoot *v1beta1.Shoot) (result *v1beta1.Shoot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(shootsResource, c.ns, shoot), &v1beta1.Shoot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Shoot), err
}

// Update takes the representation of a shoot and updates it. Returns the server's representation of the shoot, and an error, if there is any.
func (c *FakeShoots) Update(shoot *v1beta1.Shoot) (result *v1beta1.Shoot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(shootsResource, c.ns, shoot), &v1beta1.Shoot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Shoot), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeShoots) UpdateStatus(shoot *v1beta1.Shoot) (*v1beta1.Shoot, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(shootsResource, "status", c.ns, shoot), &v1beta1.Shoot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Shoot), err
}

// Delete takes name of the shoot and deletes it. Returns an error if one occurs.
func (c *FakeShoots) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(shootsResource, c.ns, name), &v1beta1.Shoot{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeShoots) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(shootsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1beta1.ShootList{})
	return err
}

// Patch applies the patch and returns the patched shoot.
func (c *FakeShoots) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.Shoot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(shootsResource, c.ns, name, data, subresources...), &v1beta1.Shoot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Shoot), err
}
