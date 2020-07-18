// Copyright 2020, Verizon Media
// Licensed under the terms of the MIT. See LICENSE file in project root for terms.

package network

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"

	"github.com/pkg/errors"
	"github.com/theparanoids/ashirt-server/backend/dtos"
)

// GetTags retrieves a list of all tags from the server for the given operation slug
func GetTags(operationSlug string) ([]dtos.Tag, error) {
	var tags []dtos.Tag

	resp, err := makeJSONRequest("GET", apiURL+"/operations/"+operationSlug+"/tags", http.NoBody)
	if err != nil {
		return tags, errors.Wrap(err, errCannotConnectMsg)
	}

	if err = evaluateResponseStatusCode(resp.StatusCode); err != nil {
		return tags, err
	}

	err = readResponseBody(&tags, resp.Body)

	return tags, err
}

// CreateTag generates a new tag on the backend. If successful, the new tag with tag ID will
// be returned
func CreateTag(operationSlug, name, colorName string) (*dtos.Tag, error) {
	type TagInput struct {
		Name      string `json:"name"`
		ColorName string `json:"colorName"`
	}
	data := TagInput{Name: name, ColorName: colorName}
	content, err := json.Marshal(data)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to create tag")
	}

	resp, err := makeJSONRequest("POST", apiURL+"/operations/"+operationSlug+"/tags", bytes.NewReader(content))

	if err != nil {
		return nil, errors.Wrap(err, errCannotConnectMsg)
	}

	if err = evaluateResponseStatusCode(resp.StatusCode); err != nil {
		return nil, err
	}

	var tag dtos.Tag
	err = readResponseBody(&tag, resp.Body)

	return &tag, err
}

func RandomTagColor() string {
	allTagColors := []string{
		"blue",
		"yellow",
		"green",
		"indigo",
		"orange",
		"pink",
		"red",
		"teal",
		"vermilion",
		"violet",
		"lightBlue",
		"lightYellow",
		"lightGreen",
		"lightIndigo",
		"lightOrange",
		"lightPink",
		"lightRed",
		"lightTeal",
		"lightVermilion",
		"lightViolet",
	}
	return allTagColors[rand.Intn(len(allTagColors))]
}