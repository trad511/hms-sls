module stash.us.cray.com/HMS/hms-sls

go 1.16

require (
	github.com/aws/aws-sdk-go v1.32.4
	github.com/golang-migrate/migrate/v4 v4.14.1
	github.com/gorilla/mux v1.7.4
	github.com/hashicorp/go-retryablehttp v0.6.0
	github.com/lib/pq v1.8.0
	github.com/namsral/flag v1.7.4-pre
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	go.uber.org/zap v1.15.0
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	stash.us.cray.com/HMS/hms-base v1.13.0
	stash.us.cray.com/HMS/hms-compcredentials v1.11.0
	stash.us.cray.com/HMS/hms-s3 v1.9.0
	stash.us.cray.com/HMS/hms-securestorage v1.12.0
)
