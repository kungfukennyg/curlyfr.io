SERVER_BINARY_NAME=bin/web

server:
	go build -o ${SERVER_BINARY_NAME} .

server-amd64: 
	GOOS=linux GOARCH=amd64 go build -o ${SERVER_BINARY_NAME}-amd64-linux .

deploy-server-amd64: server-amd64
	scp ${SERVER_BINARY_NAME}-amd64-linux curlyfr.io:web/new-server
	hugo && rsync -avz --delete public curlyfr.io:web/
	ssh curlyfr.io './web/restart.sh'

clean:
	go clean
	rm ${bin/web}