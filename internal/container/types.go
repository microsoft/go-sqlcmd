package container

type RunOptions struct {
	Network         string
	Env             []string
	PortInternal    int
	Port            int
	Name            string
	Hostname        string
	Architecture    string
	Os              string
	Command         []string
	UnitTestFailure bool
}
