package consts

var strBox StrBank

type Index struct {
	offset int32
	length int32
}

func (i Index) String() string {
	return strBox.Get(i.offset, i.length)
}

func IndexOf(val string) Index { return strBox.IndexOf(val) }

type StrBank struct {
	data string
}

func (s *StrBank) Get(index, offset int32) string {
	return s.data[index : index+offset]
}

func (s *StrBank) IndexOf(str string) Index {
	s.data += str
	return Index{int32(len(s.data)), int32(len(str))}
}
