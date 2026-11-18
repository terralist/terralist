package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"terralist/pkg/auth"
	"terralist/pkg/session/cookie"

	"github.com/gin-gonic/gin"
	. "github.com/smartystreets/goconvey/convey"
)

func TestParseActiveSession_CrossOrigin(t *testing.T) {
	Convey("Subject: Cross-origin protection for session-authenticated requests", t, func() {
		gin.SetMode(gin.TestMode)

		store, err := (&cookie.Creator{}).New(&cookie.Config{
			Name:   "_session",
			Secret: "0123456789abcdef0123456789abcdef",
		})
		So(err, ShouldBeNil)

		loginReq := httptest.NewRequest(http.MethodGet, "https://terralist.example/", nil)
		sess, err := store.New(loginReq)
		So(err, ShouldBeNil)
		sess.Set("user", &auth.User{Name: "alice", Email: "alice@example.com"})

		w := httptest.NewRecorder()
		So(store.Save(loginReq, w, sess), ShouldBeNil)
		sessionCookie := w.Result().Cookies()[0]

		crossOrigin := http.NewCrossOriginProtection()
		So(crossOrigin.AddTrustedOrigin("https://terralist.example"), ShouldBeNil)

		a := &Authentication{Store: store, CrossOrigin: crossOrigin}

		parse := func(method string, headers map[string]string) (*auth.User, error) {
			r := httptest.NewRequest(method, "https://terralist.example/v1/api/modules/acme/vpc/aws/9.9.9/upload", nil)
			r.AddCookie(sessionCookie)
			for k, v := range headers {
				if k == "Host" {
					r.Host = v
					continue
				}
				r.Header.Set(k, v)
			}

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = r

			return a.parseActiveSession(c)
		}

		Convey("When a cross-site POST carries the session cookie", func() {
			user, err := parse(http.MethodPost, map[string]string{
				"Sec-Fetch-Site": "cross-site",
				"Origin":         "https://attacker.example",
			})

			Convey("Then the request is rejected", func() {
				So(user, ShouldBeNil)
				So(errors.Is(err, ErrCrossOrigin), ShouldBeTrue)
			})
		})

		Convey("When an older browser sends a POST with a foreign Origin", func() {
			user, err := parse(http.MethodPost, map[string]string{
				"Origin": "https://attacker.example",
			})

			Convey("Then the request is rejected", func() {
				So(user, ShouldBeNil)
				So(errors.Is(err, ErrCrossOrigin), ShouldBeTrue)
			})
		})

		Convey("When a same-origin POST carries the session cookie", func() {
			user, err := parse(http.MethodPost, map[string]string{
				"Sec-Fetch-Site": "same-origin",
				"Origin":         "https://terralist.example",
			})

			Convey("Then the user is authenticated", func() {
				So(err, ShouldBeNil)
				So(user.Name, ShouldEqual, "alice")
			})
		})

		Convey("When a POST comes from the configured URL through a proxy that rewrites Host", func() {
			user, err := parse(http.MethodPost, map[string]string{
				"Origin": "https://terralist.example",
				"Host":   "terralist.internal:5758",
			})

			Convey("Then the user is authenticated", func() {
				So(err, ShouldBeNil)
				So(user.Name, ShouldEqual, "alice")
			})
		})

		Convey("When a cross-site GET carries the session cookie", func() {
			user, err := parse(http.MethodGet, map[string]string{
				"Sec-Fetch-Site": "cross-site",
			})

			Convey("Then the user is authenticated", func() {
				So(err, ShouldBeNil)
				So(user.Name, ShouldEqual, "alice")
			})
		})

		Convey("When a non-browser client sends a POST without Origin headers", func() {
			user, err := parse(http.MethodPost, nil)

			Convey("Then the user is authenticated", func() {
				So(err, ShouldBeNil)
				So(user.Name, ShouldEqual, "alice")
			})
		})
	})
}
