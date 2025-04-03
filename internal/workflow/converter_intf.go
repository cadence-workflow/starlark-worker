package workflow

// DataConverterInterface defines the methods that a DataConverter should implement.
type DataConverterInterface interface {
	ToData(values ...any) ([]byte, error)
	FromData(data []byte, to ...any) error
}
