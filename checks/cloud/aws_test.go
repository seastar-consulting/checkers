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

	"github.com/seastar-consulting/checkers/types"
)

// Save original functions for testing
var (
	originalNewSession = newSession
	originalNewSTS    = newSTS
	originalNewS3     = newS3
	originalTimeNow   = timeNow
)

func TestCheckAwsAuthentication(t *testing.T) {
	// Save original functions and restore them after test
	defer func() {
		newSession = originalNewSession
		newSTS = originalNewSTS
	}()

	tests := []struct {
		name      string
		checkItem types.CheckItem
		identity  string
		want      types.CheckResult
		wantErr   bool
	}{
		{
			name: "successful authentication",
			checkItem: types.CheckItem{
				Name: "test-check",
				Type: "cloud.aws_authentication",
				Parameters: map[string]string{
					"identity": "arn:aws:iam::123456789012:user/test",
				},
			},
			identity: "arn:aws:iam::123456789012:user/test",
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "cloud.aws_authentication",
				Status: types.Success,
				Output: "Successfully authenticated with AWS as 'arn:aws:iam::123456789012:user/test'",
			},
		},
		{
			name: "wrong identity",
			checkItem: types.CheckItem{
				Name: "test-check",
				Type: "cloud.aws_authentication",
				Parameters: map[string]string{
					"identity": "arn:aws:iam::123456789012:user/test",
				},
			},
			identity: "arn:aws:iam::123456789012:user/wrong",
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "cloud.aws_authentication",
				Status: types.Failure,
				Output: "Expected identity 'arn:aws:iam::123456789012:user/test', but got 'arn:aws:iam::123456789012:user/wrong'",
			},
		},
		{
			name: "missing identity",
			checkItem: types.CheckItem{
				Name:       "test-check",
				Type:       "cloud.aws_authentication",
				Parameters: map[string]string{},
			},
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "cloud.aws_authentication",
				Status: types.Error,
				Error:  "identity parameter is required",
			},
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

			got, err := CheckAwsAuthentication(tt.checkItem)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckAwsAuthentication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCheckAwsS3Access(t *testing.T) {
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
		name      string
		checkItem types.CheckItem
		putErr    error
		getErr    error
		deleteErr error
		want      types.CheckResult
		wantErr   bool
	}{
		{
			name: "successful write access (no key provided)",
			checkItem: types.CheckItem{
				Name: "test-check",
				Type: "cloud.aws_s3_access",
				Parameters: map[string]string{
					"bucket": "test-bucket",
				},
			},
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "cloud.aws_s3_access",
				Status: types.Success,
				Output: "Successfully verified write access to bucket 'test-bucket'",
			},
		},
		{
			name: "successful read access (key provided)",
			checkItem: types.CheckItem{
				Name: "test-check",
				Type: "cloud.aws_s3_access",
				Parameters: map[string]string{
					"bucket": "test-bucket",
					"key":    "test-key",
				},
			},
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "cloud.aws_s3_access",
				Status: types.Success,
				Output: "Successfully verified read access to object 'test-key' in bucket 'test-bucket'",
			},
		},
		{
			name: "missing bucket",
			checkItem: types.CheckItem{
				Name: "test-check",
				Type: "cloud.aws_s3_access",
				Parameters: map[string]string{
					"key": "test-key",
				},
			},
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "cloud.aws_s3_access",
				Status: types.Error,
				Error:  "bucket parameter is required",
			},
		},
		{
			name: "write access denied",
			checkItem: types.CheckItem{
				Name: "test-check",
				Type: "cloud.aws_s3_access",
				Parameters: map[string]string{
					"bucket": "test-bucket",
				},
			},
			putErr: fmt.Errorf("access denied"),
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "cloud.aws_s3_access",
				Status: types.Failure,
				Output: "Failed to write to bucket 'test-bucket': access denied",
			},
		},
		{
			name: "read access denied",
			checkItem: types.CheckItem{
				Name: "test-check",
				Type: "cloud.aws_s3_access",
				Parameters: map[string]string{
					"bucket": "test-bucket",
					"key":    "test-key",
				},
			},
			getErr: fmt.Errorf("access denied"),
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "cloud.aws_s3_access",
				Status: types.Failure,
				Output: "Failed to read object 'test-key' from bucket 'test-bucket': access denied",
			},
		},
		{
			name: "delete access denied",
			checkItem: types.CheckItem{
				Name: "test-check",
				Type: "cloud.aws_s3_access",
				Parameters: map[string]string{
					"bucket": "test-bucket",
				},
			},
			deleteErr: fmt.Errorf("access denied"),
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "cloud.aws_s3_access",
				Status: types.Failure,
				Output: "Failed to delete test object from bucket 'test-bucket': access denied",
			},
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

			got, err := CheckAwsS3Access(tt.checkItem)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckAwsS3Access() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
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
