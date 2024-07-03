package plugin

import (
	"testing"

	"github.com/luthermonson/go-proxmox"
	"github.com/stretchr/testify/require"
)

func TestInstanceGroup_templateCloneOptions(t *testing.T) {
	type testCase struct {
		name              string
		isTemplate        bool
		configuredStorage string
		expectedFull      uint8
		expectedErr       error
	}

	testCases := []testCase{
		{
			name:              "VM with unconfigured storage", // Error?
			isTemplate:        false,
			configuredStorage: "",
			expectedFull:      1,
			expectedErr:       ErrCloneVMWithoutConfiguredStorage,
		},
		{
			name:              "VM with configured storage",
			isTemplate:        false,
			configuredStorage: "local",
			expectedFull:      1,
			expectedErr:       nil,
		},
		{
			name:              "Template with unconfigured storage",
			isTemplate:        true,
			configuredStorage: "",
			expectedFull:      0,
			expectedErr:       nil,
		},
		{
			name:              "Template with configured storage",
			isTemplate:        true,
			configuredStorage: "local",
			expectedFull:      1,
			expectedErr:       nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			template := proxmox.VirtualMachine{
				Template: proxmox.IsTemplate(testCase.isTemplate),
			}

			ig := InstanceGroup{
				Settings: Settings{
					Storage: testCase.configuredStorage,
				},
			}

			result, err := ig.getTemplateCloneOptions(&template)
			require.ErrorIs(t, err, testCase.expectedErr)
			if err == nil {
				require.Equal(t, testCase.configuredStorage, result.Storage)
				require.Equal(t, testCase.expectedFull, result.Full)
			}
		})
	}
}
