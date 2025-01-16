package cloud

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"

	"github.com/seastar-consulting/checkers/checks"
)

// for testing
var (
	newSession = defaultNewSession
	newSTS     = defaultNewSTS
)

func init() {
	checks.Register("cloud.aws_authentication", "Verifies AWS authentication and identity", AwsAuthentication)
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
