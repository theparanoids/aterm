// Copyright 2020, Verizon Media
// Licensed under the terms of the MIT. See LICENSE file in project root for terms.

package appdialogs

import (
	"errors"
	"fmt"

	"github.com/theparanoids/aterm/cmd/aterm/config"
	"github.com/theparanoids/aterm/fancy"
)

var ErrCancelled = fmt.Errorf("Cancelled")
var ErrAlreadyExists = fmt.Errorf("Already Exists")

// ShowInvalidConfigurationMessage renders user-messaging when a validation error occurs.
// To actually validate the config, see config.ValidateConfig/ValidateLoadedConfig
func ShowInvalidConfigurationMessage(validationErr error) {
	if validationErr == nil {
		return
	}

	showAccessCorrection := ShowInvalidConfigMessageNoHelp(validationErr)

	println("These errors can be corrected by editing the configuration in the main menu, or " +
		"by editing the configuration file directly at " + fancy.WithBold(config.ATermConfigPath()) + ".")

	if showAccessCorrection {
		println("If you have lost your access key pair, you can generate a new pair from the ASHIRT servers.")
	}
}

func ShowInvalidConfigMessageNoHelp(validationErr error) bool {
	if validationErr == nil {
		return false
	}
	hasAccessIssue := false

	println("I detected problems with this configuration:")
	if errors.Is(validationErr, config.ErrAccessKeyNotSet) {
		println(" * Access Key has not been set")
		hasAccessIssue = true
	}
	if errors.Is(validationErr, config.ErrSecretKeyNotSet) {
		println(" * Secret Key has not been set")
		hasAccessIssue = true
	}
	if errors.Is(validationErr, config.ErrSecretKeyMalformed) {
		println(" * Secret Key is invalid")
		hasAccessIssue = true
	}
	if errors.Is(validationErr, config.ErrAPIURLUnparsable) {
		println(" * API URL is invalid")
	}
	println()

	return hasAccessIssue
}

// ShowConfigurationParsingErrorMessage renders user-messaging when the configuration file has issues
// parsing. It is assumed that these errors are NOT file-does-not-exist errors, as this should
// indicate a first run
func ShowConfigurationParsingErrorMessage(err error) {
	println("I had a problem parsing the configuration file:")
	println(" " + fancy.WithPizzazz(err.Error(), fancy.Red))
	println("Execution will continue, but some features may not work until the above issue is fixed")
}
