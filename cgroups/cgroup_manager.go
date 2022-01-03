package cgroups

import (
	"fmt"

	"github.com/ForeverSRC/MyDocker/cgroups/subsystems"
	log "github.com/sirupsen/logrus"
)

type CgroupManager struct {
	Path     string
	Resource *subsystems.ResourceConfig
}

const CgroupPathFormat = "my-docker-cgroup/%s"

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

func (m *CgroupManager) Apply(pid int) error {
	var err error
	for _, subSysIns := range subsystems.SubsystemIns {
		err = subSysIns.Apply(m.Path, pid)
		if err != nil {
			return fmt.Errorf("apply sub system %s error: %v", subSysIns.Name(), err)
		}

	}

	return nil
}

func (m *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	var err error
	for _, subSysIns := range subsystems.SubsystemIns {
		err = subSysIns.Set(m.Path, res)
		if err != nil {
			return fmt.Errorf("set sub system %s error: %v", subSysIns.Name(), err)
		}

	}
	return nil
}

func (m *CgroupManager) Destroy() error {
	for _, subSysIns := range subsystems.SubsystemIns {
		if err := subSysIns.Remove(m.Path); err != nil {
			log.Warnf("remove cgroup fail %v", err)
		}

	}
	return nil
}
