//package p4click

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	prompt "github.com/c-bata/go-prompt"
	"p4click/arguments"
	. "p4click/context"
	pipeline "p4click/pipeline"
	. "p4click/ssh"
	. "p4click/types"
)


var p4ClickConfig *P4ClickConfig
var currentDeployment *DeploymentConfig
var currentMachine *MachineConfig
var currentSwitch *SwitchPipeline
var currentController *ControllerConfig
var allDeployments []DeploymentConfig


var LivePrefixState struct {
	LivePrefix string
	IsEnable   bool
}

var context []ContextType //Represent in which point of config we are right now. (p4click -> projectname -> controller ...etc

var p4ClickDir string


func Complete(d prompt.Document) []prompt.Suggest {
	args := strings.Split(d.TextBeforeCursor(), " ")
	return arguments.ArgumentsCompleter(args, context, currentDeployment, allDeployments,
		currentMachine, currentController, currentSwitch)
}

func Executor(in string) {
	in = strings.TrimSpace(in)
	if in == "" {
		return
	} else if in == "quit" {
		fmt.Println("Bye!")
		os.Exit(0)
		return
	}
	blocks := strings.Split(in, " ")

	command := blocks[0]
	switch command {

		case "default":
			item := strings.TrimSpace(blocks[1])
			if (item == "config") {

				p4ClickConfig.Repositories = append(p4ClickConfig.Repositories, "modules.p4.click")
				//p4ClickConfig.Repositories = append(p4ClickConfig.Repositories, "/home/p4/go/src/p4click/repository")

				var d DeploymentConfig

				currentDeployment = &d
				currentDeployment.Id = uuid.New()
				currentDeployment.Name = "d1"
				//currentDeployment.LocalTempDir = createRandomTempDirectory(p4ClickDir, currentDeployment.Id.String())

				includeDir := filepath.Join(p4ClickDir, "dep2")
				os.MkdirAll(includeDir, os.ModePerm)
				currentDeployment.LocalTempDir = includeDir

				fmt.Println(fmt.Sprintf("%s", includeDir))

				allDeployments = append(allDeployments, d)

				projectType := ContextType{"d1", DEPLOYMENT}
				context = append(context, projectType)


				for i := 1; i < 5; i++ {
					var switchPipeline SwitchPipeline

					switchPipeline.Name = fmt.Sprintf("%s%d", "s", i) // s1, s2, ...
					switchPipeline.Model = "v1model"
					switchPipeline.Target = "bmv2"
					switchPipeline.Modules = append(switchPipeline.Modules, "l2forwarding")
					switchPipeline.Modules = append(switchPipeline.Modules, "l3forwarding")
					switchPipeline.Modules = append(switchPipeline.Modules, "aclhandler")

					currentDeployment.Switches = append(currentDeployment.Switches, switchPipeline)
				}

				currentSwitch = &currentDeployment.Switches[0]

				switchType := ContextType{"s1", SWITCH}
				context = append(context, switchType)

				updatePrefix()
			}

		case "create":
			item := strings.TrimSpace(blocks[1])
			if (item == "deployment") {

				name := strings.TrimSpace(blocks[2])
				if(name == ""){
					log.Fatal("The deployment needs a name")
					break
				}

				var d DeploymentConfig

				currentDeployment = &d
				currentDeployment.Id = uuid.New()
				currentDeployment.Name = name
				currentDeployment.LocalTempDir = createRandomTempDirectory(p4ClickDir, currentDeployment.Id.String())

				allDeployments = append(allDeployments, d)

			} else
			if item == "switch" {
				if currentDeployment == nil{
					log.Fatal("Before creating a switch one needs to create and use a deployment")
				}

				var switchPipeline SwitchPipeline

				name := strings.TrimSpace(blocks[2])
				switchPipeline.Name = name
				currentDeployment.Switches = append(currentDeployment.Switches, switchPipeline)
			}else
			if item == "machine" {
				name := strings.TrimSpace(blocks[2])
				if(name == ""){
					log.Fatal("The deployment needs a name")
					break
				}

				var mc MachineConfig

				currentMachine = &mc
				currentMachine.Name = name
				currentDeployment.ControllerMachines = append(currentDeployment.ControllerMachines, mc)
			}else
			if item == "controller" {

				if len(blocks) < 5 {
					fmt.Println("Error. Command has to be: create controller <controller_type> <version> <name>")
					return
				}

				if context[len(context) - 1].Ctype != MACHINE{
					fmt.Println("Error. The context needs to point to a machine.")
					return
				}

				controllerType := strings.TrimSpace(blocks[2])
				if controllerType == "onos" {
					version := strings.TrimSpace(blocks[3])
					name := strings.TrimSpace(blocks[4])

					var controllerConfig ControllerConfig
					currentController = &controllerConfig

					currentController.ControllerType = controllerType
					currentController.ControllerVersion = version
					currentController.ControllerName = name
					currentController.ExternalIpAddress = currentMachine.IpAddress

					numControllers := len(currentMachine.ControllersConfig)

					currentController.PortREST = uint16(ONOSPORTREST + numControllers)
					currentController.PortOpenFlow = uint16(ONOSOPENFLOW + numControllers)
					currentController.PortKarafSSH = uint16(ONOSKARAFSSH - numControllers)
					currentController.PortDebug = uint16(ONOSPORTDEBUG + numControllers)
					currentController.PortIntraCluster = uint16(ONOSPORTCLUSTER + numControllers)
					currentController.PortOVSDB = uint16(ONOSOVSDB - numControllers)
					currentController.PortNetconf = uint16(ONOSPORTNETCONF + numControllers)
					currentController.ExternalIpAddress = currentMachine.IpAddress
					currentController.Container_Name = fmt.Sprintf("%s_%s", name, currentDeployment.Id.String()[0:7])

					currentMachine.ControllersConfig = append(currentMachine.ControllersConfig, controllerConfig)
				}else {
					fmt.Println(fmt.Sprintf("Your controller type [%s] is currently not supported", controllerType))
				}

			}

		case "show":

			if context[len(context)-1].Ctype == CONTROLLER {
				s, _ := json.MarshalIndent(currentController, "", "\t");
				fmt.Print(string(s))
				break
			}

			item := blocks[1]
			if item == "modules" {
				if len(currentSwitch.Modules) < 1 {
					fmt.Println("The current switch has no modules")
				}
				fmt.Println("Modules: ")
				for i := range currentSwitch.Modules {
					fmt.Println(" - "+currentSwitch.Modules[i])
				}
			}

		case "set":
			item := strings.TrimSpace(blocks[1])
			if (item == "model") {
				modelName := blocks[2]
				currentSwitch.Model = modelName
			}else
			if (item == "target") {
				targetName := blocks[2]
				currentSwitch.Target = targetName
			}else
			if (item == "module-order") {
				item := strings.TrimSpace(blocks[2])
				modules := strings.Split(item, ",")

				var moduleOrder []string

				for _, module := range modules {
					found := false
					for _, listModule := range currentSwitch.Modules {
						if module == listModule {
							found = true
						}
					}
					if !found {
						log.Fatal("Module "+module+" is not a one of the modules added to the current switch")
						break
					}
					moduleOrder = append(moduleOrder, module)
				}

				currentSwitch.Modules = moduleOrder

				fmt.Println( "start -> "+strings.Join(moduleOrder, " -> ")+" -> end")

			}else
			if (item == "address") {

				if currentDeployment == nil{
					log.Fatal("Please create and use a machine to execute this command")
					break
				}

				if context[len(context)-1].Ctype != MACHINE {
					log.Fatal("The context of the CLI needs to be within a project")
					break
				}

				address := strings.TrimSpace(blocks[2])
				addressParameters := strings.Split(address, ":")
				if len(addressParameters) != 2 {
					log.Fatal("Please write the address as ip_address:port for SSH with a colon between ipaddress and port.")
					break
				}

				ip, port := addressParameters[0], addressParameters[1]

				if net.ParseIP(ip) == nil {
					log.Fatal("Please provide a valid Ipv4 address")
				}

				portNumber, err := strconv.ParseUint(port, 10, 16)
				if err != nil {
					fmt.Println("Please provide a valid port number.")
					break
				}
				if portNumber < 1 || portNumber > 65535 {
					fmt.Println("Please provide a valid port number in the range of [1 - 65535]")
					break
				}

				//Keep controllers updated with public IP, maybe handy
				for i:= range currentMachine.ControllersConfig {
					currentMachine.ControllersConfig[i].ExternalIpAddress = ip
				}

				currentMachine.IpAddress = ip
				currentMachine.PortSSH = uint16(portNumber)

			} else
			if (item == "ssh") {

				if currentDeployment == nil{
					log.Fatal("Please create and use a machine to execute this command")
					break
				}
				if context[len(context)-1].Ctype != MACHINE {
					log.Fatal("The context of the CLI needs to be within a project")
					break
				}

				item := strings.TrimSpace(blocks[2])
				if item == "auth" && len(blocks) > 2{
					typeofConnection := strings.TrimSpace(blocks[3])
					if typeofConnection != "pka" && typeofConnection != "password"{
						log.Fatal("Only 'pka' or 'password' are the accepted authentication methods for SSH.")
						break
					}

					if typeofConnection == "pka" {
						enablePkaConnection(currentMachine)
						break
					}

					if typeofConnection == "password" {
						enablePasswordConnection(currentMachine)
					}
				}else
				if item == "user"{
					user := strings.TrimSpace(blocks[3])
					if user == ""{
						log.Fatal("Please provide a proper user for the SSH connection")
						break
					}

					currentMachine.User = user
				}else
				if item == "keypath"{
					public := strings.TrimSpace(blocks[3])
					private := strings.TrimSpace(blocks[4])

					_, err := os.Stat(public)
					if os.IsNotExist(err) {
						panic("The public key does not exist.")
					}

					_, err = os.Stat(public)
					if os.IsNotExist(err) {
						panic("The private key does not exist.")
					}

					currentMachine.PubKeyPath = public
					currentMachine.PrivKeyPath = private
				}

			}

		case "use":
			item := blocks[1]
			if item == "deployment" {
				name := strings.TrimSpace(blocks[2])
				if name == "" {
					log.Fatal("A name for the project has to be provided.")
					break;
				}
				found := false
				for i := range allDeployments {
					if (allDeployments[i].Name == name) {
						currentDeployment = &allDeployments[i]
						found = true
						break
					}
				}
				if !found{
					log.Fatal("There is no project with name: "+name)
				}

				projectType := ContextType{name, DEPLOYMENT}
				context = append(context, projectType)

			}else if item == "switch" {

				if currentDeployment == nil{
					log.Fatal("Before using a switch one needs to create and use a deployment, then create a switch")
					break
				}

				name := strings.TrimSpace(blocks[2])
				if name == "" {
					log.Fatal("A proper name for the switch has to be provided.")
					break;
				}
				found := false
				for i := range currentDeployment.Switches {
					if currentDeployment.Switches[i].Name == name {
						currentSwitch = &currentDeployment.Switches[i]
						found = true
					}
				}
				if !found{
					log.Fatal("There is no switch with name: "+name)
					break
				}

				switchType := ContextType{name, SWITCH}
				context = append(context, switchType)

			}else if item == "controller" {

				if currentDeployment == nil{
					log.Fatal("Before using a switch one needs to create and use a deployment, then create a switch")
					break
				}

				if context[len(context)-1].Ctype != MACHINE {
					log.Fatal("the context of your CLI needs to be within a machine.")
					break;
				}

				name := strings.TrimSpace(blocks[2])
				if name == "" {
					log.Fatal("A proper name for the switch has to be provided.")
					break;
				}
				found := false
				controllers := currentMachine.ControllersConfig
				for i := range controllers {
					if controllers[i].ControllerName == name {
						currentController = &controllers[i]
						found = true
						break
					}
				}
				if !found{
					log.Fatal("There is no controller with name '"+name+"' for the current '"+ currentMachine.Name+"' machine")
					break
				}

				controllerType := ContextType{name, CONTROLLER}
				context = append(context, controllerType)

			}else if item == "machine" {
				if currentDeployment == nil{
					log.Fatal("Before creating a switch one needs to create and use a deployment")
					break
				}

				if context[len(context)-1].Ctype != DEPLOYMENT {
					log.Fatal("The context of the CLI needs to be within a project")
					break
				}

				name := strings.TrimSpace(blocks[2])
				if name == "" {
					log.Fatal("A proper name for the machine has to be provided.")
					break;
				}
				found := false
				controllerMachines := currentDeployment.ControllerMachines
				for i := range controllerMachines {
					if controllerMachines[i].Name == name {
						currentMachine = &controllerMachines[i]
						found = true
						break
					}
				}
				if !found{
					log.Fatal("There is no machine with name '"+name+"' for the current '"+currentDeployment.Name+"' project")
					break
				}

				machineType := ContextType{name, MACHINE}
				context = append(context, machineType)

			}

			updatePrefix()

		case "add":
			if currentDeployment == nil {
				log.Fatal("You have to create and set for a project and create and use a switch.")
				break
			}

			item := blocks[1]
			if (item == "module") {
				module := blocks[2]

				/*if _, err := os.Stat("/home/p4/go/src/p4click/repository/"+currentSwitch.Model+"/"+module); os.IsNotExist(err) {
					log.Fatal("Module "+module+" does not exist.")
					break
				}*/

				currentSwitch.Modules = append(currentSwitch.Modules, module)
			}else if item == "repository" {
				domain := blocks[2]

				if domain == "" {
					log.Fatal("A proper domain has to be provided.")
					break;
				}

				p4ClickConfig.Repositories = append(p4ClickConfig.Repositories, domain)
			}

		case "test":
			if currentDeployment == nil {
				log.Fatal("You have to create and set for a project and create and use a switch.")
				break
			}

			if context[len(context)-1].Ctype == MACHINE{
				if currentMachine.PkaMode {

					client := &SSH{
						Ip: currentMachine.IpAddress,
						User : currentMachine.User,
						Port: int(currentMachine.PortSSH),
						Cert: currentMachine.PrivKeyPath,
					}
					client.Connect(CERT_PUBLIC_KEY_FILE)
					client.Close()

					fmt.Println("Connected '✓' ")
					break

				}else
				if currentMachine.PasswordMode {
					//https://stackoverflow.com/questions/2137357/getpasswd-functionality-in-go

					fmt.Print("Enter Password: ")
					bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
					pwd := string(bytePassword)

					client := &SSH{
						Ip: currentMachine.IpAddress,
						User : currentMachine.User,
						Port: int(currentMachine.PortSSH),
						Cert: pwd,
					}
					client.Connect(CERT_PASSWORD)
					//client.RunCmd("docker ps")
					client.Close()

					fmt.Println("Connected '✓' ")
					break

				}else {
					fmt.Println("Please enable either PKA or PasswordMode-based authentication, also a user and private key path if necessary.")
				}

			}

			item := blocks[1]
			if (item == "module") {
				module := blocks[2]
				if _, err := os.Stat("/home/p4/go/src/p4click/repository/"+currentSwitch.Model+"/"+module); os.IsNotExist(err) {
					log.Fatal("Module "+module+" does not exist.")
					break
				}

				currentSwitch.Modules = append(currentSwitch.Modules, module)
			}else
			if item == "controller"{

				if currentDeployment == nil{
					log.Fatal("Please create and use a machine to execute this command")
					break
				}
				if context[len(context)-1].Ctype != MACHINE {
					log.Fatal("The context of the CLI needs to be within a project")
					break
				}

				controllerType := strings.TrimSpace(blocks[2])
				if controllerType == "onos" {
					version := strings.TrimSpace(blocks[3])
					name := strings.TrimSpace(blocks[4])

					var controllerConfig ControllerConfig
					controllerConfig.ControllerType = controllerType
					controllerConfig.ControllerVersion = version
					controllerConfig.ControllerName = name
					controllerConfig.ExternalIpAddress = currentMachine.IpAddress

					numControllers := len(currentMachine.ControllersConfig)

					currentController.PortREST = uint16(ONOSPORTREST + numControllers)
					currentController.PortOpenFlow = uint16(ONOSOPENFLOW + numControllers)
					currentController.PortKarafSSH = uint16(ONOSKARAFSSH - numControllers)
					currentController.PortDebug = uint16(ONOSPORTDEBUG + numControllers)
					currentController.PortIntraCluster = uint16(ONOSPORTCLUSTER + numControllers)
					currentController.PortOVSDB = uint16(ONOSOVSDB - numControllers)

					currentMachine.ControllersConfig = append(currentMachine.ControllersConfig, controllerConfig)
				}

			}

		case "deployc":
			deployController("docker-compose-onos.yaml")

		case "deploym":
			deployDataPlaneInfrastructure("topo-p4click.py","/tmp/deploy/docker-compose-mininet.yaml")

		case "aclapp":
			installOnosApp("p4click-aclhandler-1.0-SNAPSHOT.oar")

		case "l2app":
			installOnosApp("p4click-l2forwarding-1.0-SNAPSHOT.oar")

		case "l3app":
			installOnosApp("p4click-l3forwarding-1.0-SNAPSHOT.oar")

		case "deploynet":
			installNetConfig("netcfg-p4click.json")

	case "exit":
			size := len(context) - 1
			context = append(context[:size])

			updatePrefix()

			os.Exit(0)

		case "save":
			fmt.Println("Save to disk here")

		case "remove":
			if context[len(context)-1].Ctype == MACHINE{
				if len(blocks) > 1 {
					removeController()
				}

			}

		case "build":

			start := time.Now()

			item1 := blocks[1]
			item2 := blocks[2]

			compound := item1+" "+item2

			if (compound == "data plane") {

				pipeline.PullConfiguration(
					currentSwitch.Modules,
					currentSwitch.Model,
					currentSwitch.Target,
					currentDeployment.LocalTempDir,
					p4ClickConfig.Repositories[0])

				duration := time.Since(start)
				fmt.Println(fmt.Sprintf("Total Pullconfig: %d", duration.Microseconds()))

				start = time.Now()
				pipeline.BuildPipeline(
					currentSwitch.Modules,
					currentDeployment.LocalTempDir,
					currentSwitch.Target,
					currentSwitch.Model)

				duration = time.Since(start)
				fmt.Println(fmt.Sprintf("build data plane Pullconfig: %d", duration.Microseconds()))
			}

	}
}

func updatePrefix(){
	prefixVisible := ""
	prefixHidden := ""

	iteration := 0
	for i := len(context)-1; i >= 0; i-- {
		iteration += 1

		if len(context) == 1 {
			prefixVisible = context[i].Name
			break
		}

		if iteration == 1 {
			prefixVisible = context[i].Name
		}else if iteration > 2 {
			prefixHidden = "../"+prefixHidden
		}else{
			prefixVisible = context[i].Name+"/"+prefixVisible
		}

	}

	LivePrefixState.LivePrefix = prefixHidden + prefixVisible + " > "
	LivePrefixState.IsEnable = true

}

func changeLivePrefix() (string, bool) {
	return LivePrefixState.LivePrefix, LivePrefixState.IsEnable
}
/*
func deployController(){

	if context[len(context)-1].Ctype == CONTROLLER{
		if currentMachine.PkaMode {
			if currentMachine.PrivKeyPath == "" || currentMachine.PubKeyPath == ""{
				fmt.Println("Please add public and private keys")
				return
			}

			client := &SSH{
				Ip: currentMachine.IpAddress,
				User : currentMachine.User,
				Port: int(currentMachine.PortSSH),
				Cert: currentMachine.PrivKeyPath,
			}
			client.Connect(CERT_PUBLIC_KEY_FILE)
			output, err := client.GetCmdOutput(fmt.Sprintf("docker run -t -d -p %d:6653 -p %d:8181 -p %d:8101 -p %d:5005 -p %d:9876 -p %d:830 " +
				          "-e JAVA_DEBUG_PORT=\"0.0.0.0:5005\" -e ONOS_APPS=proxyarp,hostprovider,lldpprovider,drivers.bmv2 " +
				          "--name %s_%s onosproject/onos:%s debug", currentController.PortOpenFlow, currentController.PortREST,
				          currentController.PortKarafSSH, currentController.PortDebug, currentController.PortIntraCluster,
				          currentController.PortNetconf, currentController.ControllerName, currentDeployment.Id.String()[0:7],
				          currentController.ControllerVersion))
			if err != nil {
				fmt.Println("There was an error deploying the controller. Type 'remove controller name' ")
				return
			}
			client.Close()

			client.Connect(CERT_PUBLIC_KEY_FILE)
			output, _ = client.GetCmdOutput(fmt.Sprintf("docker ps -aqf \"name=%s_%s\"",
				currentController.ControllerName, currentDeployment.Id.String()[0:7]))
			currentController.Container_Id = output
			fmt.Println("Controller deployed! Container ID: %s", output)
			client.Close()

			client.Connect(CERT_PUBLIC_KEY_FILE)
			output, _ = client.GetCmdOutput(fmt.Sprintf("docker inspect --format " +
				"'{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' %s",currentController.Container_Id))
			currentController.InternalIpAddress = output
			client.Close()

			if err != nil {
				fmt.Print(err)
			}

		}

	}

}
*/
func removeController(){

	if context[len(context)-1].Ctype == MACHINE{
		if currentMachine.PkaMode {

			if currentMachine.PrivKeyPath == "" || currentMachine.PubKeyPath == ""{
				fmt.Println("Please add public and private keys")
				return
			}

			client := &SSH{
				Ip: currentMachine.IpAddress,
				User : currentMachine.User,
				Port: int(currentMachine.PortSSH),
				Cert: currentMachine.PrivKeyPath,
			}
			client.Connect(CERT_PUBLIC_KEY_FILE)
			output, err := client.GetCmdOutput(fmt.Sprintf("docker container stop %s", ))
			if err != nil {
				fmt.Println("There was an error deploying the controller. Type 'remove controller name' ")
				return
			}
			client.Close()

			client.Connect(CERT_PUBLIC_KEY_FILE)
			output, _ = client.GetCmdOutput(fmt.Sprintf("docker ps -aqf \"name=%s_%s\"",
				currentController.ControllerName, currentDeployment.Id.String()[0:7]))
			currentController.Container_Id = output
			fmt.Println("Controller deployed! Container ID: %s", output)
			client.Close()

			client.Connect(CERT_PUBLIC_KEY_FILE)
			output, _ = client.GetCmdOutput(fmt.Sprintf("docker inspect --format " +
				"'{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' %s",currentController.Container_Id))
			currentController.InternalIpAddress = output
			client.Close()

			if err != nil {
				fmt.Print(err)
			}

		}

	}

}


func deployController(dockerComposeFile string){


	exect := "/bin/sh"
	script := "/tmp/deploy/dep-onos.sh"

	cmdRunMininet := &exec.Cmd {
		Path: exect,
		Args: []string{ exect, script },
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	// see command represented by `cmdGoVer`
	fmt.Println( cmdRunMininet.String() )

	if err:= cmdRunMininet.Run(); err != nil {
		fmt.Println("Error: ", err)
	}


}

func deployDataPlaneInfrastructure(topologyFilePath string, dockerComposeFile string){

	varDefinition := fmt.Sprintf("TOPOLOGY_PY=%s", topologyFilePath)
	dockerComFile := fmt.Sprintf("-f %s", dockerComposeFile)
	executable, _ := exec.LookPath( "docker-compose" )
	UP := "up"
	mD := "-d"


	// `go version` command
	_ = &exec.Cmd{
		Path:   executable,
		Args:   []string{executable, dockerComFile, UP, mD},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
		Env:    append(os.Environ(), varDefinition),
	}

	exect := "/bin/sh"
	script := "/tmp/deploy/dep-mn.sh"

	cmdRunMininet := &exec.Cmd {
		Path: exect,
		Args: []string{ exect, script },
		Stdout: os.Stdout,
		Stderr: os.Stdout,
		Env: append(os.Environ(), varDefinition),
	}

	// see command represented by `cmdGoVer`
	fmt.Println( cmdRunMininet.String() )

	if err:= cmdRunMininet.Run(); err != nil {
		fmt.Println("Error: ", err)
	}

}

func installNetConfig(netCfgFilePath string){

	executable, _ := exec.LookPath( "curl" )
	arguments1 := "--fail -sSL --user onos:rocks --noproxy localhost"
	arguments2 := "-X POST -H 'Content-Type:application/json'"
	url := "http://localhost:8181/onos/v1/network/configuration"
	netCfgFile := fmt.Sprintf(" -d /tmp/deploy/%s", netCfgFilePath)

	// `go version` command
	cmdRunMininet := &exec.Cmd {
		Path: executable,
		Args: []string{ executable, arguments1, arguments2, url, netCfgFile},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
		Env: os.Environ(),
	}

	fmt.Println( cmdRunMininet.String() )

	if err:= cmdRunMininet.Run(); err != nil {
		fmt.Println("Error: ", err)
	}
}

func installOnosApp(appName string){

	executable, _ := exec.LookPath( "curl" )
	arguments1 := "--fail -sSL --user onos:rocks --noproxy localhost"
	arguments2 := "-X POST -HContent-Type:application/octet-stream"
	url := "'http://localhost:8181/onos/v1/applications?activate=true'"
	app := fmt.Sprintf("--data-binary /tmp/deploy/%s", appName)

	// `go version` command
	cmdRunMininet := &exec.Cmd {
		Path: executable,
		Args: []string{ executable, arguments1, arguments2, url, app},
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}

	// see command represented by `cmdGoVer`
	fmt.Println( cmdRunMininet.String() )

	if err:= cmdRunMininet.Run(); err != nil {
		fmt.Println("Error: ", err)
	}
}

//    make app-install-acl
//    make netcfg-p4click
//    make app-install-l2
//    make app-install-l3

func checkExposedPorts() error {

	if context[len(context)-1].Ctype == MACHINE {
		if currentMachine.PkaMode {
			if currentMachine.PrivKeyPath == "" || currentMachine.PubKeyPath == ""{
				fmt.Println("Please add public and private keys")
				return errors.New("No private key")
			}

			client := &SSH{
				Ip: currentMachine.IpAddress,
				User : currentMachine.User,
				Port: int(currentMachine.PortSSH),
				Cert: currentMachine.PrivKeyPath,
			}
			client.Connect(CERT_PUBLIC_KEY_FILE)
			output, err := client.GetCmdOutput("netstat -tulpn | awk '{print $1 \",\" $4}'")
			if err != nil{
				return err
			}

			//ports := make(map[string]struct{})
			var ports []string
			scanner := bufio.NewScanner(strings.NewReader(output))
			for scanner.Scan() {
				line := scanner.Text()
				port := strings.Split(line, ":")
				port = port[:len(port)-1]
				ports = append(ports, port[0])
			}
			client.Close()
		}
	}

	return nil
}

func enablePasswordConnection (config *MachineConfig){
	config.PasswordMode = true
	config.PkaMode = false
	config.PrivKeyPath = ""
	config.PubKeyPath = ""
}

func enablePkaConnection (config *MachineConfig){
	config.PasswordMode = false
	config.PkaMode = true
	config.PrivKeyPath = ""
	config.PubKeyPath = ""
}

func main() {

	var pcc P4ClickConfig
	p4ClickConfig = &pcc

	// Used to download modules, build pipeline, etc.
	//p4ClickDir = createRandomTempDirectory(os.TempDir(), P4CLICK)
	p4ClickDir = filepath.Join("/tmp", P4CLICK)
	os.MkdirAll(p4ClickDir, os.ModePerm)
	fmt.Println(fmt.Sprintf("%s", p4ClickDir))

	//t := prompt.Input("p4click - "+CONTEXT+">", completer)
	p4clickType := ContextType{"p4click", TOOL}
	context = append(context, p4clickType)

	p := prompt.New(
		Executor,
		Complete,
		prompt.OptionPrefix("p4click > "),
		prompt.OptionLivePrefix(changeLivePrefix),
		prompt.OptionTitle("live-prefix-example"),
	)
	p.Run()
}

func createRandomTempDirectory(parent string, dirName string)(string) {

	dir, err := ioutil.TempDir(parent, dirName)
	if err != nil {
		log.Fatal(err)
	}
	return dir;
}
