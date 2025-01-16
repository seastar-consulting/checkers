package cloud

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seastar-consulting/checkers/checks"
)

// Save original functions for testing
var (
	originalNewSession = newSession
	originalNewSTS    = newSTS
	originalNewS3     = newS3
	originalTimeNow   = timeNow
)

func TestAwsAuthentication(t *testing.T) {
	// Save original functions and restore them after test
	defer func() {
		newSession = originalNewSession
		newSTS = originalNewSTS
	}()

	tests := []struct {
		name          string
		params        map[string]interface{}
		identity      string
		wantErr       bool
		wantStatus    string
		wantOutputStr string
	}{
		{
			name:          "successful authentication",
			params:        map[string]interface{}{"identity": "arn:aws:iam::123456789012:user/test"},
			identity:      "arn:aws:iam::123456789012:user/test",
			wantStatus:    "Success",
			wantOutputStr: "successfully authenticated with AWS",
		},
		{
			name:          "wrong identity",
			params:        map[string]interface{}{"identity": "arn:aws:iam::123456789012:user/test"},
			identity:      "arn:aws:iam::123456789012:user/wrong",
			wantStatus:    "Failure",
			wantOutputStr: "expected identity \"arn:aws:iam::123456789012:user/test\", but got \"arn:aws:iam::123456789012:user/wrong\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock AWS session
			newSession = func(profile string) (*session.Session, error) {
				return &session.Session{}, nil
			}

			// Mock STS client
			newSTS = func(sess *session.Session) stsiface.STSAPI {
				return &mockSTSClient{
					getCallerIdentityOutput: &sts.GetCallerIdentityOutput{
						Arn: aws.String(tt.identity),
					},
				}
			}

			// Get the check
			check, err := checks.Get("cloud.aws_authentication")
			require.NoError(t, err)
			require.Equal(t, "cloud.aws_authentication", check.Name)
			require.Equal(t, "Verifies AWS authentication and identity", check.Description)

			// Run the check
			result, err := check.Func(tt.params)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, result["status"])
			assert.Equal(t, tt.wantOutputStr, result["output"])
		})
	}
}

func TestAwsS3Access(t *testing.T) {
	// Save original functions and restore them after test
	defer func() {
		newSession = originalNewSession
		newS3 = originalNewS3
		timeNow = originalTimeNow
	}()

	// Mock time.Now
	mockTime := time.Date(2025, 1, 16, 17, 18, 59, 0, time.UTC)
	timeNow = func() time.Time {
		return mockTime
	}

	tests := []struct {
		name          string
		params        map[string]interface{}
		putErr        error
		getErr        error
		deleteErr     error
		wantErr       bool
		wantStatus    string
		wantOutputStr string
	}{
		{
			name: "successful write access (no key provided)",
			params: map[string]interface{}{
				"bucket": "test-bucket",
			},
			wantStatus:    "Success",
			wantOutputStr: "successfully verified write access to bucket test-bucket",
		},
		{
			name: "successful read access (key provided)",
			params: map[string]interface{}{
				"bucket": "test-bucket",
				"key":    "test-key",
			},
			wantStatus:    "Success",
			wantOutputStr: "successfully verified read access to object test-key in bucket test-bucket",
		},
		{
			name: "missing bucket",
			params: map[string]interface{}{
				"key": "test-key",
			},
			wantErr: true,
		},
		{
			name: "write access denied",
			params: map[string]interface{}{
				"bucket": "test-bucket",
			},
			putErr:        fmt.Errorf("access denied"),
			wantStatus:    "Failure",
			wantOutputStr: "failed to write to bucket test-bucket: access denied",
		},
		{
			name: "read access denied",
			params: map[string]interface{}{
				"bucket": "test-bucket",
				"key":    "test-key",
			},
			getErr:        fmt.Errorf("access denied"),
			wantStatus:    "Failure",
			wantOutputStr: "failed to read object test-key from bucket test-bucket: access denied",
		},
		{
			name: "delete access denied",
			params: map[string]interface{}{
				"bucket": "test-bucket",
			},
			deleteErr:     fmt.Errorf("access denied"),
			wantStatus:    "Failure",
			wantOutputStr: "failed to delete test object from bucket test-bucket: access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock AWS session
			newSession = func(profile string) (*session.Session, error) {
				return &session.Session{}, nil
			}

			// Mock S3 client
			newS3 = func(sess *session.Session) s3iface.S3API {
				return &mockS3Client{
					putErr:    tt.putErr,
					getErr:    tt.getErr,
					deleteErr: tt.deleteErr,
				}
			}

			// Get the check
			check, err := checks.Get("cloud.aws_s3_access")
			require.NoError(t, err)
			require.Equal(t, "cloud.aws_s3_access", check.Name)
			require.Equal(t, "Verifies read/write access to an S3 bucket", check.Description)

			// Run the check
			result, err := check.Func(tt.params)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, result["status"])
			assert.Equal(t, tt.wantOutputStr, result["output"])
		})
	}
}

type mockSTSClient struct {
	stsiface.STSAPI
	getCallerIdentityOutput *sts.GetCallerIdentityOutput
	err                    error
}

func (m *mockSTSClient) GetCallerIdentity(*sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.getCallerIdentityOutput, nil
}

type mockS3Client struct {
	s3iface.S3API
	putErr    error
	getErr    error
	deleteErr error
}

func (m *mockS3Client) PutObject(*s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	if m.putErr != nil {
		return nil, m.putErr
	}
	return &s3.PutObjectOutput{}, nil
}

func (m *mockS3Client) GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return &s3.GetObjectOutput{
		Body: io.NopCloser(strings.NewReader("test content")),
	}, nil
}

func (m *mockS3Client) DeleteObject(*s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	if m.deleteErr != nil {
		return nil, m.deleteErr
	}
	return &s3.DeleteObjectOutput{}, nil
}
