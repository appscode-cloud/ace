/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package printer

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/nats-io/nats.go"
)

type natsMessage struct {
	ID     string `json:"id"`
	Step   string `json:"step,omitempty"`
	Status string `json:"status"`
}

const (
	stepStarted   = "Started"
	stepSucceeded = "Success"
	stepFailed    = "Failed"
)

func PrintNATSJobSteps(wg *sync.WaitGroup, nc *nats.Conn, responseID string, done <-chan os.Signal) error {
	defer wg.Done()

	subject := fmt.Sprintf("natjobs.resp.%s", responseID)
	steps := make(map[string]string)
	parentID := ""

	msgStream := make(chan *nats.Msg, 100)
	sub, err := nc.ChanSubscribe(subject, msgStream)
	if err != nil {
		return fmt.Errorf("failed to subscribe. Reason: %w", err)
	}

	stopListening := func(err error) error {
		unsubErr := sub.Unsubscribe()
		if unsubErr != nil {
			return fmt.Errorf("failed to unsbuscribe. Reason: %w", unsubErr)
		}
		close(msgStream)
		return err
	}
	for {
		select {
		case <-done:
			return stopListening(fmt.Errorf("command terminated by user"))
		case msg := <-msgStream:
			resp := natsMessage{}
			err := json.Unmarshal(msg.Data, &resp)
			if err != nil {
				return stopListening(fmt.Errorf("failed to parse message. Reason: %w", err))
			}
			if resp.Step != "" {
				if parentID == "" {
					parentID = resp.ID
				}
				steps[resp.ID] = resp.Step
			}
			if isStepStartedOrCompleted(resp.Status) {
				switch resp.Status {
				case stepSucceeded:
					color.Green(fmt.Sprintf("%s %s", strings.ToUpper(resp.Status), steps[resp.ID]))
				case stepFailed:
					color.Red(fmt.Sprintf("%s %s", strings.ToUpper(resp.Status), steps[resp.ID]))
				default:
					color.Blue(fmt.Sprintf("%s %s", strings.ToUpper(resp.Status), steps[resp.ID]))

				}
				if resp.ID == parentID && (resp.Status == "Success" || resp.Status == "Failed") {
					return stopListening(nil)
				}
			}
		}
	}
}

func isStepStartedOrCompleted(status string) bool {
	return status == stepStarted || status == stepSucceeded || status == stepFailed
}
