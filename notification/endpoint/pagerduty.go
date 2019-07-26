package endpoint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/influxdata/influxdb"
)

var _ influxdb.NotificationEndpoint = &PagerDuty{}

const routingKeySuffix = "-routing-key"

// PagerDuty is the notification endpoint config of pagerduty.
type PagerDuty struct {
	Base
	// Path is the PagerDuty API URL, should not need to be changed.
	URL string `json:"url"`
	// RoutingKey is a version 4 UUID expressed as a 32-digit hexadecimal number.
	// This is the Integration Key for an integration on any given service.
	RoutingKey influxdb.SecretField `json:"routing-key"`
}

// BackfillSecretKeys fill back fill the secret field key during the unmarshalling
// if value of that secret field is not nil.
func (s *PagerDuty) BackfillSecretKeys() {
	if s.RoutingKey.Key == "" && s.RoutingKey.Value != nil {
		s.RoutingKey.Key = s.ID.String() + routingKeySuffix
	}
}

// SecretFields return available secret fields.
func (s PagerDuty) SecretFields() []influxdb.SecretField {
	return []influxdb.SecretField{
		s.RoutingKey,
	}
}

// Valid returns error if some configuration is invalid
func (s PagerDuty) Valid() error {
	if err := s.Base.valid(); err != nil {
		return err
	}
	if s.URL == "" {
		return &influxdb.Error{
			Code: influxdb.EInvalid,
			Msg:  "pagerduty endpoint URL is empty",
		}
	}
	if _, err := url.Parse(s.URL); err != nil {
		return &influxdb.Error{
			Code: influxdb.EInvalid,
			Msg:  fmt.Sprintf("pagerduty endpoint URL is invalid: %s", err.Error()),
		}
	}
	if s.RoutingKey.Key != s.ID.String()+routingKeySuffix {
		return &influxdb.Error{
			Code: influxdb.EInvalid,
			Msg:  "pagerduty routing key is invalid",
		}
	}
	return nil
}

type pagerdutyAlias PagerDuty

// MarshalJSON implement json.Marshaler interface.
func (s PagerDuty) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		struct {
			pagerdutyAlias
			Type string `json:"type"`
		}{
			pagerdutyAlias: pagerdutyAlias(s),
			Type:           s.Type(),
		})
}

// Type returns the type.
func (s PagerDuty) Type() string {
	return PagerDutyType
}

// ParseResponse will parse the http response from pagerduty.
func (s PagerDuty) ParseResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return &influxdb.Error{
				Msg: "error parsing error body",
			}
		}
		type response struct {
			Status  string   `json:"status"`
			Message string   `json:"message"`
			Errors  []string `json:"errors"`
		}
		r := &response{Message: fmt.Sprintf("failed to understand PagerDuty2 response. code: %d content: %s", resp.StatusCode, string(body))}
		b := bytes.NewReader(body)
		dec := json.NewDecoder(b)
		dec.Decode(r)
		return &influxdb.Error{
			Msg: fmt.Sprintf("Status: %s, Message: %s Errors: %v", r.Status, r.Message, r.Errors),
		}
	}
	return nil
}
