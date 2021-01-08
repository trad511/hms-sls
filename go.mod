module stash.us.cray.com/HMS/hms-sls

go 1.12

require (
	github.com/aws/aws-sdk-go v1.32.4
	github.com/golang-migrate/migrate/v4 v4.13.0
	github.com/gorilla/mux v1.7.4
	github.com/hashicorp/go-retryablehttp v0.6.0
	github.com/lib/pq v1.3.0
	github.com/namsral/flag v1.7.4-pre
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0 // indirect
	github.com/stretchr/testify v1.5.1
	go.uber.org/zap v1.15.0
	stash.us.cray.com/HMS/hms-base v1.8.4
	stash.us.cray.com/HMS/hms-compcredentials v1.7.0
	stash.us.cray.com/HMS/hms-s3 v1.4.1
	stash.us.cray.com/HMS/hms-securestorage v1.8.0
	stash.us.cray.com/HMS/hms-shcd-parser v1.1.1
)
