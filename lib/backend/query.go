package backend

import (
	"encoding/json"
)

type (
	Query struct {
		Verb   string          `json:"verb"`
		Object json.RawMessage `json:"object"`
	}

	NginxQuery struct {
		User    string `json:"user"`
		Pass    string `json:"pass"`
		Proto   string `json:"protocol,omitempty"`
		Attempt string `json:"attempt,omitempty"`
		IpCli   string `json:"client,omitempty"`
	}

	DovecotQuery struct {
		Namespace string `json:"context"`
		Object    string `json:"object"`
	}
)

func NewQuery(verb string, object interface{}) (q *Query, err error) {
	*q = Query{
		Verb: verb,
	}

	q.Object, err = json.Marshal(object)
	if err != nil {
		q = nil
		return
	}
	return
}

// Postfix decoding
func (request *Query) Postfix() (data string, err error) {
	err = json.Unmarshal(request.Object, &data)
	return
}

// Nginx decoding
func (request *Query) Nginx() (data NginxQuery, err error) {
	err = json.Unmarshal(request.Object, &data)
	return
}

// Dovecot decoding
func (request *Query) Dovecot() (data DovecotQuery, err error) {
	err = json.Unmarshal(request.Object, &data)
	return
}

// MakeReply forge a reply struct based on the query
func (request *Query) MakeReply() (reply *Reply) {
	*reply = Reply{
		Verb:   request.Verb,
		Object: request.Object,
	}
	return
}
