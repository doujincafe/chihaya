package cutenanami

import (
	"net/http"
	"encoding/json"
	"fmt"
)

type ApprovalInfo struct {
	approvedTorrents []string
	approvedClients  []string
	approvedUsers    []string
}

type INanamiCommunication interface {
	RequestApprovalInformation()(approvalInfo *ApprovalInfo, err error)
}

type NanamiCommunication struct {
	config Config
}

func New(config Config) NanamiCommunication {
	communication := NanamiCommunication{config}
	return communication;
}

func (c NanamiCommunication) RequestApprovalInformation()(approvalInfo *ApprovalInfo, err error) {

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