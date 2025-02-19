.PHONY : server-build server-test server-docker worker-build worker-test worker-docker test check

server-build:
	go build -o /dev/null -v ./src/server/server.go

server-run:
	go run -v ./src/server/server.go

server-test:
	go test -v ./src/server/...

server-docker:
	docker build . --file ./docker/server/Dockerfile

worker-build:
	go build -o /dev/null -v ./src/worker/worker.go

worker-run:
	go run -v ./src/worker/worker.go

worker-test:
	go test -v ./src/worker/...

worker-docker:
	docker build . --file ./docker/worker/Dockerfile

track-test:
	go test -v ./src/shared/integration_test/...

test:
	go test -v ./src/...

check: server-build worker-build test server-docker worker-docker

backup-db:
	python3 ./local_db/scripts/dynamodump.py -m backup -s "*" --host localhost --port 8000 --accessKey local --secretKey local --region local --dumpPath ./local_db/db_dump

backup-production:
	python3 ./local_db/scripts/dynamodump.py -m backup -r us-east-2 -s "*" --accessKey ${PRODUCTION_AWS_ACCESS_KEY_ID} --secretKey ${PRODUCTION_AWS_SECRET_ACCESS_KEY} --dumpPath ./local_db/db_dump

restore-db:
	python3 ./local_db/scripts/dynamodump.py -m restore -s "*" --host localhost --port 8000 --accessKey local --secretKey local --region local --dumpPath ./local_db/db_dump

dynamo:
	docker run -p 8000:8000 amazon/dynamodb-local

rabbit:
	docker run -p 5672:5672 rabbitmq:3.10