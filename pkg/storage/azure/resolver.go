// FILE: pkg/storage/azure/resolver.go
package azure

import (
	"context"
	"fmt"
	"log"
	"time"

	"terralist/pkg/storage"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
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

	if r.DefaultCredentials {
		serviceClient := r.Client.ServiceClient()

		// Prepare SAS Token Input field
		now := time.Now().UTC().Add(-10 * time.Second)
		expiry := now.Add(48 * time.Hour)
		info := service.KeyInfo{
			Start:  to.Ptr(now.UTC().Format(sas.TimeFormat)),
			Expiry: to.Ptr(expiry.UTC().Format(sas.TimeFormat)),
		}

		// Get Delegation Key
		udc, err := serviceClient.GetUserDelegationCredential(context.TODO(), info, nil)
		if err != nil {
			return "", fmt.Errorf("could not get UserDelegationCredential: %v", err)
		}

		sasQueryParams, err := sas.BlobSignatureValues{
			Protocol:      sas.ProtocolHTTPS,
			StartTime:     time.Now().UTC().Add(time.Second * -10),
			ExpiryTime:    time.Now().UTC().Add(15 * time.Minute),
			Permissions:   to.Ptr(sas.BlobPermissions{Read: true, List: true}).String(),
			ContainerName: r.ContainerName,
			BlobName:      blobName,
		}.SignWithUserDelegation(udc)
		if err != nil {
			log.Fatal(err.Error())
		}

		sasURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", r.AccountName, r.ContainerName, blobName) + "?" + sasQueryParams.Encode()
		return sasURL, nil
	}
	start := time.Now()
	expiry := start.Add(15 * time.Minute)

	return r.Client.ServiceClient().NewContainerClient(r.ContainerName).NewBlobClient(blobName).GetSASURL(
		sas.BlobPermissions{Read: true, List: true},
		expiry,
		&blob.GetSASURLOptions{StartTime: &start},
	)

}
func (r *Resolver) Find(keys string) (string, error) {

	pager := r.Client.NewListBlobsFlatPager(r.ContainerName, &container.ListBlobsFlatOptions{})

	for pager.More() {
		page, err := pager.NextPage(context.Background())
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
