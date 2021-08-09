package cutenanami

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const Buffer_size = 1000
const Batch_size = 1

// Approval info related
type ApprovalInfo struct {
	ApprovedTorrents []string `json:"approved_torrents"`
	ApprovedClients  []string `json:"approved_clients"`
	ApprovedUsers    []string `json:"approved_users"`
}

type NanamiCommunication struct {
	config                  Config
	approvalChannelOutbound chan PeriodicApprovalUpdateResult
	announceChannelInbound  chan SingleUserAnnounce
}

type PeriodicApprovalUpdateResult struct {
	info *ApprovalInfo
	err  error
}

func (c NanamiCommunication) RequestApprovalInformation() (approvalInfo *ApprovalInfo, err error) {

	// Perform GET to nanami
	resp, err := http.Get(c.config.NanamiAddress + "approval")
	if err != nil {
		return nil, err
	}

	// Read response
	defer resp.Body.Close()

	// Convert response to JSON
	var parsedInfo ApprovalInfo
	err = json.NewDecoder(resp.Body).Decode(&parsedInfo)
	if err != nil {
		return nil, err
	}

	// All good
	return &parsedInfo, nil
}

func (c NanamiCommunication) HandlePeriodicApprovalUpdate() {
	for {
		approvalInfo, err := c.RequestApprovalInformation()
		c.approvalChannelOutbound <- PeriodicApprovalUpdateResult{approvalInfo, err}
	}
}

func NewNanamiCommunication(config Config) NanamiCommunication {
	communication := NanamiCommunication{config, make(chan PeriodicApprovalUpdateResult), make(chan SingleUserAnnounce, Buffer_size)}
	go communication.HandlePeriodicApprovalUpdate()
	go communication.HandleAnnounceBatch()
	return communication
}

func PrintApprovalInfo(info *ApprovalInfo) {
	// Print contents
	fmt.Println("Allowed torrents")
	for _, id := range info.ApprovedTorrents {
		fmt.Println(id)
	}

	fmt.Println("Allowed clients")
	for _, id := range info.ApprovedClients {
		fmt.Println(id)
	}

	fmt.Println("Allowed users")
	for _, id := range info.ApprovedUsers {
		fmt.Println(id)
	}
}

// Batched announces
type SingleUserAnnounce struct {
	UserToken  string `json:"user_token"`
	Infohash   string `json:"infohash"`
	Event      uint8  `json:"event"`
	Downloaded uint64 `json:"downloaded"`
	Uploaded   uint64 `json:"uploaded"`
}

func (c NanamiCommunication) PushAnnounceBatch(announceBatch [Batch_size]SingleUserAnnounce) (err error) {
	// Serialize
	res, err := json.Marshal(announceBatch)
	if err != nil {
		return err
	}

	// POST to nanami
	resp, err := http.Post(
		c.config.NanamiAddress+"announce_batch",
		"application/json; charset=UTF-8",
		bytes.NewBuffer(res))

	fmt.Print(string(res))

	// Check errors
	if err != nil {
		return err
	}

	// Close response
	defer resp.Body.Close()

	// All good
	return nil
}

func (c NanamiCommunication) HandleAnnounceBatch() {
	// Main loop
	for {
		// Work loop, construct data for sending
		current_count := 0
		var arr [Batch_size]SingleUserAnnounce
		for current_count < Batch_size {
			announce := <-c.announceChannelInbound
			arr[current_count] = announce
			current_count++
		}

		err := c.PushAnnounceBatch(arr)
		if err != nil {
			fmt.Fprintln(os.Stdout, "Did not push batch because of error: ", err)
		}
	}
}
