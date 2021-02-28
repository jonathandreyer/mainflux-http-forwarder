// Copyright (c) J.Dreyer
// SPDX-License-Identifier: Apache-2.0

package http_forwarder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/mainflux/mainflux/errors"
	"github.com/mainflux/mainflux/transformers/senml"
	"github.com/mainflux/mainflux/writers"
)

var errSaveMessage = errors.New("failed to send message to host")

var _ writers.MessageRepository = (*httpforwarderRepo)(nil)

type httpforwarderRepo struct {
	remoteUrl   string
	remoteToken string
}

type Address struct {
	FullTopic string
	Published string
	Protocol  string
}
type fields map[string]interface{}

// New returns new HTTP forwarder.
func New(url string, token string) writers.MessageRepository {
	return &httpforwarderRepo{
		remoteUrl: url,
		remoteToken: token,
	}
}

func (repo *httpforwarderRepo) Save(messages ...senml.Message) error {
	messagesSorted := repo.sortMessages(messages)

	messagesFormatted := make(map[Address][]*fields)
	for addr, msgs := range messagesSorted {
		var basefields = repo.extractBaseFields(msgs)

		for _, msg := range msgs {
			var m = fields{}

			// Add base fields when map is empty
			if _, ok := messagesFormatted[addr]; !ok {
				m = basefields
			}

			m = repo.appendFields(&msg, basefields, m)
			messagesFormatted[addr] = append(messagesFormatted[addr], &m)
		}
	}

	for address, msg := range messagesFormatted {
		data, err := json.Marshal(msg)
		if err != nil {
			return errors.Wrap(errSaveMessage, err)
		}

		url := fmt.Sprintf("%s/%s", strings.TrimRight(repo.remoteUrl, "/"), address.FullTopic)
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
		if err != nil {
			return errors.Wrap(errSaveMessage, err)
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("MF-Publisher", address.Published)
		if repo.remoteToken != "" {
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", repo.remoteToken) )
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return errors.Wrap(errSaveMessage, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusAccepted {
			return errors.Wrap(errSaveMessage, errors.New(resp.Status))
		}
	}

	return nil
}

func (repo *httpforwarderRepo) sortMessages(messages []senml.Message) map[Address][]senml.Message {
	sortedMessages := make(map[Address][]senml.Message)

	for _, msg := range messages {
		a := Address{
			FullTopic: strings.ReplaceAll(fmt.Sprintf("channels.%s.%s", msg.Channel, msg.Subtopic), ".", "/"),
			Published: msg.Publisher,
			Protocol: msg.Protocol,
		}
 		sortedMessages[a] = append(sortedMessages[a], msg)
	}

	return sortedMessages
}

func (repo *httpforwarderRepo) extractBaseFields(messages []senml.Message) fields {
	var f = fields{}

	if len(messages) < 2 {
		return f
	}

	// Create list of names, times & units1
	var names, units []string
	var times []float64
	for _, msg := range messages {
		names = append(names, msg.Name)
		times = append(times, msg.Time)
		units = append(units, msg.Unit)
	}

	// Determine Base name with "Greatest common name"
	name := strings.Split(names[0], ":")
	nbCommon := len(name)
	for _, n := range names[1:] {
		elements := strings.Split(n, ":")
		// Reduce parsing space (message name is shorter)
		if len(elements) < len(name) {
			nbCommon = len(elements)
		}
		for i, e := range elements {
			// Reduce parsing space (common base name is shorter)
			if i > (nbCommon-1) {
				break
			}
			// Break loop when the indexed name differs from the base name
			if name[i] != e {
				nbCommon = i
				break
			}
		}
	}

	// If an common name has been found, create this base name
	if nbCommon > 0 {
		f["bn"] = strings.Join(name[:nbCommon], ":")
		// If the base name is a substring of names, add separator at the end of base name
		if nbCommon < len(name) {
			f["bn"] = fmt.Sprintf("%s:", f["bn"])
		}
	}

	// Base time
	sort.Float64s(times)
	f["bt"] = times[0]

	// Base unit
	sort.Strings(units)
	if units[0] == units[len(units)-1] && units[0] != ""{
		f["bu"] = units[0]
	}

	// Version
	f["bver"] = 5

	return f
}

func (repo *httpforwarderRepo) appendFields(msg *senml.Message, baseField fields, field fields) fields {
	// Remove base name of name when it is available
	if prefix, ok := baseField["bn"].(string); ok {
		if value := strings.TrimPrefix(msg.Name, prefix); value != "" {
			field["n"] = value
		}
	} else {
		field["n"] = msg.Name
	}

	// Add time when the base time is missing or add delta time regarding the base time
	if value, ok := baseField["bt"].(float64); ok {
		if value != msg.Time {
			field["t"] = msg.Time - value
		}
	} else {
		field["t"] = msg.Time
	}

	switch {
	case msg.Value != nil:
		field["v"] = *msg.Value
	case msg.StringValue != nil:
		field["vs"] = *msg.StringValue
	case msg.DataValue != nil:
		field["vd"] = *msg.DataValue
	case msg.BoolValue != nil:
		field["vb"] = *msg.BoolValue
	}

	// Add unit when the base unit is missing or when it is another unit
	if value, ok := baseField["bu"].(string); ok {
		if value != msg.Unit {
			field["u"] = msg.Unit
		}
	} else if msg.Unit != "" {
		field["u"] = msg.Unit
	}

	if msg.UpdateTime != 0 {
		field["ut"] = msg.UpdateTime
	}

	if msg.Sum != nil {
		field["s"] = *msg.Sum
	}

	return field
}
