//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceCertSignerOperatorConfig) DeepCopyInto(out *ServiceCertSignerOperatorConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceCertSignerOperatorConfig.
func (in *ServiceCertSignerOperatorConfig) DeepCopy() *ServiceCertSignerOperatorConfig {
	if in == nil {
		return nil
	}
	out := new(ServiceCertSignerOperatorConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceCertSignerOperatorConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceCertSignerOperatorConfigList) DeepCopyInto(out *ServiceCertSignerOperatorConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ServiceCertSignerOperatorConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceCertSignerOperatorConfigList.
func (in *ServiceCertSignerOperatorConfigList) DeepCopy() *ServiceCertSignerOperatorConfigList {
	if in == nil {
		return nil
	}
	out := new(ServiceCertSignerOperatorConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceCertSignerOperatorConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceCertSignerOperatorConfigSpec) DeepCopyInto(out *ServiceCertSignerOperatorConfigSpec) {
	*out = *in
	in.OperatorSpec.DeepCopyInto(&out.OperatorSpec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceCertSignerOperatorConfigSpec.
func (in *ServiceCertSignerOperatorConfigSpec) DeepCopy() *ServiceCertSignerOperatorConfigSpec {
	if in == nil {
		return nil
	}
	out := new(ServiceCertSignerOperatorConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceCertSignerOperatorConfigStatus) DeepCopyInto(out *ServiceCertSignerOperatorConfigStatus) {
	*out = *in
	in.OperatorStatus.DeepCopyInto(&out.OperatorStatus)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceCertSignerOperatorConfigStatus.
func (in *ServiceCertSignerOperatorConfigStatus) DeepCopy() *ServiceCertSignerOperatorConfigStatus {
	if in == nil {
		return nil
	}
	out := new(ServiceCertSignerOperatorConfigStatus)
	in.DeepCopyInto(out)
	return out
}
