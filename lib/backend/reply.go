package backend

import (
	"encoding/json"
)

type (
	Reply struct {
		Verb   string          `json:"verb"`
		Object json.RawMessage `json:"object"`
		Status string          `json:"status"`
		Data   json.RawMessage `json:"data"`
	}

	NginxReply struct {
		Host    string `json:"host,omitempty"`
		Port    string `json:"port,omitempty"`
		WaitFor int    `json:"waitfor,omitempty"`
	}
)

// KO …
func (reply *Reply) KO(data interface{}) error {
	reply.Status = "KO"
	if data == nil {
		reply.Data = json.RawMessage("")
		return nil
	}

	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}
	reply.Data = buf
	return nil
}

// OK …
func (reply *Reply) OK(data interface{}) error {
	reply.Status = "OK"
	if data == nil {
		reply.Data = json.RawMessage("")
		return nil
	}

	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}
	reply.Data = buf
	return nil
}

// Postfix decoding
func (reply *Reply) Postfix() (data string, err error) {
	err = json.Unmarshal(reply.Data, &data)
	return
}

// Nginx decoding
func (reply *Reply) Nginx() (data NginxReply, err error) {
	err = json.Unmarshal(reply.Data, &data)
	return
}

// Dovecot decoding
func (reply *Reply) Dovecot() (data DovecotQuery, err error) {
	err = json.Unmarshal(reply.Data, &data)
	return
}
