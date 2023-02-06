package tool

type Tool interface {
	Init()
	Name() (name string)
	Run(args []string) (exitCode int, err error, stdout string, stdin string)
	IsInstalled() bool
	HowToInstall()
}
