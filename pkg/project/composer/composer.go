package composer

import "ttp.sh/frost/pkg/project"

func init() {
	project.AddHandler(DefaultHandler)
}

var DefaultHandler = new(handler)

const lockFile = "composer.lock"

type handler struct{}

func (handler) Name() string {
	return "Composer"
}

func (h *handler) GetLockFile() (lock *Lock) {
	f := os.Open(lockFile)
	json.NewDecoder(f).Decode(&lock)
	return 
}

func (h *handler) Execute() {
	lock := h.GetLockFile()
}


