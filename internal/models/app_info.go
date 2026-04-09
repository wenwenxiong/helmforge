package models

// DockerComposeConfig Docker Compose 配置结构
type DockerComposeConfig struct {
	Version  string             `json:"version"`
	Services map[string]Service `json:"services"`
	Networks map[string]Network `json:"networks,omitempty"`
	Volumes  map[string]Volume  `json:"volumes,omitempty"`
}

// Service 服务结构
type Service struct {
	Image       string             `json:"image,omitempty"`
	Build       *BuildConfig       `json:"build,omitempty"`
	Ports       []string           `json:"ports,omitempty"`
	Environment []string           `json:"environment,omitempty"`
	EnvVars     map[string]string  `json:"env,omitempty"`
	Volumes     []string           `json:"volumes,omitempty"`
	Networks    []string           `json:"networks,omitempty"`
	DependsOn   []string           `json:"depends_on,omitempty"`
	Healthcheck *HealthcheckConfig `json:"healthcheck,omitempty"`
}

// BuildConfig 构建配置结构
type BuildConfig struct {
	Context    string   `json:"context"`
	Dockerfile string   `json:"dockerfile,omitempty"`
	Args       []string `json:"args,omitempty"`
}

// HealthcheckConfig 健康检查配置结构
type HealthcheckConfig struct {
	Test        []string `json:"test"`
	Interval    string   `json:"interval,omitempty"`
	Timeout     string   `json:"timeout,omitempty"`
	Retries     int      `json:"retries,omitempty"`
	StartPeriod string   `json:"start_period,omitempty"`
}

// Network 网络结构
type Network struct {
	Driver   string      `json:"driver,omitempty"`
	External bool        `json:"external,omitempty"`
	Name     string      `json:"name,omitempty"`
	IPAM     *IPAMConfig `json:"ipam,omitempty"`
}

// IPAMConfig IPAM 配置结构
type IPAMConfig struct {
	Driver string     `json:"driver,omitempty"`
	Config []IPConfig `json:"config,omitempty"`
}

// IPConfig IP 配置结构
type IPConfig struct {
	Subnet  string `json:"subnet,omitempty"`
	Gateway string `json:"gateway,omitempty"`
}

// Volume 卷结构
type Volume struct {
	Driver     string            `json:"driver,omitempty"`
	External   bool              `json:"external,omitempty"`
	Name       string            `json:"name,omitempty"`
	DriverOpts map[string]string `json:"driver_opts,omitempty"`
}
