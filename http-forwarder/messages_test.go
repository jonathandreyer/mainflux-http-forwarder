// Copyright (c) J.Dreyer
// SPDX-License-Identifier: Apache-2.0

package http_forwarder_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	writer "github.com/jonathandreyer/mainflux-httpforwarder/http-forwarder"
	"github.com/mainflux/mainflux/errors"
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/transformers/senml"
	"github.com/stretchr/testify/assert"
)

const valueFields = 5

var (
	testLog, _  = log.New(os.Stdout, log.Info.String())
	streamsSize = 250
	host     = "http://localhost:9000"
	token    = ""
	subtopic = "messages"
)

var (
	v       float64 = 5
	stringV         = "value"
	boolV           = true
	dataV           = "base64"
	sum     float64 = 42
)

func TestForwarder(t *testing.T) {
	repo := writer.New(host, token)

	cases := []struct {
		desc         string
		msgsNum      int
		expectedSize int
	}{
		{
			desc:         "transfer a single message",
			msgsNum:      1,
			expectedSize: 1,
		},
		{
			desc:         "save a batch of messages",
			msgsNum:      streamsSize,
			expectedSize: streamsSize,
		},
	}

	for _, tc := range cases {
		now := time.Now().UnixNano()
		msg := senml.Message{
			Channel:    "45",
			Publisher:  "2580",
			Protocol:   "http",
			Name:       "test name",
			Unit:       "km",
			UpdateTime: 5456565466,
		}
		var msgs []senml.Message

		for i := 0; i < tc.msgsNum; i++ {
			// Mix possible values as well as value sum.
			count := i % valueFields
			switch count {
			case 0:
				msg.Subtopic = subtopic
				msg.Value = &v
			case 1:
				msg.BoolValue = &boolV
			case 2:
				msg.StringValue = &stringV
			case 3:
				msg.DataValue = &dataV
			case 4:
				msg.Sum = &sum
			}

			msg.Time = float64(now)/float64(1e9) + float64(i)
			msgs = append(msgs, msg)
		}

		err := repo.Save(msgs...)
		assert.Nil(t, err, fmt.Sprintf("Send messages operation expected to succeed: %s.\\n", err))

		for i, _ := range msgs {
			msgs[i].Subtopic = "topic"
		}
		err = repo.Save(msgs...)
		errExpected := errors.New("failed to send message to host")
		assert.Truef(t, errors.Contains(err, errExpected), "Error should be: %v, got: %v", errExpected, err)
	}
}
