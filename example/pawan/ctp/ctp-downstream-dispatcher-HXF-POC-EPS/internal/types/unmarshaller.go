package types

import (
	"encoding/json"
	"net/http"
)

type headers Header

// UnmarshalJSON for unmarshalling Header struct expecting both struct and http.Header
func (h *Header) UnmarshalJSON(b []byte) (err error) {
	s, t := headers{}, http.Header{}

	if err = json.Unmarshal(b, &s); err == nil {
		*h = Header(s)
		return
	}

	if err = json.Unmarshal(b, &t); err == nil {
		h.ACMHeaders = t
	}
	return
}
