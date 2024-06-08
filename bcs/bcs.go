package bcs

// Marshaler is an interface for any type that can be serialized into BCS
type Marshaler interface {
	MarshalBCS(ser *Serializer)
}

// Unmarshaler is an interface for any type that can be deserialized from BCS
type Unmarshaler interface {
	UnmarshalBCS(des *Deserializer)
}

// Struct is an interface for an on-chain type.  It must be able to be both Marshaler and Unmarshaler for BCS
type Struct interface {
	Marshaler
	Unmarshaler
}
