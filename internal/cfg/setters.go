package cfg

func SetLoginGate(val bool) {
	lock.Lock()
	defer lock.Unlock()
	globalConfig.loginGate = val
}
