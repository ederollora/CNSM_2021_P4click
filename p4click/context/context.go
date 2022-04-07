package contextdefine

import (
	"github.com/google/uuid"
	"p4click/config"
)

type ControllerConfig struct
{
	ControllerName string    //Supports only ONOS for now
	ControllerType string
	ControllerVersion string //2.1, 2.2 etc.
	Formcluster bool
	ExternalIpAddress string
	InternalIpAddress string
	PortREST uint16          //8181 normally
	PortOpenFlow uint16      //6653 normally
	PortKarafSSH uint16      //8101 normally
	PortDebug uint16         //5005 normally
	PortIntraCluster uint16  //9876 normally
	PortOVSDB uint16         //6640 normally
	PortNetconf uint16       //830 normally
	Container_Id string
	Container_Name string
}

type SwitchPipeline struct {
	Name string
	Model string
	Target string
	IpAddress string
	Modules []string
}

type MachineConfig struct {
	Name              string
	IpAddress         string
	PortSSH           uint16
	User              string
	PasswordMode      bool
	PkaMode           bool
	PubKeyPath        string
	PrivKeyPath       string
	ControllersConfig []ControllerConfig
}

type DeploymentConfig struct {
	Name               string
	Id                 uuid.UUID
	ControllerMachines []MachineConfig
	Switches           []SwitchPipeline
	LocalTempDir       string
}

type P4ClickConfig struct {
	Repositories 	   []string
}

type MergedParser struct {

	UnionParseStates   map[string]config.ParsingState
	OrderedStates      map[int][]string

}




