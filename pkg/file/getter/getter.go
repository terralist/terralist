package getter

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-getter"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"sync"
)

var (
	errorSystemFailure     = errors.New("system failure")
	errorDownloadFailure   = errors.New("download failure")
	errorDownloadInterrupt = errors.New("download interrupt")
)

var (
	ErrSystemFailure     = func() error { return errorSystemFailure }()
	ErrDownloadFailure   = func() error { return errorDownloadFailure }()
	ErrDownloadInterrupt = func() error { return errorDownloadInterrupt }()
)

// Get downloads a file/directory from a given URL and stores it to a given destination
// The destination must be a directory (it doesn't have to exist)
// If the URL points to a file, the file will be downloaded to the destination and
// 		stored with the same name as the URL basename
// If the URL points to a directory, the directory will be downloaded to the destination
//		with depth of 1
// If the URL points to an archive containing a directory, it will be unarchived and
//		stored as a directory
func Get(url string, destination string) error {
	log.Info().
		Str("URL", url).
		Str("Destination", destination).
		Msg("Getting file")

	// Initialize HC getter
	ctx, cancel := context.WithCancel(context.Background())
	client := &getter.Client{
		Ctx:  ctx,
		Src:  url,
		Dst:  destination,
		Pwd:  destination,
		Mode: getter.ClientModeAny,
		Options: []getter.ClientOption{
			getter.WithInsecure(),
		},
	}

	// Launch the download process
	wg := sync.WaitGroup{}
	wg.Add(1)
	ech := make(chan error, 2)
	go func() {
		defer wg.Done()
		defer cancel()

		log.Debug().
			Str("URL", url).
			Str("Destination", destination).
			Msg("Download started")

		if err := client.Get(); err != nil {
			ech <- err
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)

	// Wait for the download process to finish
	select {
	case sig := <-sc:
		signal.Reset(os.Interrupt)
		cancel()
		wg.Wait()
		log.Info().
			Str("Signal", sig.String()).
			Msg("Download interrupted by signal")

		return fmt.Errorf("%w: signal %v received", ErrDownloadInterrupt, sig.String())
	case err := <-ech:
		wg.Wait()

		log.Error().
			Str("URL", url).
			Str("Destination", destination).
			AnErr("Error", err).
			Msg("Download failed")

		return fmt.Errorf("%w: %v", ErrDownloadFailure, err)
	case <-ctx.Done():
		wg.Wait()

		log.Debug().
			Str("URL", url).
			Str("Destination", destination).
			Msg("Download completed")

		return nil
	}
}
