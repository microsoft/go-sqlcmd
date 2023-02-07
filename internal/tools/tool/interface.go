package tool

type Tool interface {
	Init()
	Name() (name string)
	Run(args []string) (exitCode int, err error)
	IsInstalled() bool
	HowToInstall() string
}
