package cookie

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func savedCookie(r *http.Request) *http.Cookie {
	creator := &Creator{}
	store, err := creator.New(&Config{Name: "_session", Secret: "0123456789abcdef0123456789abcdef"})
	So(err, ShouldBeNil)

	sess, err := store.New(r)
	So(err, ShouldBeNil)
	sess.Set("key", "value")

	w := httptest.NewRecorder()
	So(store.Save(r, w, sess), ShouldBeNil)

	cookies := w.Result().Cookies()
	So(cookies, ShouldHaveLength, 1)

	return cookies[0]
}

func TestStoreSave(t *testing.T) {
	Convey("Subject: Session cookie attributes", t, func() {
		Convey("Given a plain HTTP request", func() {
			r := httptest.NewRequest(http.MethodGet, "http://terralist.example/", nil)
			c := savedCookie(r)

			Convey("Then the cookie is HttpOnly, SameSite=Lax and not Secure", func() {
				So(c.HttpOnly, ShouldBeTrue)
				So(c.SameSite, ShouldEqual, http.SameSiteLaxMode)
				So(c.Secure, ShouldBeFalse)
				So(c.Path, ShouldEqual, "/")
			})
		})

		Convey("Given a TLS request", func() {
			r := httptest.NewRequest(http.MethodGet, "https://terralist.example/", nil)
			r.TLS = &tls.ConnectionState{}
			c := savedCookie(r)

			Convey("Then the cookie is Secure", func() {
				So(c.Secure, ShouldBeTrue)
				So(c.SameSite, ShouldEqual, http.SameSiteLaxMode)
			})
		})

		Convey("Given a request forwarded by an HTTPS proxy", func() {
			r := httptest.NewRequest(http.MethodGet, "http://terralist.example/", nil)
			r.Header.Set("X-Forwarded-Proto", "https")
			c := savedCookie(r)

			Convey("Then the cookie is Secure", func() {
				So(c.Secure, ShouldBeTrue)
			})
		})
	})
}
