package container

type RunOptions struct {
	Env             []string
	Port            int
	Name            string
	Hostname        string
	Architecture    string
	Os              string
	Command         []string
	UnitTestFailure bool
}

type ExecOptions struct {
	User string
	Env  []string
}
