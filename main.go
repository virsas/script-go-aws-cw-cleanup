package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print(".env file does not exist, will get the variables from the environment")
	}
}

func main() {
	session, err := createSession()
	if err != nil {
		panic(err)
	}
	cwl := cloudwatchlogs.New(session)

	logGroups, err := getLogGroups(cwl)
	if err != nil {
		panic(err)
	}

	for _, lg := range logGroups {
		if lg.RetentionInDays != nil {
			now := time.Now().Unix()
			oldestPossibleEvent := now - (*lg.RetentionInDays * 24 * 60 * 60)

			for {
				streams, nextToken, err := getStreams(cwl, lg.LogGroupName, 50)
				if err != nil {
					panic(err)
				}

				for _, stream := range streams {
					var lastStreamTime int64
					if stream.LastEventTimestamp != nil {
						lastStreamTime = *stream.LastEventTimestamp / 1000
					} else {
						lastStreamTime = *stream.CreationTime / 1000
					}
					if lastStreamTime < oldestPossibleEvent {
						err = deleteStream(cwl, lg.LogGroupName, stream.LogStreamName)
						if err != nil {
							panic(err)
						}
					}
				}

				if nextToken == nil {
					break
				}
			}
		}
	}

}

func createSession() (*session.Session, error) {
	var awsID string = ""
	awsIDValue, awsIDPresent := os.LookupEnv("AWS_CWL_CLEANUP_SCRIPT_ID")
	if awsIDPresent {
		awsID = awsIDValue
	} else {
		return nil, errors.New("missing ENV Variable - AWS_CWL_CLEANUP_SCRIPT_ID")
	}

	var awsKey string = ""
	awsKeyValue, awsKeyPresent := os.LookupEnv("AWS_CWL_CLEANUP_SCRIPT_KEY")
	if awsKeyPresent {
		awsKey = awsKeyValue
	} else {
		return nil, errors.New("missing ENV Variable - AWS_CWL_CLEANUP_SCRIPT_KEY")
	}

	var region string = "eu-west-1"
	regionValue, regionPresent := os.LookupEnv("AWS_CWL_CLEANUP_SCRIPT_REGION")
	if regionPresent {
		region = regionValue
	} else {
		return nil, errors.New("missing ENV Variable - AWS_CWL_CLEANUP_SCRIPT_REGION")
	}

	session, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(awsID, awsKey, ""),
	})
	if err != nil {
		return nil, err
	}
	return session, nil
}

func getLogGroups(cwl *cloudwatchlogs.CloudWatchLogs) ([]*cloudwatchlogs.LogGroup, error) {
	opt := &cloudwatchlogs.DescribeLogGroupsInput{}
	logGroups, err := cwl.DescribeLogGroups(opt)
	if err != nil {
		return nil, err
	}

	return logGroups.LogGroups, nil
}

func getStreams(cwl *cloudwatchlogs.CloudWatchLogs, logGroup *string, limit int64) ([]*cloudwatchlogs.LogStream, *string, error) {

	var enabled bool = true
	opt := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: logGroup,
		Limit:        &limit,
		Descending:   &enabled,
	}

	streams, err := cwl.DescribeLogStreams(opt)
	if err != nil {
		return nil, nil, err
	}

	return streams.LogStreams, streams.NextToken, nil
}

func deleteStream(cwl *cloudwatchlogs.CloudWatchLogs, logGroup *string, stream *string) error {
	fmt.Println("#### Deleting stream", *stream, "in loggroup", *logGroup, "####")
	opt := &cloudwatchlogs.DeleteLogStreamInput{
		LogGroupName:  logGroup,
		LogStreamName: stream,
	}

	_, err := cwl.DeleteLogStream(opt)
	if err != nil {
		return err
	}

	return nil
}
