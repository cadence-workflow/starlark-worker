package encoded

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
