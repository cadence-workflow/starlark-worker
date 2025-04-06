package encoded

type DataConvertor interface {
	// ToData implements conversion of a list of values.
	ToData(value ...interface{}) ([]byte, error)
	// FromData implements conversion of an array of values of different types.
	// Useful for deserializing arguments of function invocations.
	FromData(input []byte, valuePtr ...interface{}) error
}

type Value interface {
	// HasValue return whether there is value encoded.
	HasValue() bool
	// Get extract the encoded value into strong typed value pointer.
	Get(valuePtr interface{}) error
}

type Values interface {
	// HasValues return whether there are values encoded.
	HasValues() bool
	// Get extract the encoded values into strong typed value pointers.
	Get(valuePtr ...interface{}) error
}
