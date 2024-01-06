package deploymentscaling

import (
	"k8s.io/apimachinery/pkg/runtime"
)

func (in *DeploymentScaling) DeepCopyInto(out *DeploymentScaling) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

func (in *DeploymentScaling) DeepCopy() *DeploymentScaling {
	if in == nil {
		return nil
	}
	out := new(DeploymentScaling)
	in.DeepCopyInto(out)
	return out
}

func (in *DeploymentScaling) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *DeploymentScalingList) DeepCopyInto(out *DeploymentScalingList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DeploymentScaling, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *DeploymentScalingList) DeepCopy() *DeploymentScalingList {
	if in == nil {
		return nil
	}
	out := new(DeploymentScalingList)
	in.DeepCopyInto(out)
	return out
}

func (in *DeploymentScalingList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
