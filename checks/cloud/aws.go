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
)

// for testing
var (
	newSession = defaultNewSession
	newSTS     = defaultNewSTS
	newS3      = defaultNewS3
	timeNow    = time.Now
)

func init() {
	checks.Register("cloud.aws_authentication", "Verifies AWS authentication and identity", AwsAuthentication)
	checks.Register("cloud.aws_s3_access", "Verifies read/write access to an S3 bucket", AwsS3Access)
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

// AwsAuthentication verifies the user can authenticate successfully with AWS and has the correct identity as returned by STS.
func AwsAuthentication(params map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	var awsProfile, identity string
	if profile, ok := params["aws_profile"].(string); ok {
		awsProfile = profile
	}
	if id, ok := params["identity"].(string); ok {
		identity = id
	}

	sess, err := newSession(awsProfile)
	if err != nil {
		return nil, fmt.Errorf("error creating AWS session: %v", err)
	}

	svc := newSTS(sess)
	input := &sts.GetCallerIdentityInput{}

	stsResult, err := svc.GetCallerIdentity(input)
	if err != nil {
		return nil, fmt.Errorf("error calling GetCallerIdentity: %v", err)
	}

	if stsResult.Arn == nil || *stsResult.Arn != identity {
		result["status"] = "Failure"
		result["output"] = fmt.Sprintf("expected identity %q, but got %q", identity, *stsResult.Arn)
		return result, nil
	}

	result["status"] = "Success"
	result["output"] = "successfully authenticated with AWS"
	return result, nil
}

// AwsS3Access verifies read/write access to an S3 bucket by attempting to put and get an object.
// If a key is provided, it verifies read access to that key. If not, it creates a new object with
// a random name, writes to it, and then deletes it.
func AwsS3Access(params map[string]interface{}) (map[string]interface{}, error) {
	// Get required parameters
	bucket, ok := params["bucket"].(string)
	if !ok || bucket == "" {
		return nil, fmt.Errorf("bucket parameter is required")
	}

	// Get optional parameters
	var awsProfile string
	if profile, ok := params["aws_profile"].(string); ok {
		awsProfile = profile
	}

	// Create AWS session
	sess, err := newSession(awsProfile)
	if err != nil {
		return nil, fmt.Errorf("error creating AWS session: %v", err)
	}

	// Create S3 client
	svc := newS3(sess)

	// Check if key is provided
	key, hasKey := params["key"].(string)
	if hasKey && key != "" {
		// Verify read access to the specified key
		_, err = svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return map[string]interface{}{
				"status": "Failure",
				"output": fmt.Sprintf("failed to read object %s from bucket %s: %v", key, bucket, err),
			}, nil
		}

		return map[string]interface{}{
			"status": "Success",
			"output": fmt.Sprintf("successfully verified read access to object %s in bucket %s", key, bucket),
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
		return map[string]interface{}{
			"status": "Failure",
			"output": fmt.Sprintf("failed to write to bucket %s: %v", bucket, err),
		}, nil
	}

	// Clean up by deleting the test object
	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(testKey),
	})
	if err != nil {
		return map[string]interface{}{
			"status": "Failure",
			"output": fmt.Sprintf("failed to delete test object from bucket %s: %v", bucket, err),
		}, nil
	}

	return map[string]interface{}{
		"status": "Success",
		"output": fmt.Sprintf("successfully verified write access to bucket %s", bucket),
	}, nil
}
