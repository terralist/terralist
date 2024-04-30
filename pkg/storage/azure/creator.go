// FILE: pkg/storage/azure/creator.go
package azure

import (
	"context"
	"fmt"
	"terralist/pkg/storage"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type Creator struct{}

func (t *Creator) New(config storage.Configurator) (storage.Resolver, error) {
	cfg, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("unsupported configurator")
	}

	var client *azblob.Client

	ctx := context.Background()

	if cfg.DefaultCredentials {
		defaultAzureCredentials, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("could not get DefaultAzureCredentials: %v", err)
		}
		client, err = azblob.NewClient(fmt.Sprintf("https://%s.blob.core.windows.net", cfg.AccountName), defaultAzureCredentials, nil)
		if err != nil {
			return nil, fmt.Errorf("could not create blob client: %v", err)
		}
	} else {
		// Creating new client with provided credentials
		creds, err := azblob.NewSharedKeyCredential(cfg.AccountName, cfg.AccountKey)
		if err != nil {
			return nil, fmt.Errorf("could not create shared key credentials: %v", err)
		}
		client, err = azblob.NewClientWithSharedKeyCredential(fmt.Sprintf("https://%s.blob.core.windows.net", cfg.AccountName), creds, nil)
		if err != nil {
			return nil, fmt.Errorf("could not create blob client: %v", err)
		}
	}

	// using the client check is the Container exists or create a new one
	pager := client.NewListContainersPager(&azblob.ListContainersOptions{
		Include: azblob.ListContainersInclude{Metadata: true, Deleted: true},
	})
	containerFound := false

	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("could not list containers: %v", err)
		}
		for _, _container := range resp.ContainerItems {
			if *_container.Name == cfg.ContainerName {
				containerFound = true
				break
			}
		}
		if containerFound {
			break
		}
	}

	if !containerFound {
		_, err := client.CreateContainer(ctx, cfg.ContainerName, nil)
		if err != nil {
			return nil, fmt.Errorf("could not create container: %v", err)
		}
	}

	return &Resolver{
		ContainerName: cfg.ContainerName,
		AccountName:   cfg.AccountName,
		AccountKey:    cfg.AccountKey,
		SASExpire:     cfg.SASExpire,
		Client:        client,

		DefaultCredentials: cfg.DefaultCredentials,
	}, nil
}
