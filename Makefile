start-db:
	ttab "docker run -p 8000:8000 amazon/dynamodb-local -jar DynamoDBLocal.jar -sharedDb"
	ttab "AWS_REGION=local AWS_ACCESS_KEY_ID=local AWS_SECRET_ACCESS_KEY=local DYNAMO_ENDPOINT=http://localhost:8000 dynamodb-admin"

backup-db:
	python ./scripts/dynamodump.py -m backup -r local -s "*" --host localhost --port 8000 --accessKey local --secretKey local --dumpPath ./db_dump

restore-db:
	python ./scripts/dynamodump.py -m restore -r local -s "*" --host localhost --port 8000 --accessKey local --secretKey local --dumpPath ./db_dump
