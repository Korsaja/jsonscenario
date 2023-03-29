generate:
	go get github.com/dmarkham/enumer
	go run github.com/dmarkham/enumer -type=ActionFile -json -transform=snake -output=pkg/models/actions_string.go pkg/models/actions.go
	go run github.com/dmarkham/enumer -type=Result -json -transform=snake -output=internal/ffuncs/const_string.go internal/ffuncs/const.go
	go mod tidy

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.1 run ./... --max-same-issues 0