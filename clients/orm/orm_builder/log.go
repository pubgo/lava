package orm_builder

type logPrintf func(s string, i ...interface{})

func (t logPrintf) Printf(s string, i ...interface{}) { t(s, i...) }
