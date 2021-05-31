package consts

var strBank StrBank

type Index struct {
	offset int32
	length int32
}

func (i Index) String() string {
	return strBank.Get(i.offset, i.length)
}

func StrOf(val string) Index { return strBank.StrOf(val) }

type StrBank struct {
	data string
}

func (s *StrBank) Get(index, offset int32) string {
	return s.data[index : index+offset]
}

func (s *StrBank) StrOf(str string) Index {
	return Index{int32(len(s.data)), int32(len(str))}
}
