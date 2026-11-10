package main

import (
	"context"

	"github.com/ferforastieri/mira/internal/accounts"
	"github.com/ferforastieri/mira/internal/library"
	"github.com/ferforastieri/mira/internal/playback"
	"github.com/ferforastieri/mira/internal/virtualtv"
)

// virtualChannels has the library list the channels Mira airs, whose
// one version is their playout.
type virtualChannels struct {
	store  *virtualtv.Store
	player *playback.Service
}

func (v virtualChannels) Channels(ctx context.Context) ([]virtualtv.Channel, error) {
	return v.store.Channels(ctx)
}

func (v virtualChannels) Version(item accounts.ID, channel virtualtv.Channel) library.Version {
	return v.player.PlayoutVersion(item, channel)
}

// playouts has playback air the channels on, their programmes played
// from the versions the server's addons and local folders list, by the
// server's version rules.
type playouts struct {
	store *virtualtv.Store
	lib   *library.Service
}

func (p playouts) Channel(ctx context.Context, id accounts.ID) (virtualtv.Channel, error) {
	channel, err := p.store.Channel(ctx, id)
	if err == nil && !channel.Enabled {
		return virtualtv.Channel{}, virtualtv.ErrNotFound
	}
	return channel, err
}

func (p playouts) Versions(ctx context.Context, item accounts.ID) ([]library.Version, error) {
	return p.lib.Versions(ctx, accounts.User{}, item)
}
