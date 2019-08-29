package inspector

import (
	"testing"

	"k8s.io/client-go/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetDaemonSetContainerImage(t *testing.T) {
	testDSName := "some-daemonset"
	testDSNamespace := "kube-system"
	testContainerName := "containerA"
	testImage := "image/A:tagA"

	testcases := []struct {
		name        string
		noDaemonSet bool
		containers  []corev1.Container
		wantErr     bool
	}{
		{
			name:        "no daemonset",
			noDaemonSet: true,
			wantErr:     true,
		},
		{
			name: "no target container",
			containers: []corev1.Container{
				{
					Image: "image/B:tagB",
					Name:  "containerB",
				},
				{
					Image: "image/C:tagC",
					Name:  "containerC",
				},
			},
			wantErr: true,
		},
		{
			name: "target container",
			containers: []corev1.Container{
				{
					Image: "image/B:tagB",
					Name:  "containerB",
				},
				{
					Image: testImage,
					Name:  testContainerName,
				},
				{
					Image: "image/C:tagC",
					Name:  "containerC",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var client kubernetes.Interface

			if !tc.noDaemonSet {
				ds := getNoContainerDS(testDSName, testDSNamespace)
				ds.Spec.Template.Spec.Containers = tc.containers
				// Create DaemonSet.
				client = fake.NewSimpleClientset(ds)
			} else {
				client = fake.NewSimpleClientset()
			}

			inspect := NewInspect(client)
			img, err := inspect.GetDaemonSetContainerImage(testDSName, testDSNamespace, testContainerName)
			if err != nil {
				if !tc.wantErr {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				if img != testImage {
					t.Errorf("unexpected image:\n\t(WNT) %s\n\t(GOT) %s", testImage, img)
				}
			}
		})
	}
}

func getNoContainerDS(name, namespace string) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "DaemonSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{},
			},
		},
	}

}
