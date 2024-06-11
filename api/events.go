package api

import "encoding/json"

// Event describes an on-chain event from Move. There are currently two types
// - V1 events, will have a [GUID] and Sequence number associated
// - V2 events will not have a [GUID] and Sequence number associated
type Event struct {
	Type           string // The Move struct type associated with the data
	Guid           *GUID  // TODO: Remove Guid and Sequence number altogether?
	SequenceNumber uint64
	Data           map[string]any // TODO: could be nice to type the events with generated code
}

func (o *Event) UnmarshalJSON(b []byte) error {
	type inner struct {
		Type           string         `json:"type"`
		Guid           *GUID          `json:"guid"`
		SequenceNumber U64            `json:"sequence_number"`
		Data           map[string]any `json:"data"`
	}
	data := &inner{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	o.Type = data.Type
	o.Guid = data.Guid
	o.SequenceNumber = data.SequenceNumber.toUint64()
	o.Data = data.Data
	return nil
}
