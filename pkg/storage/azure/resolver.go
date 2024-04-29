// FILE: pkg/storage/azure/resolver.go
package azure

import (
	"context"
	"fmt"
	"time"

	"terralist/pkg/storage"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
)

type Resolver struct {
	ContainerName string
	AccountName   string
	AccountKey    string

	// DefaultAzureCredentials *azidentity.DefaultAzureCredential

	DefaultCredentials bool
}

func (r *Resolver) getClient() (*azblob.Client, error) {
	var client *azblob.Client
	// var err error

	if r.DefaultCredentials {
		defaultAzureCredentials, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("could not get DefaultAzureCredentials: %v", err)
		}
		client, err = azblob.NewClient(
			fmt.Sprintf("https://%s.blob.core.windows.net", r.AccountName),
			defaultAzureCredentials,
			nil)
		if err != nil {
			return nil, fmt.Errorf("could not get Client using DefaultAzureCredentials: %v", err)
		}
	} else {
		creds, err := azblob.NewSharedKeyCredential(r.AccountName, r.AccountKey)
		if err != nil {
			return nil, fmt.Errorf("could not get Shared Key Credentials: %v", err)
		}
		client, err = azblob.NewClientWithSharedKeyCredential(
			fmt.Sprintf("https://%s.blob.core.windows.net", r.AccountName),
			creds,
			nil)
		if err != nil {
			return nil, fmt.Errorf("could not get Client using Shared Key Credentials: %v", err)
		}
	}
	return client, nil
}

func (r *Resolver) Store(in *storage.StoreInput) (string, error) {
	// Create a new block blob URL using the container URL and the specified key
	key := fmt.Sprintf("%s/%s", in.KeyPrefix, in.FileName)

	ctx := context.Background()
	buffer := []byte(in.Content)

	var client *azblob.Client
	var err error

	client, err = r.getClient()
	if err != nil {
		return "", fmt.Errorf("could not get client: %v", err)
	}

	_, err = client.UploadBuffer(ctx,
		r.ContainerName,
		key,
		buffer,
		nil)
	if err != nil {
		return "", fmt.Errorf("could not upload archive: %v", err)
	}
	return key, nil

}

func (r *Resolver) GetSASURL(client azblob.Client, blobName string) (string, error) {

	//TODO: check https://github.com/Azure/azure-sdk-for-go/blob/main/sdk/storage/azblob/container/examples_test.go#L275

	// Set current and past time and create key
	now := time.Now().UTC().Add(-10 * time.Second)
	expiry := now.Add(48 * time.Hour)
	info := service.KeyInfo{
		Start:  to.Ptr(now.UTC().Format(sas.TimeFormat)),
		Expiry: to.Ptr(expiry.UTC().Format(sas.TimeFormat)),
	}

	sasQueryParams := sas.BlobSignatureValues{
		Protocol:      sas.ProtocolHTTPS,
		StartTime:     now,
		ExpiryTime:    expiry,
		Permissions:   to.Ptr(sas.ContainerPermissions{Read: true, List: true}).String(),
		BlobName:      blobName,
		ContainerName: r.ContainerName,
	}

	if r.DefaultCredentials {
		udc, err := client.ServiceClient().GetUserDelegationCredential(
			context.Background(),
			info,
			nil)
		if err != nil {
			return "", fmt.Errorf("could not get UserDelegationCredential: %v", err)
		}
		sasQueryParams, err := sasQueryParams.SignWithUserDelegation(udc)
		if err != nil {
			return "", fmt.Errorf("could not sign SAS query params: %v", err)
		}
		return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s?%s", r.AccountName, r.ContainerName, blobName, sasQueryParams.Encode()), nil
	} else {
		credential, err := azblob.NewSharedKeyCredential(r.AccountName, r.AccountKey)
		if err != nil {
			return "", fmt.Errorf("could not get Shared Key Credential: %v", err)
		}
		sasQueryParams, err := sasQueryParams.SignWithSharedKey(credential)
		if err != nil {
			return "", fmt.Errorf("could not sign SAS query params: %v", err)
		}
		return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s?%s", r.AccountName, r.ContainerName, blobName, sasQueryParams.Encode()), nil
		// return new(string), nil
	}
}

func (r *Resolver) Find(keys string) (string, error) {
	// Implement the Find method

	client, err := r.getClient()
	if err != nil {
		return "", fmt.Errorf("could not get client: %v", err)
	}
	pager := client.NewListBlobsFlatPager(keys, &container.ListBlobsFlatOptions{Prefix: &keys})

	// continue fetching pages until no more remain
	for pager.More() {
		// advance to the next page
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return "", fmt.Errorf("could not get next page: %v", err)
		}
		// print the blob names for this page
		for _, blob := range page.Segment.BlobItems {
			if *blob.Name == keys {
				url, err := r.GetSASURL(*client, *blob.Name)
				if err != nil {
					return "", fmt.Errorf("could not get SAS URL: %v", err)
				}
				return url, nil
			}
		}
	}

	return "", nil
}

func (r *Resolver) Purge(key string) error {
	// Implement the Purge method

	client, err := r.getClient()
	if err != nil {
		return fmt.Errorf("could not get client: %v", err)
	}
	_, err = client.DeleteBlob(context.Background(), r.ContainerName, key, nil)
	if err != nil {
		return fmt.Errorf("could not delete blob: %v", err)
	}
	return nil
}
