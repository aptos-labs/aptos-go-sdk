package api

import "encoding/json"

//region Event

// Event describes an on-chain event from Move. There are currently two types:
//
// - Handle events, will have a [GUID] and Sequence number associated
//
//	{
//		"type": "0x1::coin::WithdrawEvent",
//		"guid": {
//			"id": {
//				"addr": "0x810026ca8291dd88b5b30a1d3ca2edd683d33d06c4a7f7c451d96f6d47bc5e8b",
//				"creation_num": "3"
//			}
//		},
//		"sequence_number": "0",
//		"data": {
//			"amount": "1000"
//		}
//	}
//
// - Module events will not have a [GUID] and Sequence number associated
//
//	{
//		"type": "0x1::fungible_asset::Withdraw",
//		"guid": {
//			"id": {
//				"addr": "0x0",
//				"creation_num": "0"
//			}
//		},
//		"sequence_number": "0",
//		"data": {
//			"store": "0x1234123412341234123412341234123412341234123412341234123412341234",
//			"amount": "1000"
//		}
//	}
type Event struct {
	Type           string         // Type is the fully qualified name of the event e.g. 0x1::coin::WithdrawEvent
	Guid           *GUID          // GUID is the unique identifier of the event, only present in V1 events
	SequenceNumber uint64         // SequenceNumber is the sequence number of the event, only present in V1 events
	Data           map[string]any // Data is the event data, a map of field name to value
}

//region Event JSON

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

//endregion
//endregion
