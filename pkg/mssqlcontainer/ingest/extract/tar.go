package extract

type tar struct {
}

func (e tar) FileTypes() []string {
	return []string{"tar"}
}

func (e tar) IsInstalled(containerId string) bool {
	return true
}

func (e tar) Extract(srcFile string, destFolder string) (string, string) {
	return "", ""
}

func (e tar) Install() {
}
