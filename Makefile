build:
	go build -v ./...
test:
	go test -race -covermode atomic -coverprofile=profile.cov ./... 