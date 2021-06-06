// Package cutenanami implements a Hook that handles:
// - Torrent approval based on info hash
// - Torrent client approval
// - Aggregating and reporting announces to nanami
// - Getting list of whitelisted clients, torrents and users
package cutenanami

import (
	"context"
	"fmt"
	"os"

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
	approvedTorrents map[bittorrent.InfoHash]struct{}
	approvedClients  map[bittorrent.ClientID]struct{}
	approvedUsers    map[string]struct{}
	communication    NanamiCommunication
}

func NewHook(cfg Config) (middleware.Hook, error) {
	h := &hook{
		approvedTorrents: make(map[bittorrent.InfoHash]struct{}),
		approvedClients:  make(map[bittorrent.ClientID]struct{}),
		approvedUsers:    make(map[string]struct{}),
		communication:    NewNanamiCommunication(cfg),
	}

	if len(cfg.NanamiAddress) <= 0 {
		return nil, fmt.Errorf("nanami address not configured")
	}

	go h.UpdateApprovals()

	return h, nil
}

func ParseUserIdFromAnnounceUrl(announceUrl string) string {
	// TODO
	return "Hello world"
}

func (h *hook) HandleAnnounce(ctx context.Context, req *bittorrent.AnnounceRequest, resp *bittorrent.AnnounceResponse) (context.Context, error) {
	infohash := req.InfoHash
	clientId := bittorrent.NewClientID(req.Peer.ID)
	userId := ParseUserIdFromAnnounceUrl(req.Params.RawQuery())

	if _, found := h.approvedUsers[userId]; !found {
		//		return ctx, ErrUserUnapproved
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

func (h *hook) UpdateApprovals() {
	for {

		pair := <-h.communication.approvalChannel
		if pair.err == nil {
			// Approved torrents
			approvedTorrents := make(map[bittorrent.InfoHash]struct{})
			for _, str := range pair.info.ApprovedTorrents {
				if len(str) == 20 {
					infoHash := bittorrent.InfoHashFromString(str)
					approvedTorrents[infoHash] = struct{}{}
				} else {
					fmt.Println("Invalid format for whitelisted torrent: " + str)
				}

			}

			// Approved torrentClients
			approvedClients := make(map[bittorrent.ClientID]struct{})
			for _, str := range pair.info.ApprovedClients {
				var clientID bittorrent.ClientID
				copy(clientID[:], []byte(str))
				approvedClients[clientID] = struct{}{}
			}

			// Approved users
			// TODO

			// Swap reference
			h.approvedTorrents = approvedTorrents
			h.approvedClients = approvedClients
		} else {
			fmt.Fprintln(os.Stdout, "Did not update approvals because of error: ", pair.err)
		}
	}
}
