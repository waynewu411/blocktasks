package jsonrpc

type Request struct {
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	Id      int64  `json:"id"`
	JsonRpc string `json:"jsonrpc"`
}

type Response[T any] struct {
	JsonRpc string `json:"jsonrpc"`
	Id      int64  `json:"id"`
	Result  T      `json:"result"`
}
