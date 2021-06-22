package scrobbler

import (
	"context"
	"time"

	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/model/request"
	"github.com/navidrome/navidrome/tests"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Broker", func() {
	var ctx context.Context
	var ds model.DataStore
	var broker Broker
	var track model.MediaFile
	var fake *fakeScrobbler
	BeforeEach(func() {
		ctx = context.Background()
		ctx = request.WithUser(ctx, model.User{ID: "u-1"})
		ds = &tests.MockDataStore{}
		broker = GetBroker(ds)
		fake = &fakeScrobbler{}
		Register("fake", func(ds model.DataStore) Scrobbler {
			return fake
		})

		track = model.MediaFile{
			ID:          "123",
			Title:       "Track Title",
			Album:       "Track Album",
			Artist:      "Track Artist",
			AlbumArtist: "Track AlbumArtist",
			TrackNumber: 1,
			Duration:    180,
			MbzTrackID:  "mbz-123",
		}
		_ = ds.MediaFile(ctx).Put(&track)
	})

	Describe("NowPlaying", func() {
		It("sends track to agent", func() {
			err := broker.NowPlaying(ctx, "player-1", "player-one", "123")
			Expect(err).ToNot(HaveOccurred())
			Expect(fake.UserID).To(Equal("u-1"))
			Expect(fake.Track.ID).To(Equal("123"))
		})
	})

	Describe("GetNowPlaying", func() {
		BeforeEach(func() {
			ctx = context.Background()
		})
		It("returns current playing music", func() {
			track2 := track
			track2.ID = "456"
			_ = ds.MediaFile(ctx).Put(&track)
			ctx = request.WithUser(ctx, model.User{UserName: "user-1"})
			_ = broker.NowPlaying(ctx, "player-1", "player-one", "123")
			ctx = request.WithUser(ctx, model.User{UserName: "user-2"})
			_ = broker.NowPlaying(ctx, "player-2", "player-two", "456")

			playing, err := broker.GetNowPlaying(ctx)

			Expect(err).ToNot(HaveOccurred())
			Expect(playing).To(HaveLen(2))
			Expect(playing[0].PlayerId).To(Equal("player-2"))
			Expect(playing[0].PlayerName).To(Equal("player-two"))
			Expect(playing[0].Username).To(Equal("user-2"))
			Expect(playing[0].TrackID).To(Equal("456"))

			Expect(playing[1].PlayerId).To(Equal("player-1"))
			Expect(playing[1].PlayerName).To(Equal("player-one"))
			Expect(playing[1].Username).To(Equal("user-1"))
			Expect(playing[1].TrackID).To(Equal("123"))
		})
	})

	Describe("Submit", func() {
		It("sends track to agent", func() {
			ctx = request.WithUser(ctx, model.User{ID: "u-1", UserName: "user-1"})
			ts := time.Now()

			err := broker.Submit(ctx, "123", ts)

			Expect(err).ToNot(HaveOccurred())
			Expect(fake.UserID).To(Equal("u-1"))
			Expect(fake.Scrobbles[0].ID).To(Equal("123"))
		})
	})

})

type fakeScrobbler struct {
	UserID    string
	Track     *model.MediaFile
	Scrobbles []Scrobble
	Error     error
}

func (f *fakeScrobbler) NowPlaying(ctx context.Context, userId string, track *model.MediaFile) error {
	if f.Error != nil {
		return f.Error
	}
	f.UserID = userId
	f.Track = track
	return nil
}

func (f *fakeScrobbler) Scrobble(ctx context.Context, userId string, scrobbles []Scrobble) error {
	if f.Error != nil {
		return f.Error
	}
	f.UserID = userId
	f.Scrobbles = scrobbles
	return nil
}