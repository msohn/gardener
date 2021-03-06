// This file was automatically generated by informer-gen

package v1beta1

import (
	internalinterfaces "github.com/gardener/gardener/pkg/client/garden/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// CloudProfiles returns a CloudProfileInformer.
	CloudProfiles() CloudProfileInformer
	// CrossSecretBindings returns a CrossSecretBindingInformer.
	CrossSecretBindings() CrossSecretBindingInformer
	// PrivateSecretBindings returns a PrivateSecretBindingInformer.
	PrivateSecretBindings() PrivateSecretBindingInformer
	// Quotas returns a QuotaInformer.
	Quotas() QuotaInformer
	// Seeds returns a SeedInformer.
	Seeds() SeedInformer
	// Shoots returns a ShootInformer.
	Shoots() ShootInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// CloudProfiles returns a CloudProfileInformer.
func (v *version) CloudProfiles() CloudProfileInformer {
	return &cloudProfileInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// CrossSecretBindings returns a CrossSecretBindingInformer.
func (v *version) CrossSecretBindings() CrossSecretBindingInformer {
	return &crossSecretBindingInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// PrivateSecretBindings returns a PrivateSecretBindingInformer.
func (v *version) PrivateSecretBindings() PrivateSecretBindingInformer {
	return &privateSecretBindingInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Quotas returns a QuotaInformer.
func (v *version) Quotas() QuotaInformer {
	return &quotaInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Seeds returns a SeedInformer.
func (v *version) Seeds() SeedInformer {
	return &seedInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// Shoots returns a ShootInformer.
func (v *version) Shoots() ShootInformer {
	return &shootInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
