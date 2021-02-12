// Copyright (c) J.Dreyer
// SPDX-License-Identifier: Apache-2.0

package http_forwarder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mainflux/mainflux/errors"
	"github.com/mainflux/mainflux/transformers/senml"
	"github.com/mainflux/mainflux/writers"
)

const pointName = "messages"

var errSaveMessage = errors.New("failed to send message to host")

var _ writers.MessageRepository = (*httpforwarderRepo)(nil)

type httpforwarderRepo struct {
	url string
}

type fields map[string]interface{}

// New returns new HTTP forwarder.
func New(url string) writers.MessageRepository {
	return &httpforwarderRepo{
		url: url,
	}
}

func (repo *httpforwarderRepo) Save(messages ...senml.Message) error {
	msgs := make(map[string][]*fields)
	for _, msg := range messages {
		t := repo.fullTopic(&msg)
		fields := repo.fieldsOf(&msg)
		msgs[t] = append(msgs[t], &fields)
	}

	for topic, msg := range msgs {
		data, err := json.Marshal(msg)
		if err != nil {
			return errors.Wrap(errSaveMessage, err)
		}

		url := fmt.Sprintf("%s/%s", strings.TrimRight(repo.url, "/"), topic)
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
		if err != nil {
			return errors.Wrap(errSaveMessage, err)
		}

		req.Header.Add("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return errors.Wrap(errSaveMessage, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return errors.Wrap(errSaveMessage, errors.New(resp.Status))
		}
	}

	return nil
}

func (repo *httpforwarderRepo) fullTopic(msg *senml.Message) string {
	t := fmt.Sprintf("channels.%s.%s", msg.Channel, msg.Subtopic)
	return strings.ReplaceAll(t, ".", "/")
}

func (repo *httpforwarderRepo) fieldsOf(msg *senml.Message) fields {
	ret := fields{
		"name":       msg.Name,
		"unit":       msg.Unit,
		"time": 	  msg.Time,
	}

	switch {
	case msg.Value != nil:
		ret["value"] = *msg.Value
	case msg.StringValue != nil:
		ret["stringValue"] = *msg.StringValue
	case msg.DataValue != nil:
		ret["dataValue"] = *msg.DataValue
	case msg.BoolValue != nil:
		ret["boolValue"] = *msg.BoolValue
	}

	if msg.Sum != nil {
		ret["sum"] = *msg.Sum
	}

	return ret
}
