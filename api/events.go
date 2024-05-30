package api

// Event describes an on-chain event from Move. There are currently two types
// V1 events, will have a GUID and Sequence number associated
// V2 events will not have a GUID and Sequence number associated
type Event struct {
	Type           string // The Move struct type associated with the data
	Guid           *GUID  // TODO: Remove Guid and Sequence number altogether?
	SequenceNumber uint64
	Data           map[string]any // TODO: could be nice to type the events with generated code
}

func (o *Event) UnmarshalJSONFromMap(data map[string]any) (err error) {
	o.Guid, err = toGuid(data, "guid")
	if err != nil {
		return err
	}
	o.SequenceNumber, err = toUint64(data, "sequence_number")
	if err != nil {
		return err
	}
	o.Data, err = toMap(data, "data")
	return err
}
