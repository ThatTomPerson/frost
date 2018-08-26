package project

type Module interface {
	Install(*Project)
	GetHandler() PackageHandler
}
