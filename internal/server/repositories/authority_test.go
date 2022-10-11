package repositories

import (
	"terralist/internal/server/models/authority"
	"terralist/pkg/database/entity"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mazen160/go-random"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	authorityRows = []string{"name", "policy_url", "owner"}
	keyRows       = []string{"authority_id", "key_id", "ascii_armor", "trust_signature"}
)

func TestAuthorityUpsert(t *testing.T) {
	Convey("Subject: Updating an existing authority", t, func() {
		db, mockDB, err := newMockDatabase()
		So(err, ShouldBeNil)

		authorityRepository := &DefaultAuthorityRepository{
			Database: db,
		}

		q := newQueryConstructor((authority.Authority{}).TableName())
		qk := newQueryConstructor((authority.Key{}).TableName())
		qak := newQueryConstructor((authority.ApiKey{}).TableName())

		Convey("Given an authority", func() {
			authorityID, _ := uuid.NewRandom()

			name, _ := random.String(16)
			policyURL, _ := random.String(16)
			owner, _ := random.String(16)

			keyID, _ := uuid.NewRandom()
			keyGPGId, _ := random.String(16)

			now := time.Now()

			currentKey := authority.Key{
				Entity: entity.Entity{
					ID:        keyID,
					CreatedAt: now,
					UpdatedAt: now,
				},
				AuthorityID: authorityID,
				KeyId:       keyGPGId,
			}

			current := authority.Authority{
				Entity: entity.Entity{
					ID:        authorityID,
					CreatedAt: now,
					UpdatedAt: now,
				},
				Name:      name,
				PolicyURL: policyURL,
				Owner:     owner,
				Keys:      []authority.Key{currentKey},
			}

			mockDB.
				ExpectQuery(q(`SELECT * FROM "%s" WHERE id = $1 ORDER BY "%s"."id" LIMIT 1`)).
				WithArgs(authorityID.String()).
				WillReturnRows(
					newRows(authorityRows).
						AddRow(current.ID, current.CreatedAt, current.UpdatedAt, current.Name, current.PolicyURL, current.Owner),
				)

			mockDB.
				ExpectQuery(qak(`"%s"."authority_id" = $1`)).
				WithArgs(authorityID.String()).
				WillReturnRows(newRows(apiKeyRows))

			mockDB.
				ExpectQuery(qk(`"%s"."authority_id" = $1`)).
				WithArgs(authorityID.String()).
				WillReturnRows(
					newRows(keyRows).
						AddRow(currentKey.ID, currentKey.CreatedAt, currentKey.UpdatedAt, currentKey.AuthorityID, currentKey.KeyId, currentKey.AsciiArmor, currentKey.TrustSignature),
				)

			Convey("If the policy URL is changed", func() {
				newPolicyURL, _ := random.String(16)
				updated := authority.Authority{
					Entity: entity.Entity{
						ID: authorityID,
					},
					PolicyURL: newPolicyURL,
				}

				mockDB.ExpectBegin()
				mockDB.
					ExpectExec(q(`UPDATE "%s" SET "created_at"=$1,"updated_at"=$2,"name"=$3,"policy_url"=$4,"owner"=$5 WHERE "id" = $6`)).
					WithArgs(AnyTime, AnyTime, AnyString, updated.PolicyURL, AnyString, updated.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mockDB.ExpectCommit()

				Convey("When the repository is queried", func() {
					actual, err := authorityRepository.Upsert(updated)

					Convey("The returned authority should have the policy URL updated", func() {
						So(actual, ShouldNotBeNil)
						So(err, ShouldBeNil)
						So(actual.PolicyURL, ShouldEqual, updated.PolicyURL)
					})
				})

				So(mockDB.ExpectationsWereMet(), ShouldBeNil)
			})

			Convey("If the policy URL is not changed", func() {

				Convey("When the repository is queried", func() {

					Convey("The returned authority should have the same policy URL", nil)

				})
			})

			Convey("If one of the existing keys is updated", func() {

				Convey("When the repository is queried", func() {

					Convey("The returned authority should have the key updated", nil)

				})
			})

			Convey("If a new key is added", func() {

				Convey("When the repository is queried", func() {

					Convey("The returned authority should have one more key", nil)

				})
			})

			Convey("If the database fails to update", func() {

				Convey("When the repository is queried", func() {

					Convey("A database failure error should be returned", nil)

				})
			})

			So(mockDB.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}
