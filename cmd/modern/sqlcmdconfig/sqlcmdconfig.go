package sqlcmdconfig

type Sqlcmdconfig struct {
	Version   string     `mapstructure:"version"`
	Databases []Database `mapstructure:"databases"`
	AddOns    []AddOn    `mapstructure:"addons"`
}

type Database struct {
	DatabaseDetails `mapstructure:"database" yaml:"database"`
}

type DatabaseDetails struct {
	Name string `mapstructure:"name" yaml:"name"`
	Use  []Use  `mapstructure:"use"`
}

type Use struct {
	Uri       string `mapstructure:"uri"`
	Mechanism string `mapstructure:"mechanism"`
}

type AddOn struct {
	AddOnDetails `mapstructure:"addon" yaml:"addon"`
}

type AddOnDetails struct {
	Type string `mapstructure:"type"`
	Use  []Use  `mapstructure:"use"`
}
