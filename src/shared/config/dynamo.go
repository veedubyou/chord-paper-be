package config

type Dynamo interface {
	DynamoConfig()
}

var _ Dynamo = ProdDynamo{}

type ProdDynamo struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
}

func (p ProdDynamo) DynamoConfig() {}

var _ Dynamo = LocalDynamo{}

type LocalDynamo struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Host            string
}

func (l LocalDynamo) DynamoConfig() {}
