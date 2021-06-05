// Package cutenanami implements a Hook that handles:
// - Torrent approval based on info hash
// - Torrent client approval
// - Aggregating and reporting announces to nanami
package cutenanami

import (
    "context"
    "encoding/hex"
    "fmt"

    yaml "gopkg.in/yaml.v2"

    "github.com/doujincafe/chihaya/bittorrent"
    "github.com/doujincafe/chihaya/middleware"
)

// Name is the name by which this middleware is registered with Chihaya.
const Name = "nanami is cutest"

func init() {
    middleware.RegisterDriver(Name, driver{})
}

var _ middleware.Driver = driver{}

type driver struct{}

func (d driver) NewHook(optionBytes []byte) (middleware.Hook, error) {
    var cfg Config
    err := yaml.Unmarshal(optionBytes, &cfg)
    if err != nil {
        return nil, fmt.Errorf("invalid options for middleware %s: %s", Name, err)
    }

    return NewHook(cfg)
}

// ErrTorrentUnapproved is the error returned when a torrent hash is invalid.
var ErrTorrentUnapproved = bittorrent.ClientError("unapproved torrent")
var ErrClientUnapproved = bittorrent.ClientError("unapproved client")
var ErrUserUnapproved = bittorrent.ClientError("unapproved user")

// Config for middleware. Address of nanami endpoint
type Config struct {
    NanamiAddress string `yaml:"nanami_address"`
}

type hook struct {
    approvedTorrents   map[bittorrent.InfoHash]struct{}
    approvedClients    map[bittorrent.ClientId]struct{}
    approvedUsers      map[string]struct{}
}

// NewHook returns an instance of the torrent approval middleware.
func NewHook(cfg Config) (middleware.Hook, error) {
    h := &hook{
        approvedTorrents:   make(map[bittorrent.InfoHash]struct{}),
        approvedClients:    make(map[bittorrent.ClientId]struct{}),
        approvedUsers:      make(map[string]struct{}),
    }

    if (len(cfg.NanamiAddress) <= 0) {
        return nil, fmt.Errorf("nanami address not configured")
    }

    return h, nil
}

func (h *hook) HandleAnnounce(ctx context.Context, req *bittorrent.AnnounceRequest, resp *bittorrent.AnnounceResponse) (context.Context, error) {
    infohash := req.InfoHash
    clientId := bittorrent.NewClientID(req.Peer.ID)
    userId := ParseUserIdFromAnnounceUrl(req.Params.RawQuery())

    if _, found := h.approvedUsers[userId]; !found {
        return ctr, ErrUserUnapproved
    }

    if _, found := h.approvedTorrents[infohash]; !found {
        return ctx, ErrTorrentUnapproved
    }

    if _, found := h.approvedClients[clientId]; !found {
        return ctx, ErrClientUnapproved
    }

    return ctx, nil
}

func (h *hook) HandleScrape(ctx context.Context, req *bittorrent.ScrapeRequest, resp *bittorrent.ScrapeResponse) (context.Context, error) {
    // Scrapes don't require any protection.
    return ctx, nil
}

func (h *hook) ParseUserIdFromAnnounceUrl(string announceUrl) (string) {
    // TODO
    return "Hello world"
}
