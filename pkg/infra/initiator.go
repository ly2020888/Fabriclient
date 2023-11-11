package infra

import (
	"github.com/pkg/errors"
)

func StartCreateProposal(args []string, config *Config, crypto *Crypto, raw chan *Elements, errorCh chan error) {

	prop, err := CreateProposal(
		crypto,
		config.Channel,
		config.Chaincode,
		config.Version,
		args...,
	)
	if err != nil {
		errorCh <- errors.Wrapf(err, "error creating proposal")
		return
	}

	raw <- &Elements{Proposal: prop}

}
