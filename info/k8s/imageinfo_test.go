package k8s

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetContainerImage(t *testing.T) {
	// Following are the attributes of the target DaemonSet and container in the
	// tests.
	testDSName := "some-daemonset"
	testDSNamespace := "kube-system"
	testContainerName := "containerA"
	testImage := "image/A:tagA"

	testcases := []struct {
		name        string
		noDaemonSet bool
		containers  []corev1.Container
		defaultDS   bool
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
		{
			// This should fail because default DaemonSet attributes are used
			// and the DaemonSet with those attributes don't exist.
			name: "default daemonset attributes",
			containers: []corev1.Container{
				{
					Image: testImage,
					Name:  testContainerName,
				},
			},
			defaultDS: true,
			wantErr:   true,
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

			info := NewImageInfo(client)

			// Do not use the default DaemonSet attributes.
			if !tc.defaultDS {
				info.SetDaemonSet(testDSName, testDSNamespace)
			}

			img, err := info.GetContainerImage(testContainerName)
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

// getNoContainerDS returns a DaemonSet without any pod containers.
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
