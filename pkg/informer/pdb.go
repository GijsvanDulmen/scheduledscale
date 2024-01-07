package informer

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	v1 "k8s.io/api/apps/v1"
	pv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	log2 "log"
	"scheduledscale/pkg/apis/scheduledscalecontroller"
	"scheduledscale/pkg/apis/scheduledscalecontroller/v1alpha1/deploymentscaling"
)

const ownerLabel = scheduledscalecontroller.GroupName + "/owner"
const deploymentLabel = scheduledscalecontroller.GroupName + "/deployment"

func (informer *Informer) DeletePodDisruptionBudgetsFor(ds *deploymentscaling.DeploymentScaling, deletePdb bool) error {
	pdbListOptions := metav1.ListOptions{
		LabelSelector: labels.Set(map[string]string{
			ownerLabel: ds.Name,
		}).String(),
	}

	log.Debug().Msg(pdbListOptions.LabelSelector)

	pdbList, err := informer.coreClientSet.
		PolicyV1().
		PodDisruptionBudgets(ds.Namespace).
		List(context.TODO(), pdbListOptions)

	if err != nil {
		return err
	}

	log.Debug().Msgf("Received %d PDBs", len(pdbList.Items))

	for _, pdb := range pdbList.Items {
		if deletePdb {
			log.Info().Msgf("Deleting PDB %s for %s in %s\n", pdb.Name, ds.Name, ds.Namespace)
			err = informer.coreClientSet.PolicyV1().PodDisruptionBudgets(ds.Namespace).
				Delete(context.TODO(), pdb.Name, metav1.DeleteOptions{})
			if err != nil {
				log.Error().Msgf("Deleting PDB %s for %s in %s gone wrong\n", pdb.Name, ds.Name, ds.Namespace)
				log.Error().Err(err)
			}
		} else {
			log.Info().Msgf("Removing owner and labels of PDB %s for %s in %s", pdb.Name, ds.Name, ds.Namespace)
			pdb.OwnerReferences = []metav1.OwnerReference{}
			delete(pdb.Labels, deploymentLabel)
			delete(pdb.Labels, ownerLabel)

			_, err = informer.coreClientSet.PolicyV1().PodDisruptionBudgets(ds.Namespace).
				Update(context.TODO(), &pdb, metav1.UpdateOptions{})

			if err != nil {
				log.Error().Msgf("Removing owner and labels of PDB %s for %s in %s gone wrong\n", pdb.Name, ds.Name, ds.Namespace)
				log.Error().Err(err)
			}
		}
	}
	return nil
}

func (informer *Informer) ReconcilePodDisruptionBudget(scaleTo *deploymentscaling.ScaleTo, ds *deploymentscaling.DeploymentScaling, deployment *v1.Deployment) {
	if scaleTo.PodDisruptionBudget != nil {
		pdb := *scaleTo.PodDisruptionBudget
		if pdb.Enabled {
			pdbListOptions := metav1.ListOptions{
				LabelSelector: labels.Set(map[string]string{
					deploymentLabel: deployment.Name,
				}).String(),
			}

			pdbList, err := informer.coreClientSet.
				PolicyV1().
				PodDisruptionBudgets(ds.Namespace).
				List(context.TODO(), pdbListOptions)

			if err != nil {
				log.Error().Msgf("Could not list pdbs for %s in %s - not doing anything", ds.Name, ds.Namespace)
				log.Error().Err(err)
				return
			} else {
				if len(pdbList.Items) > 1 {
					errorMessage := fmt.Sprintf("Multiple pdbs found for %s in %s - not doing anything", ds.Name, ds.Namespace)
					log.Error().Msg(errorMessage)
					return
				} else if len(pdbList.Items) > 0 {
					err = informer.coreClientSet.PolicyV1().PodDisruptionBudgets(ds.Namespace).
						Delete(context.TODO(), pdbList.Items[0].Name, metav1.DeleteOptions{})
					if err != nil {
						log.Error().Msgf("Could not delete pdb for %s in %s", ds.Name, ds.Namespace)
						return
					}
				}
			}

			LogForDeploymentScaling(*ds, fmt.Sprintf("Creating pdb for %s in %s", ds.Name, ds.Namespace), zerolog.InfoLevel)
			_, err = informer.CreatePodDisruptionBudgetFromDeploymentScaling(scaleTo, ds, deployment)
			if err != nil {
				log.Error().Msgf("Could not create pdb for %s in %s", ds.Name, ds.Namespace)
				log2.Print(err)
				log.Error().Err(err)
				return
			}
		} else {
			LogForDeploymentScaling(*ds, fmt.Sprintf("Deleting pdb for %s in %s", ds.Name, ds.Namespace), zerolog.InfoLevel)
			err := informer.DeletePodDisruptionBudgetsFor(ds, true)
			if err != nil {
				log.Error().Msgf("Could not delete pdb for %s in %s", ds.Name, ds.Namespace)
				log.Error().Err(err)
				return
			}
		}
	}
}

func (informer *Informer) CreatePodDisruptionBudgetFromDeploymentScaling(scaleTo *deploymentscaling.ScaleTo, ds *deploymentscaling.DeploymentScaling, deployment *v1.Deployment) (*pv1.PodDisruptionBudget, error) {
	pdb := CreatePodDisruptionBudget(scaleTo, ds, deployment)

	return informer.coreClientSet.PolicyV1().PodDisruptionBudgets(ds.Namespace).
		Create(context.TODO(), &pdb, metav1.CreateOptions{})
}

func CreatePodDisruptionBudget(scaleTo *deploymentscaling.ScaleTo, ds *deploymentscaling.DeploymentScaling, deployment *v1.Deployment) pv1.PodDisruptionBudget {
	pdbSpec := pv1.PodDisruptionBudgetSpec{
		Selector: deployment.Spec.Selector,
	}

	if scaleTo.PodDisruptionBudget.MaxUnavailable != nil {
		maxUnavailable := intstr.FromInt32(*scaleTo.PodDisruptionBudget.MaxUnavailable)
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
				deploymentLabel: deployment.Name,
				ownerLabel:      ds.Name,
			},
			Annotations: make(map[string]string),
		},
		Spec: pdbSpec,
	}
	return pdb
}
