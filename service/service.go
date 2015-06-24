package service

type Service interface {
	Name() string
	Run() error
}

type ServiceIniter interface {
	Init() error
}

type NewServiceFunc func(name string, args []string) Service

type ServiceType struct {
	Name string
	New  NewServiceFunc
	Desc string
	Use  bool
	Args []string
	Ins  Service
}
