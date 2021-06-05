// Package cutenanami implements a Hook that handles:
// - Torrent approval based on info hash
// - Torrent client approval
// - Aggregating and reporting announces to nanami
package cutenanami

import (
    "context"
    "fmt"

    yaml "gopkg.in/yaml.v2"

    "github.com/doujincafe/chihaya/bittorrent"
    "github.com/doujincafe/chihaya/middleware"
)

const Name = "cutenanami"

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

var ErrTorrentUnapproved = bittorrent.ClientError("unapproved torrent")
var ErrClientUnapproved = bittorrent.ClientError("unapproved client")
var ErrUserUnapproved = bittorrent.ClientError("unapproved user")

type Config struct {
    NanamiAddress string `yaml:"nanami_address"`
}

type hook struct {
    approvedTorrents   map[bittorrent.InfoHash]struct{}
    approvedClients    map[bittorrent.ClientID]struct{}
    approvedUsers      map[string]struct{}
}

func NewHook(cfg Config) (middleware.Hook, error) {
    h := &hook{
        approvedTorrents:   make(map[bittorrent.InfoHash]struct{}),
        approvedClients:    make(map[bittorrent.ClientID]struct{}),
        approvedUsers:      make(map[string]struct{}),
    }

    if (len(cfg.NanamiAddress) <= 0) {
        return nil, fmt.Errorf("nanami address not configured")
    }

    return h, nil
}


func ParseUserIdFromAnnounceUrl(announceUrl string) (string) {
    // TODO
    return "Hello world"
}

func (h *hook) HandleAnnounce(ctx context.Context, req *bittorrent.AnnounceRequest, resp *bittorrent.AnnounceResponse) (context.Context, error) {
    infohash := req.InfoHash
    clientId := bittorrent.NewClientID(req.Peer.ID)
    userId := ParseUserIdFromAnnounceUrl(req.Params.RawQuery())

    if _, found := h.approvedUsers[userId]; !found {
        return ctx, ErrUserUnapproved
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

