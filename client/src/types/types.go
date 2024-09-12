package types

import (
	"fmt"
	"time"
)



type Metric struct {
	Duration time.Duration
}

type Config struct {
	SutAddresses []string `yaml:"sut_addresses"`
	LocalAddress string `yaml:"local_address"`
	SutDockerHostTemplate string `yaml:"sut_docker_host_template"`
	PortWrite    int    `yaml:"sut_port_write"`
	PortRead     int    `yaml:"sut_port_read"`
	OutputFolder string `yaml:"output_folder"`
	InputData    string `yaml:"input_data"`
	TaskManagerMemory string `yaml:"taskmanager_memory"`

	PrinkDockerImage string `yaml:"prink_docker_image"`
}

type Experiment struct {
		// k = [5,10,20,40,80]
		// delta = [1250,5000,20000,80000]
		// l diversity = [0,2,4,8] (wenn null bedeutet, dass l diversity nicht beachtet wird)
		// beta (active clusters). Damit das keinen Einfluss hat, so hoch setzen wie Daten: beta= 321728
		// mu= 100 (wie im original paper)

		K int
		Delta int
		L int
		Beta int
		Zeta int
		Mu int

		LocalHost string
		SutHost string
		SutPortWrite int
		SutPortRead int
}

func (e Experiment) String() string {
	return fmt.Sprintf("Experiment: k=%d, delta=%d, l=%d, beta=%d, zeta=%d, mu=%d, local_host=%s, sut_host=%s, sut_port_write=%d, sut_port_read=%d", e.K, e.Delta, e.L, e.Beta, e.Zeta, e.Mu, e.LocalHost, e.SutHost, e.SutPortWrite, e.SutPortRead)
}

func (e Experiment) ToFileName () string {
	return fmt.Sprintf("k%d_delta%d_l%d_beta%d_zeta%d_mu%d", e.K, e.Delta, e.L, e.Beta, e.Zeta, e.Mu)
}

func (e Experiment) ToArgs() []string {
	return []string{
		"--k", fmt.Sprintf("%d", e.K),
		"--delta", fmt.Sprintf("%d", e.Delta),
		"--l", fmt.Sprintf("%d", e.L),
		"--beta", fmt.Sprintf("%d", e.Beta),
		"--zeta", fmt.Sprintf("%d", e.Zeta),
		"--mu", fmt.Sprintf("%d", e.Mu),
		"--sut_host", e.LocalHost, // in the container, the SUT host is the local host
		"--sut_port_write", fmt.Sprintf("%d", e.SutPortWrite),
		"--sut_port_read", fmt.Sprintf("%d", e.SutPortRead),
	}
}
