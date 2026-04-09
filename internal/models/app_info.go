package models

// DockerComposeConfig Docker Compose 配置结构
type DockerComposeConfig struct {
	Version  string             `json:"version" yaml:"version"`
	Services map[string]Service `json:"services" yaml:"services"`
	Networks map[string]Network `json:"networks,omitempty" yaml:"networks,omitempty"`
	Volumes  map[string]Volume  `json:"volumes,omitempty" yaml:"volumes,omitempty"`
}

// Service 服务结构
type Service struct {
	Image       string             `json:"image,omitempty" yaml:"image,omitempty"`
	Build       *BuildConfig       `json:"build,omitempty" yaml:"build,omitempty"`
	Ports       []string           `json:"ports,omitempty" yaml:"ports,omitempty"`
	Environment []string           `json:"environment,omitempty" yaml:"environment,omitempty"`
	EnvVars     map[string]string  `json:"env,omitempty" yaml:"env,omitempty"`
	Volumes     []string           `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	Networks    []string           `json:"networks,omitempty" yaml:"networks,omitempty"`
	DependsOn   []string           `json:"depends_on,omitempty" yaml:"depends_on,omitempty"`
	Healthcheck *HealthcheckConfig `json:"healthcheck,omitempty" yaml:"healthcheck,omitempty"`
}

// BuildConfig 构建配置结构
type BuildConfig struct {
	Context    string   `json:"context" yaml:"context"`
	Dockerfile string   `json:"dockerfile,omitempty" yaml:"dockerfile,omitempty"`
	Args       []string `json:"args,omitempty" yaml:"args,omitempty"`
}

// HealthcheckConfig 健康检查配置结构
type HealthcheckConfig struct {
	Test        []string `json:"test" yaml:"test"`
	Interval    string   `json:"interval,omitempty" yaml:"interval,omitempty"`
	Timeout     string   `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Retries     int      `json:"retries,omitempty" yaml:"retries,omitempty"`
	StartPeriod string   `json:"start_period,omitempty" yaml:"start_period,omitempty"`
}

// Network 网络结构
type Network struct {
	Driver   string      `json:"driver,omitempty" yaml:"driver,omitempty"`
	External bool        `json:"external,omitempty" yaml:"external,omitempty"`
	Name     string      `json:"name,omitempty" yaml:"name,omitempty"`
	IPAM     *IPAMConfig `json:"ipam,omitempty" yaml:"ipam,omitempty"`
}

// IPAMConfig IPAM 配置结构
type IPAMConfig struct {
	Driver string     `json:"driver,omitempty" yaml:"driver,omitempty"`
	Config []IPConfig `json:"config,omitempty" yaml:"config,omitempty"`
}

// IPConfig IP 配置结构
type IPConfig struct {
	Subnet  string `json:"subnet,omitempty" yaml:"subnet,omitempty"`
	Gateway string `json:"gateway,omitempty" yaml:"gateway,omitempty"`
}

// Volume 卷结构
type Volume struct {
	Driver     string            `json:"driver,omitempty" yaml:"driver,omitempty"`
	External   bool              `json:"external,omitempty" yaml:"external,omitempty"`
	Name       string            `json:"name,omitempty" yaml:"name,omitempty"`
	DriverOpts map[string]string `json:"driver_opts,omitempty" yaml:"driver_opts,omitempty"`
}
