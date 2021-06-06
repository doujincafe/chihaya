package cutenanami

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ApprovalInfo struct {
	ApprovedTorrents []string
	ApprovedClients  []string
	ApprovedUsers    []string
}

type NanamiCommunication struct {
	config          Config
	approvalChannel chan PeriodicUpdateResult
}

type PeriodicUpdateResult struct {
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

func (c NanamiCommunication) GetPeriodicUpdate() {
	for {
		approvalInfo, err := c.RequestApprovalInformation()
		c.approvalChannel <- PeriodicUpdateResult{approvalInfo, err}
		time.Sleep(time.Second)
	}
}

func NewNanamiCommunication(config Config) NanamiCommunication {
	communication := NanamiCommunication{config, make(chan PeriodicUpdateResult)}
	go communication.GetPeriodicUpdate()
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
}
