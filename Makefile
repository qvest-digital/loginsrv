build:
	GOOS=linux go build -a --ldflags '-extldflags "-static"' .
