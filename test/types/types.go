package types

type EncodedValue interface {
	// HasValue return whether there is value encoded.
	HasValue() bool
	// Get extract the encoded value into strong typed value pointer.
	//
	// Note, values should not be reused for extraction here because merging on
	// top of existing values may result in unexpected behavior similar to
	// json.Unmarshal.
	Get(valuePtr interface{}) error
}

type StarTestActivitySuite interface {
	RegisterActivity(a interface{})
	ExecuteActivity(a interface{}, opts ...interface{}) (EncodedValue, error)
}
