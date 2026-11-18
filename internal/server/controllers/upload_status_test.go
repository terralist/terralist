package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"terralist/internal/server/services"
	"terralist/pkg/file"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUploadStatus(t *testing.T) {
	Convey("Subject: The status answering a failed upload", t, func() {
		cases := []struct {
			name   string
			err    error
			status int
		}{
			{"a download failure", file.ErrDownloadFailure, http.StatusBadGateway},
			{"a download interrupt", file.ErrDownloadInterrupt, http.StatusBadGateway},
			{"a refused source", file.ErrSourceRefused, http.StatusUnprocessableEntity},
			{"a checksum mismatch", file.ErrChecksumMismatch, http.StatusUnprocessableEntity},
			{"an invalid archive", file.ErrInvalidArchive, http.StatusUnprocessableEntity},
			{"a storage failure", services.ErrStorageFailure, http.StatusInternalServerError},
			{"a system failure", file.ErrSystemFailure, http.StatusInternalServerError},
			{"a business refusal", errors.New("version 1.0.0 already exists"), http.StatusConflict},
		}

		for _, c := range cases {
			Convey("Given "+c.name+", wrapped by the service", func() {
				err := fmt.Errorf("could not fetch shaSums file: %w", c.err)

				So(uploadStatus(err), ShouldEqual, c.status)
			})
		}
	})
}
