package subsystems

type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	CpuSet      string
}

type Subsystem interface {
	// Name 返回subsystem的名字
	Name() string
	// Set 设置某个cgroup在此subsystem中的资源限制
	Set(path string, res *ResourceConfig) error
	// Apply 将进程添加到某个cgroup中
	Apply(path string, pid int) error
	// Remove 移除某个cgroup
	Remove(path string) error
}

var (
	SubsystemIns = []Subsystem{
		&MemorySubsystem{},
		&CpuSubsystem{},
	}
)
