package cad

type contextKey int

const (
	contextKeyHeaders contextKey = iota
)

func GetContextHeaders(ctx interface{ Value(key any) any }) map[string][]byte {
	return ctx.Value(contextKeyHeaders).(map[string][]byte)
}
