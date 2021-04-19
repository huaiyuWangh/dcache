package byteview

type ByteView struct {
	B []byte
}

func (v ByteView) Len() int {
	return len(v.B)
}

func (v ByteView) String() string {
	return string(v.B)
}

func (v *ByteView) ByteSlice() []byte {
	dst := make([]byte, len(v.B))
	copy(dst, v.B)
	return dst
}