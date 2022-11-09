# script-go-aws-cw-cleanup

Golang script to delete empty streams in cloudwatch. Events get cleaned up by retention policy.

## .env configuration

``` bash
AWS_CW_CLEANUP_SCRIPT_ID="yyyyyyyyyyyyyyyy"
AWS_CW_CLEANUP_SCRIPT_KEY="xxxxxxxxxxxxxxxxxxxxxxxx"
AWS_CW_CLEANUP_SCRIPT_REGION="us-east-1"
```

## run

go run main.go