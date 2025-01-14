package cloud

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seastar-consulting/checkers/checks"
)

type mockSTSClient struct {
	stsiface.STSAPI
	getCallerIdentityOutput *sts.GetCallerIdentityOutput
	err                    error
}

func (m *mockSTSClient) GetCallerIdentity(*sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error) {
	return m.getCallerIdentityOutput, m.err
}

func TestAwsAuthentication(t *testing.T) {
	// Save the original functions
	origNewSession := newSession
	origNewSTS := newSTS
	defer func() {
		// Restore the original functions after the test
		newSession = origNewSession
		newSTS = origNewSTS
	}()

	// Mock the session creation
	newSession = func(profile string) (*session.Session, error) {
		return session.Must(session.NewSession()), nil
	}

	check, err := checks.Get("cloud.aws_authentication")
	require.NoError(t, err)
	require.Equal(t, "cloud.aws_authentication", check.Name)
	require.Equal(t, "Verifies AWS authentication and identity", check.Description)

	tests := []struct {
		name           string
		params         map[string]interface{}
		mockOutput     *sts.GetCallerIdentityOutput
		mockError      error
		expectedResult map[string]interface{}
		expectedError  string
	}{
		{
			name: "successful authentication with matching identity",
			params: map[string]interface{}{
				"aws_profile": "test-profile",
				"identity":    "arn:aws:iam::123456789012:user/test-user",
			},
			mockOutput: &sts.GetCallerIdentityOutput{
				Arn: aws.String("arn:aws:iam::123456789012:user/test-user"),
			},
			expectedResult: map[string]interface{}{
				"status": "Success",
				"output": "successfully authenticated with AWS",
			},
		},
		{
			name: "authentication successful but identity mismatch",
			params: map[string]interface{}{
				"aws_profile": "test-profile",
				"identity":    "arn:aws:iam::123456789012:user/expected-user",
			},
			mockOutput: &sts.GetCallerIdentityOutput{
				Arn: aws.String("arn:aws:iam::123456789012:user/actual-user"),
			},
			expectedResult: map[string]interface{}{
				"status": "Failure",
				"output": `expected identity "arn:aws:iam::123456789012:user/expected-user", but got "arn:aws:iam::123456789012:user/actual-user"`,
			},
		},
		{
			name: "authentication error",
			params: map[string]interface{}{
				"aws_profile": "test-profile",
				"identity":    "arn:aws:iam::123456789012:user/test-user",
			},
			mockError:     fmt.Errorf("authentication failed"),
			expectedError: "error calling GetCallerIdentity: authentication failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock STS client for this test
			mockSTS := &mockSTSClient{
				getCallerIdentityOutput: tt.mockOutput,
				err:                     tt.mockError,
			}
			newSTS = func(sess *session.Session) stsiface.STSAPI {
				return mockSTS
			}

			result, err := check.Func(tt.params)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
