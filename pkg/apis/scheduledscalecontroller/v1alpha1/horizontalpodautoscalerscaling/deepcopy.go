package horizontalpodautoscalerscaling

import (
	"k8s.io/apimachinery/pkg/runtime"
)

func (in *HorizontalPodAutoscalerScaling) DeepCopyInto(out *HorizontalPodAutoscalerScaling) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

func (in *HorizontalPodAutoscalerScaling) DeepCopy() *HorizontalPodAutoscalerScaling {
	if in == nil {
		return nil
	}
	out := new(HorizontalPodAutoscalerScaling)
	in.DeepCopyInto(out)
	return out
}

func (in *HorizontalPodAutoscalerScaling) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *HorizontalPodAutoscalerScalingList) DeepCopyInto(out *HorizontalPodAutoscalerScalingList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]HorizontalPodAutoscalerScaling, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *HorizontalPodAutoscalerScalingList) DeepCopy() *HorizontalPodAutoscalerScalingList {
	if in == nil {
		return nil
	}
	out := new(HorizontalPodAutoscalerScalingList)
	in.DeepCopyInto(out)
	return out
}

func (in *HorizontalPodAutoscalerScalingList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
