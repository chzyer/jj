package rpcprot

type Data struct {
	underlay interface{}
	buf      []byte
}

func NewData(d interface{}) *Data {
	return &Data{
		underlay: d,
	}
}

func NewRawData(buf []byte) *Data {
	return &Data{
		buf: buf,
	}
}
