package service

type Init interface {
	Init()
}

type Close interface {
	Close()
}

type Service interface {
	Start()
	Stop()
	Run()
}
