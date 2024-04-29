// FILE: pkg/storage/azure/resolver.go
package azure

import (
	"context"
	"fmt"
	"time"

	"terralist/pkg/storage"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
)

type Resolver struct {
	ContainerName string
	AccountName   string
	AccountKey    string
	Client        *azblob.Client

	// DefaultAzureCredentials *azidentity.DefaultAzureCredential

	DefaultCredentials bool
}

func (r *Resolver) Store(in *storage.StoreInput) (string, error) {
	// Create a new block blob URL using the container URL and the specified key
	key := fmt.Sprintf("%s/%s", in.KeyPrefix, in.FileName)

	ctx := context.Background()
	buffer := []byte(in.Content)

	_, err := r.Client.UploadBuffer(ctx,
		r.ContainerName,
		key,
		buffer,
		nil)
	if err != nil {
		return "", fmt.Errorf("could not upload archive: %v", err)
	}
	return key, nil

}

func (r *Resolver) GetSASURL(blobName string) (string, error) {

	//TODO: check https://github.com/Azure/azure-sdk-for-go/blob/main/sdk/storage/azblob/container/examples_test.go#L275

	start := time.Now()
	expiry := start.Add(15 * time.Minute)

	fmt.Printf("Blob Name: %s\n", blobName)

	return r.Client.ServiceClient().NewContainerClient(r.ContainerName).NewBlobClient(blobName).GetSASURL(
		sas.BlobPermissions{Read: true, List: true},
		expiry,
		&blob.GetSASURLOptions{StartTime: &start},
	)

}
func (r *Resolver) Find(keys string) (string, error) {
	pager := r.Client.NewListBlobsFlatPager(keys, &container.ListBlobsFlatOptions{Prefix: &keys})

	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return "", fmt.Errorf("could not get next page: %v", err)
		}
		for _, blob := range page.Segment.BlobItems {
			if *blob.Name == keys {
				url, err := r.GetSASURL(*blob.Name)
				if err != nil {
					return "", fmt.Errorf("could not get SAS URL: %v", err)
				}
				fmt.Printf("URL: %s\n", url)

				return url, nil
			}
		}
	}
	return "", fmt.Errorf("could not find: %s", keys)
}

func (r *Resolver) Purge(key string) error {
	// Implement the Purge method

	_, err := r.Client.DeleteBlob(context.Background(), r.ContainerName, key, nil)
	if err != nil {
		return fmt.Errorf("could not delete blob: %v", err)
	}
	return nil
}
