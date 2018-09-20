package project

type Package interface {
}

type Cache interface {
	Get(Module) Package
	Set(Module, Package) bool
}

type Module interface {
	Download() Package
	Install(*Project, Package)
	PostInstall(*Project)
	GetHandler() PackageHandler
	GetCache() Cache
}

// M Wrapps a pacakge manager
type M interface {
	// Execute a PackageManager
	Execute(*Project) error
}
