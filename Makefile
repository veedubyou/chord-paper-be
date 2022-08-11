run-server:
	go run -v ./server/src/server.go

test-server:
	go test -v ./server/src/...

run-worker:
	go run -v ./worker/src/worker.go

test-worker:
	go test -v ./worker/src/...

backup-db:
	python3 ./local_db/scripts/dynamodump.py -m backup -s "*" --host localhost --port 8000 --accessKey local --secretKey local --region local --dumpPath ./local_db/db_dump

backup-production:
	python3 ./local_db/scripts/dynamodump.py -m backup -r us-east-2 -s "*" --accessKey ${PRODUCTION_AWS_ACCESS_KEY_ID} --secretKey ${PRODUCTION_AWS_SECRET_ACCESS_KEY} --dumpPath ./local_db/db_dump

restore-db:
	python3 ./local_db/scripts/dynamodump.py -m restore -s "*" --host localhost --port 8000 --accessKey local --secretKey local --region local --dumpPath ./local_db/db_dump
