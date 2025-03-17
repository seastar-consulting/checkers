package cloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"

	"github.com/seastar-consulting/checkers/checks"
	"github.com/seastar-consulting/checkers/types"
)

// for testing
var (
	newSession = defaultNewSession
	newSTS     = defaultNewSTS
	newS3      = defaultNewS3
	timeNow    = time.Now
)

func init() {
	checks.Register(
		"cloud.aws_authentication",
		"Verifies AWS authentication and identity",
		types.CheckSchema{
			Parameters: map[string]types.ParameterSchema{
				"aws_profile": {
					Type:        types.StringType,
					Description: "AWS profile to use for authentication. If not specified, default credentials will be used.",
					Required:    false,
				},
				"identity": {
					Type:        types.StringType,
					Description: "Expected AWS IAM identity ARN to verify against.",
					Required:    true,
				},
			},
		},
		CheckAwsAuthentication,
	)

	checks.Register(
		"cloud.aws_s3_access",
		"Verifies read/write access to an S3 bucket",
		types.CheckSchema{
			Parameters: map[string]types.ParameterSchema{
				"aws_profile": {
					Type:        types.StringType,
					Description: "AWS profile to use for authentication. If not specified, default credentials will be used.",
					Required:    false,
				},
				"bucket": {
					Type:        types.StringType,
					Description: "Name of the S3 bucket to check access for.",
					Required:    true,
				},
				"key": {
					Type:        types.StringType,
					Description: "Optional key to verify read access to. If not specified, a test object will be created and deleted.",
					Required:    false,
				},
			},
		},
		CheckAwsS3Access,
	)
}

func defaultNewSession(profile string) (*session.Session, error) {
	if profile != "" {
		return session.NewSessionWithOptions(session.Options{
			Config: aws.Config{
				Region: aws.String("us-east-1"),
			},
			Profile: profile,
		})
	}
	return session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
}

func defaultNewSTS(sess *session.Session) stsiface.STSAPI {
	return sts.New(sess)
}

func defaultNewS3(sess *session.Session) s3iface.S3API {
	return s3.New(sess)
}

// CheckAwsAuthentication verifies the user can authenticate successfully with AWS and has the correct identity as returned by STS.
func CheckAwsAuthentication(item types.CheckItem) (types.CheckResult, error) {
	// Get optional AWS profile
	awsProfile := item.Parameters["aws_profile"]

	// Get required identity
	identity := item.Parameters["identity"]
	if identity == "" {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Error,
			Error:  "identity parameter is required",
		}, nil
	}

	sess, err := newSession(awsProfile)
	if err != nil {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Error,
			Error:  fmt.Sprintf("error creating AWS session: %v", err),
		}, nil
	}

	svc := newSTS(sess)
	input := &sts.GetCallerIdentityInput{}

	stsResult, err := svc.GetCallerIdentity(input)
	if err != nil {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Error,
			Error:  fmt.Sprintf("error calling GetCallerIdentity: %v", err),
		}, nil
	}

	if stsResult.Arn == nil || *stsResult.Arn != identity {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Failure,
			Output: fmt.Sprintf("Expected identity '%s', but got '%s'", identity, *stsResult.Arn),
		}, nil
	}

	return types.CheckResult{
		Name:   item.Name,
		Type:   item.Type,
		Status: types.Success,
		Output: fmt.Sprintf("Successfully authenticated with AWS as '%s'", *stsResult.Arn),
	}, nil
}

// CheckAwsS3Access verifies read/write access to an S3 bucket by attempting to put and get an object.
// If a key is provided, it verifies read access to that key. If not, it creates a new object with
// a random name, writes to it, and then deletes it.
func CheckAwsS3Access(item types.CheckItem) (types.CheckResult, error) {
	// Get required parameters
	bucket := item.Parameters["bucket"]
	if bucket == "" {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Error,
			Error:  "bucket parameter is required",
		}, nil
	}

	// Get optional parameters
	awsProfile := item.Parameters["aws_profile"]

	// Create AWS session
	sess, err := newSession(awsProfile)
	if err != nil {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Error,
			Error:  fmt.Sprintf("error creating AWS session: %v", err),
		}, nil
	}

	// Create S3 client
	svc := newS3(sess)

	// Check if key is provided
	key := item.Parameters["key"]
	if key != "" {
		// Verify read access to the specified key
		_, err = svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return types.CheckResult{
				Name:   item.Name,
				Type:   item.Type,
				Status: types.Failure,
				Output: fmt.Sprintf("Failed to read object '%s' from bucket '%s': %v", key, bucket, err),
			}, nil
		}

		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Success,
			Output: fmt.Sprintf("Successfully verified read access to object '%s' in bucket '%s'", key, bucket),
		}, nil
	}

	// Generate a random key for testing write access
	timestamp := timeNow().UTC().Format("20060102-150405.000")
	testKey := fmt.Sprintf("access-check/%s.txt", timestamp)

	// Test write access by putting a small object
	content := "test content"
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(testKey),
		Body:   strings.NewReader(content),
	})
	if err != nil {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Failure,
			Output: fmt.Sprintf("Failed to write to bucket '%s': %v", bucket, err),
		}, nil
	}

	// Clean up by deleting the test object
	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(testKey),
	})
	if err != nil {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Failure,
			Output: fmt.Sprintf("Failed to delete test object from bucket '%s': %v", bucket, err),
		}, nil
	}

	return types.CheckResult{
		Name:   item.Name,
		Type:   item.Type,
		Status: types.Success,
		Output: fmt.Sprintf("Successfully verified read/write access to bucket '%s'", bucket),
	}, nil
}
