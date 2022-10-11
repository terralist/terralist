package repositories

import (
	"errors"
	"terralist/internal/server/models/authority"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	apiKeyRows = []string{"authority_id", "expiration"}
)

func TestApiKeyFind(t *testing.T) {
	Convey("Subject: Finding an API key", t, func() {
		db, mockDB, err := newMockDatabase()
		So(err, ShouldBeNil)

		apiKeyRepository := &DefaultApiKeyRepository{
			Database: db,
		}

		q := newQueryConstructor((authority.ApiKey{}).TableName())

		Convey("Given an API key ID", func() {
			apiKeyID, _ := uuid.NewRandom()

			Convey("If the API key does not exist in the database", func() {
				mockDB.
					ExpectQuery(q(`SELECT * FROM "%s" WHERE id = $1 ORDER BY "%s"."id" LIMIT 1`)).
					WithArgs(apiKeyID.String()).
					WillReturnError(errors.New(""))

				Convey("When the repository is queried", func() {
					apiKey, err := apiKeyRepository.Find(apiKeyID)

					Convey("An error should be returned", func() {
						So(apiKey, ShouldBeNil)
						So(err, ShouldNotBeNil)
					})
				})

				So(mockDB.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("If the API key exists in the database", func() {
				authorityID, _ := uuid.NewRandom()

				Convey("If the API key has no expiration", func() {
					mockDB.
						ExpectQuery(q(`SELECT * FROM "%s" WHERE id = $1 ORDER BY "%s"."id" LIMIT 1`)).
						WithArgs(apiKeyID.String()).
						WillReturnRows(
							newRows(apiKeyRows).
								AddRow(apiKeyID, time.Now(), time.Now(), authorityID, nil),
						)

					Convey("When the repository is queried", func() {
						apiKey, err := apiKeyRepository.Find(apiKeyID)

						Convey("The API key should be returned", func() {
							So(apiKey, ShouldNotBeNil)
							So(apiKey.ID, ShouldEqual, apiKeyID)
							So(apiKey.Expiration, ShouldBeNil)
							So(err, ShouldBeNil)
						})
					})

					So(mockDB.ExpectationsWereMet(), ShouldBeNil)
				})

				Convey("If the API key expired", func() {
					mockDB.
						ExpectQuery(q(`SELECT * FROM "%s" WHERE id = $1 ORDER BY "%s"."id" LIMIT 1`)).
						WithArgs(apiKeyID.String()).
						WillReturnRows(
							newRows(apiKeyRows).
								AddRow(apiKeyID, time.Now(), time.Now(), authorityID, time.Now().Add(time.Duration(-1)*time.Hour)),
						)

					mockDB.ExpectBegin()
					mockDB.
						ExpectExec(q(`DELETE FROM "%s" WHERE "authority_api_keys"."id" = $1`)).
						WithArgs(apiKeyID.String()).
						WillReturnResult(sqlmock.NewResult(1, 1))
					mockDB.ExpectCommit()

					Convey("When the repository is queried", func() {
						apiKey, err := apiKeyRepository.Find(apiKeyID)

						Convey("An expired error should be returned", func() {
							So(apiKey, ShouldBeNil)
							So(err, ShouldResemble, ErrApiKeyExpired)
						})
					})

					So(mockDB.ExpectationsWereMet(), ShouldBeNil)
				})

				Convey("If the API key did not expire", func() {
					mockDB.
						ExpectQuery(q(`SELECT * FROM "%s" WHERE id = $1 ORDER BY "%s"."id" LIMIT 1`)).
						WithArgs(apiKeyID.String()).
						WillReturnRows(
							newRows(apiKeyRows).
								AddRow(apiKeyID, time.Now(), time.Now(), authorityID, time.Now().Add(time.Hour)),
						)

					Convey("When the repository is queried", func() {
						apiKey, err := apiKeyRepository.Find(apiKeyID)

						Convey("The API key should be returned", func() {
							So(apiKey, ShouldNotBeNil)
							So(apiKey.ID, ShouldEqual, apiKeyID)
							So(apiKey.Expiration, ShouldNotBeNil)
							So(err, ShouldBeNil)
						})
					})

					So(mockDB.ExpectationsWereMet(), ShouldBeNil)
				})
			})
		})
	})
}
