start-db:
	ttab -w "docker run -p 8000:8000 amazon/dynamodb-local -jar DynamoDBLocal.jar -sharedDb"
	ttab "AWS_REGION=local AWS_ACCESS_KEY_ID=local AWS_SECRET_ACCESS_KEY=local DYNAMO_ENDPOINT=http://localhost:8000 dynamodb-admin"
	sleep 5
	make restore-db

backup-db:
	python3 ./scripts/dynamodump.py -m backup -s "*" --host localhost --port 8000 --accessKey local --secretKey local --region local --dumpPath ./db_dump

backup-production:
	python3 ./scripts/dynamodump.py -m backup -r us-east-2 -s "*" --accessKey ${PRODUCTION_AWS_ACCESS_KEY_ID} --secretKey ${PRODUCTION_AWS_SECRET_ACCESS_KEY} --dumpPath ./db_dump

restore-db:
	python3 ./scripts/dynamodump.py -m restore -s "*" --host localhost --port 8000 --accessKey local --secretKey local --region local --dumpPath ./db_dump
