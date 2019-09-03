package main

import (
	"errors"
	"os"
	"testing"

	"github.com/storageos/init/info/k8s"
	"github.com/storageos/init/mocks"

	"github.com/golang/mock/gomock"
)

func TestGetParamsForK8SImageInfo(t *testing.T) {
	testcases := []struct {
		name            string
		dsNameVar       string
		dsNamespaceVar  string
		envvars         map[string]string
		wantDSName      string
		wantDSNamespace string
	}{
		{
			name:            "flag variables",
			dsNameVar:       "ds1",
			dsNamespaceVar:  "ns1",
			wantDSName:      "ds1",
			wantDSNamespace: "ns1",
		},
		{
			name: "env variables",
			envvars: map[string]string{
				daemonSetNameEnvVar:      "ds2",
				daemonSetNamespaceEnvVar: "ns2",
			},
			wantDSName:      "ds2",
			wantDSNamespace: "ns2",
		},
		{
			name:            "default values",
			wantDSName:      k8s.DefaultDaemonSetName,
			wantDSNamespace: k8s.DefaultDaemonSetNamespace,
		},
		{
			name:           "flag and env var",
			dsNameVar:      "ds3",
			dsNamespaceVar: "ns3",
			envvars: map[string]string{
				daemonSetNameEnvVar:      "ds4",
				daemonSetNamespaceEnvVar: "ns4",
			},
			wantDSName:      "ds3",
			wantDSNamespace: "ns3",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// Set env vars if provided.
			if len(tc.envvars) > 0 {
				for k, v := range tc.envvars {
					os.Setenv(k, v)

					// Unset env vars at the end.
					defer os.Unsetenv(k)
				}
			}

			name, namespace := getParamsForK8SImageInfo(tc.dsNameVar, tc.dsNamespaceVar)
			if name != tc.wantDSName {
				t.Errorf("unexpected daemonset name:\n\t(WNT) %s\n\t(GOT) %s", tc.wantDSName, name)
			}
			if namespace != tc.wantDSNamespace {
				t.Errorf("unexpected daemonset namespace:\n\t(WNT) %s\n\t(GOT) %s", tc.wantDSNamespace, namespace)
			}
		})
	}
}

func TestRunScript(t *testing.T) {
	testcases := []struct {
		name    string
		scripts []string
		envvars map[string]string
		retErr  error
		wantErr bool
	}{
		{
			name:    "simple run",
			scripts: []string{"script1", "script2", "script3"},
			envvars: map[string]string{
				"FOO": "val1",
				"BAR": "val2",
			},
		},
		{
			name:    "error run",
			scripts: []string{"sc1"},
			retErr:  errors.New("some-error"),
			wantErr: true,
		},
		{
			name: "no script",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockRunner := mocks.NewMockRunner(mockCtrl)

			// Returned error is tc.retErr. Avoid adding multiple scripts when
			// tc.retErr is set. All the calls will return an error. This will
			// result in unexpected number of times the function is called
			// because the first error will end the parent runScript function.
			mockRunner.EXPECT().
				RunScript(gomock.Any(), tc.envvars).
				Return(nil, nil, tc.retErr).
				Times(len(tc.scripts))

			if err := runScripts(mockRunner, tc.scripts, tc.envvars); err != nil {
				if !tc.wantErr {
					t.Errorf("unexpected error while running scripts: %v", err)
				}
			}
		})
	}
}
