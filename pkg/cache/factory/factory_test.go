package factory

import (
	"context"
	"errors"
	"testing"
	"time"

	"terralist/pkg/cache"
	"terralist/pkg/cache/memory"

	. "github.com/smartystreets/goconvey/convey"
)

type invalidConfig struct{}

func (invalidConfig) SetDefaults() {}
func (invalidConfig) Validate() error {
	return errors.New("invalid")
}

func TestNewCache(t *testing.T) {
	Convey("Subject: Creating a cache from a backend name", t, func() {
		Convey("When the memory backend is requested", func() {
			c, err := NewCache(cache.MEMORY, &memory.Config{})

			Convey("Then a working cache wrapped with metrics should be returned", func() {
				So(err, ShouldBeNil)
				So(c, ShouldHaveSameTypeAs, &cache.MetricsCache{})
				So(c.Set(context.Background(), "key", []byte("value"), time.Hour), ShouldBeNil)
				value, ok, err := c.Get(context.Background(), "key")
				So(err, ShouldBeNil)
				So(ok, ShouldBeTrue)
				So(string(value), ShouldEqual, "value")
			})
		})

		Convey("When the configuration is invalid", func() {
			_, err := NewCache(cache.MEMORY, invalidConfig{})

			Convey("Then it should fail", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When an unknown backend is requested", func() {
			_, err := NewCache(cache.Backend(99), &memory.Config{})

			Convey("Then it should fail", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
