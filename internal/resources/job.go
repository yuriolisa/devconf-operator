package resources

import (
	devconfczv1alpha1 "github.com/opdev/devconf-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// MySQLConfigMapForrecipe creates a ConfigMap for MySQL configuration
func JobForMySqlRestore(recipe *devconfczv1alpha1.Recipe, scheme *runtime.Scheme) (*batchv1.Job, error) {
	var job *batchv1.Job
	if recipe.Spec.Database.InitRestore {
		job = &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mysql-restore-job",
				Namespace: recipe.Namespace,
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image:           "fradelg/mysql-cron-backup",
							Name:            "mysql-restore-job",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{
									Name:  "CRON_TIME",
									Value: recipe.Spec.Database.BackupPolicy.Schedule,
								},
								{
									Name:  "INIT_RESTORE_LATEST",
									Value: "1",
								},
								{
									Name: "MYSQL_HOST",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: recipe.Name + "-mysql-config",
											},
											Key: "DB_HOST",
										},
									},
								}, {
									Name: "MYSQL_USER",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: recipe.Name + "-mysql-config",
											},
											Key: "MYSQL_USER",
										},
									},
								}, {
									Name: "MYSQL_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: recipe.Name + "-mysql",
											},
											Key: "MYSQL_PASSWORD",
										},
									},
								}, {
									Name: "MYSQL_ROOT_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: recipe.Name + "-mysql",
											},
											Key: "MYSQL_ROOT_PASSWORD",
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      recipe.Name + recipe.Spec.Database.BackupPolicy.VolumeName,
									MountPath: "/backup",
								},
							},
						}},
						Volumes: []corev1.Volume{
							{
								Name: recipe.Name + recipe.Spec.Database.BackupPolicy.VolumeName,
								VolumeSource: corev1.VolumeSource{
									PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
										ClaimName: recipe.Name + recipe.Spec.Database.BackupPolicy.VolumeName,
									},
								},
							},
						},
						RestartPolicy: "OnFailure",
					},
				},
			},
		}
	}
	if err := ctrl.SetControllerReference(recipe, job, scheme); err != nil {
		return nil, err
	}

	return job, nil
}
