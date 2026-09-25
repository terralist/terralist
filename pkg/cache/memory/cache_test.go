package memory

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func newTestCache(now *time.Time) *Cache {
	return newCache(maxEntries, func() time.Time { return *now })
}

func TestMemoryCache(t *testing.T) {
	Convey("Subject: An in-memory cache", t, func() {
		ctx := context.Background()
		now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
		cache := newTestCache(&now)

		Convey("When a missing key is read", func() {
			value, ok, err := cache.Get(ctx, "missing")

			Convey("Then it should be a miss", func() {
				So(err, ShouldBeNil)
				So(ok, ShouldBeFalse)
				So(value, ShouldBeNil)
			})
		})

		Convey("Given a stored entry", func() {
			So(cache.Set(ctx, "key", []byte("value"), time.Hour), ShouldBeNil)

			Convey("When it is read before its retention ends", func() {
				now = now.Add(59 * time.Minute)
				value, ok, err := cache.Get(ctx, "key")

				Convey("Then it should be a hit", func() {
					So(err, ShouldBeNil)
					So(ok, ShouldBeTrue)
					So(string(value), ShouldEqual, "value")
				})
			})

			Convey("When it is read after its retention ended", func() {
				now = now.Add(time.Hour)
				_, ok, err := cache.Get(ctx, "key")

				Convey("Then it should be a miss", func() {
					So(err, ShouldBeNil)
					So(ok, ShouldBeFalse)
				})
			})

			Convey("When it is overwritten", func() {
				So(cache.Set(ctx, "key", []byte("other"), time.Hour), ShouldBeNil)
				value, ok, _ := cache.Get(ctx, "key")

				Convey("Then the latest value should be returned", func() {
					So(ok, ShouldBeTrue)
					So(string(value), ShouldEqual, "other")
				})
			})

			Convey("When it is deleted", func() {
				So(cache.Delete(ctx, "key"), ShouldBeNil)
				_, ok, _ := cache.Get(ctx, "key")

				Convey("Then it should be a miss", func() {
					So(ok, ShouldBeFalse)
				})
			})

			Convey("When the stored value is mutated by the caller", func() {
				value, _, _ := cache.Get(ctx, "key")
				value[0] = 'X'
				again, _, _ := cache.Get(ctx, "key")

				Convey("Then the cache should keep its own copy", func() {
					So(string(again), ShouldEqual, "value")
				})
			})
		})

		Convey("When an entry is stored without a positive retention", func() {
			err := cache.Set(ctx, "key", []byte("value"), 0)

			Convey("Then it should be rejected", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When deleting a missing key", func() {
			err := cache.Delete(ctx, "missing")

			Convey("Then it should not fail", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("Given expired and live entries", func() {
			So(cache.Set(ctx, "old", []byte("1"), time.Minute), ShouldBeNil)
			So(cache.Set(ctx, "fresh", []byte("2"), time.Hour), ShouldBeNil)
			now = now.Add(2 * time.Minute)

			Convey("When the cache is swept", func() {
				cache.sweep()

				Convey("Then only the expired entries should be gone", func() {
					So(cache.entries.Len(), ShouldEqual, 1)
					_, ok, _ := cache.Get(ctx, "fresh")
					So(ok, ShouldBeTrue)
				})
			})
		})
	})
}

func TestMemoryCacheEviction(t *testing.T) {
	Convey("Subject: An in-memory cache holding as many entries as it may", t, func() {
		ctx := context.Background()
		now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
		cache := newCache(2, func() time.Time { return now })

		So(cache.Set(ctx, "first", []byte("1"), time.Hour), ShouldBeNil)
		So(cache.Set(ctx, "second", []byte("2"), time.Hour), ShouldBeNil)

		Convey("When a new entry is stored", func() {
			So(cache.Set(ctx, "third", []byte("3"), time.Hour), ShouldBeNil)

			Convey("Then the least recently used entry should be evicted", func() {
				_, ok, _ := cache.Get(ctx, "first")
				So(ok, ShouldBeFalse)
				_, ok, _ = cache.Get(ctx, "second")
				So(ok, ShouldBeTrue)
				_, ok, _ = cache.Get(ctx, "third")
				So(ok, ShouldBeTrue)
			})
		})

		Convey("When an entry is read before a new one is stored", func() {
			_, ok, _ := cache.Get(ctx, "first")
			So(ok, ShouldBeTrue)
			So(cache.Set(ctx, "third", []byte("3"), time.Hour), ShouldBeNil)

			Convey("Then the entry read should be kept and the other evicted", func() {
				_, ok, _ := cache.Get(ctx, "first")
				So(ok, ShouldBeTrue)
				_, ok, _ = cache.Get(ctx, "second")
				So(ok, ShouldBeFalse)
			})
		})
	})
}

func TestCreator(t *testing.T) {
	Convey("Subject: Creating an in-memory cache", t, func() {
		Convey("Given a configuration without a sweep interval", func() {
			cfg := &Config{}
			cfg.SetDefaults()

			Convey("Then the interval should default", func() {
				So(cfg.Validate(), ShouldBeNil)
				So(cfg.SweepInterval, ShouldEqual, time.Minute)
			})
		})

		Convey("Given a negative sweep interval", func() {
			cfg := &Config{SweepInterval: -time.Second}

			Convey("Then validation should fail", func() {
				So(cfg.Validate(), ShouldNotBeNil)
			})
		})

		Convey("When a cache is created", func() {
			cfg := &Config{}
			cfg.SetDefaults()
			cache, err := (&Creator{}).New(cfg)

			Convey("Then it should be usable", func() {
				So(err, ShouldBeNil)
				So(cache.Set(context.Background(), "key", []byte("value"), time.Hour), ShouldBeNil)
				value, ok, _ := cache.Get(context.Background(), "key")
				So(ok, ShouldBeTrue)
				So(string(value), ShouldEqual, "value")
			})
		})

		Convey("When a cache is created with a foreign configurator", func() {
			_, err := (&Creator{}).New(nil)

			Convey("Then it should fail", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
