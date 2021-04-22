start-db:
	ttab -w "docker run -p 8000:8000 amazon/dynamodb-local -jar DynamoDBLocal.jar -sharedDb"
	ttab "AWS_REGION=local AWS_ACCESS_KEY_ID=local AWS_SECRET_ACCESS_KEY=local DYNAMO_ENDPOINT=http://localhost:8000 dynamodb-admin"
	sleep 5
	make restore-db

backup-db:
	python ./scripts/dynamodump.py -m backup -r local -s "*" --host localhost --port 8000 --accessKey local --secretKey local --dumpPath ./db_dump

backup-production:
	python ./scripts/dynamodump.py -m backup -r us-east-2 -s "*" --accessKey ${ACCESS_KEY_ID} --secretKey ${SECRET_ACCESS_KEY} --dumpPath ./db_dump

restore-db:
	python ./scripts/dynamodump.py -m restore -r local -s "*" --host localhost --port 8000 --accessKey local --secretKey local --dumpPath ./db_dump
