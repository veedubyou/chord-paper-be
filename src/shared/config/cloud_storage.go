package config

type CloudStorage interface {
	GetStorageHost() string
	GetBucket() string
}

var _ CloudStorage = ProdCloudStorage{}

type ProdCloudStorage struct {
	StorageHost string
	SecretKey   string
	BucketName  string
}

func (p ProdCloudStorage) GetStorageHost() string {
	return p.StorageHost
}

func (p ProdCloudStorage) GetBucket() string {
	return p.BucketName
}

var _ CloudStorage = LocalCloudStorage{}

type LocalCloudStorage struct {
	StorageHost  string
	HostEndpoint string
	BucketName   string
}

func (l LocalCloudStorage) GetStorageHost() string {
	return l.StorageHost
}

func (l LocalCloudStorage) GetBucket() string {
	return l.BucketName
}
