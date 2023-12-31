package informer

import (
	"context"
	"fmt"
	v1 "k8s.io/api/apps/v1"
	pv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"log"
	"vandulmen.net/scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/deploymentscaling"
)

func (informer *Informer) DeletePodDisruptionBudgetsFor(ds *deploymentscaling.DeploymentScaling) error {
	pdbListOptions := metav1.ListOptions{
		LabelSelector: labels.Set(map[string]string{
			"scheduledscale.vandulmen.net/owner": ds.Name,
		}).String(),
	}

	pdbList, err := informer.coreClientSet.
		PolicyV1().
		PodDisruptionBudgets(ds.Namespace).
		List(context.TODO(), pdbListOptions)

	if err != nil {
		return err
	}

	for _, pdb := range pdbList.Items {
		deletePdb := true
		if ds.Spec.OnDelete != nil {
			if ds.Spec.OnDelete.RemovePodDisruptionBudget != nil {
				deletePdb = *ds.Spec.OnDelete.RemovePodDisruptionBudget
			}
		}

		if deletePdb {
			log.Println("Deleting PDB %s for %s in %s", pdb.Name, ds.Name, ds.Namespace)
			_ = informer.coreClientSet.PolicyV1().PodDisruptionBudgets(ds.Namespace).
				Delete(context.TODO(), pdb.Name, metav1.DeleteOptions{})
		} else {
			log.Println("Removing owner and labels of PDB %s for %s in %s", pdb.Name, ds.Name, ds.Namespace)
			pdb.OwnerReferences = []metav1.OwnerReference{}
			delete(pdb.Labels, "scheduledscale.vandulmen.net/deployment")
			delete(pdb.Labels, "scheduledscale.vandulmen.net/owner")

			informer.coreClientSet.PolicyV1().PodDisruptionBudgets(ds.Namespace).
				Update(context.TODO(), &pdb, metav1.UpdateOptions{})
		}
	}
	return nil
}

func (informer *Informer) ReconcilePodDisruptionBudget(scaleTo *deploymentscaling.ScaleTo, ds *deploymentscaling.DeploymentScaling, deployment *v1.Deployment) {
	if scaleTo.PodDisruptionBudget != nil {
		pdbListOptions := metav1.ListOptions{
			LabelSelector: labels.Set(map[string]string{
				"scheduledscale.vandulmen.net/deployment": deployment.Name,
			}).String(),
		}

		pdbList, err := informer.coreClientSet.
			PolicyV1().
			PodDisruptionBudgets(ds.Namespace).
			List(context.TODO(), pdbListOptions)

		if err != nil {
			errorMessage := fmt.Sprintf("Could not list pdbs for %s in %s - not doing anything", ds.Name, ds.Namespace)
			log.Println(errorMessage)
			ds.Status.ErrorMessage = errorMessage
			return
		} else {
			if len(pdbList.Items) > 1 {
				errorMessage := fmt.Sprintf("Multiple pdbs found for %s in %s - not doing anything", ds.Name, ds.Namespace)
				log.Println(errorMessage)
				ds.Status.ErrorMessage = errorMessage
				return
			} else if len(pdbList.Items) > 0 {
				err := informer.coreClientSet.PolicyV1().PodDisruptionBudgets(ds.Namespace).
					Delete(context.TODO(), pdbList.Items[0].Name, metav1.DeleteOptions{})
				if err != nil {
					errorMessage := fmt.Sprintf("Could not delete pdb for %s in %s", ds.Name, ds.Namespace)
					log.Println(errorMessage)
					ds.Status.ErrorMessage = errorMessage
					return
				}
			}
		}

		_, err = informer.CreatePodDisruptionBudgetFromDeploymentScaling(scaleTo, ds, deployment)
		if err != nil {
			errorMessage := fmt.Sprintf("Could not create pdb for %s in %s", ds.Name, ds.Namespace)
			log.Println(errorMessage)
			ds.Status.ErrorMessage = errorMessage
			return
		}
	}
}

func (informer *Informer) CreatePodDisruptionBudgetFromDeploymentScaling(scaleTo *deploymentscaling.ScaleTo, ds *deploymentscaling.DeploymentScaling, deployment *v1.Deployment) (*pv1.PodDisruptionBudget, error) {
	pdbSpec := pv1.PodDisruptionBudgetSpec{
		Selector: deployment.Spec.Selector,
	}

	if scaleTo.PodDisruptionBudget.MaxAvailable != nil {
		maxUnavailable := intstr.FromInt32(*scaleTo.PodDisruptionBudget.MaxAvailable)
		pdbSpec.MaxUnavailable = &maxUnavailable
	}

	if scaleTo.PodDisruptionBudget.MinAvailable != nil {
		minAvailable := intstr.FromInt32(*scaleTo.PodDisruptionBudget.MinAvailable)
		pdbSpec.MinAvailable = &minAvailable
	}

	pdb := pv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.Name,
			Namespace: deployment.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: ds.APIVersion,
					Kind:       ds.Kind,
					Name:       ds.Name,
					UID:        ds.UID,
				},
			},
			Labels: map[string]string{
				"scheduledscale.vandulmen.net/deployment": deployment.Name,
				"scheduledscale.vandulmen.net/owner":      ds.Name,
			},
			Annotations: make(map[string]string),
		},
		Spec: pdbSpec,
	}

	return informer.coreClientSet.PolicyV1().PodDisruptionBudgets(ds.Namespace).
		Create(context.TODO(), &pdb, metav1.CreateOptions{})
}
