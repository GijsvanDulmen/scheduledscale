package cronjobsuspend

import (
	"k8s.io/apimachinery/pkg/runtime"
)

func (in *CronJobSuspend) DeepCopyInto(out *CronJobSuspend) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

func (in *CronJobSuspend) DeepCopy() *CronJobSuspend {
	if in == nil {
		return nil
	}
	out := new(CronJobSuspend)
	in.DeepCopyInto(out)
	return out
}

func (in *CronJobSuspend) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *CronJobSuspendList) DeepCopyInto(out *CronJobSuspendList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CronJobSuspend, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *CronJobSuspendList) DeepCopy() *CronJobSuspendList {
	if in == nil {
		return nil
	}
	out := new(CronJobSuspendList)
	in.DeepCopyInto(out)
	return out
}

func (in *CronJobSuspendList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
