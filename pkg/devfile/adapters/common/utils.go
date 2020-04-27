package common

import (
	"fmt"
	"os"

	"github.com/golang/glog"

	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// PredefinedDevfileCommands encapsulates constants for predefined devfile commands
type PredefinedDevfileCommands string

const (
	// DefaultDevfileBuildCommand is a predefined devfile command for build
	DefaultDevfileBuildCommand PredefinedDevfileCommands = "devbuild"

	// DefaultDevfileRunCommand is a predefined devfile command for run
	DefaultDevfileRunCommand PredefinedDevfileCommands = "devrun"

	// SupervisordInitContainerName The init container name for supervisord
	SupervisordInitContainerName = "copy-supervisord"

	// Default Image that will be used containing the supervisord binary and assembly scripts
	// use GetBootstrapperImage() function instead of this variable
	defaultBootstrapperImage = "registry.access.redhat.com/openshiftdo/odo-init-image-rhel7:1.1.2"

	// SupervisordVolumeName Create a custom name and (hope) that users don't use the *exact* same name in their deployment (occlient.go)
	SupervisordVolumeName = "odo-supervisord-shared-data"

	// SupervisordMountPath The supervisord Mount Path for the container mounting the supervisord volume
	SupervisordMountPath = "/opt/odo/"

	// SupervisordBinaryPath The supervisord binary path inside the container volume mount
	SupervisordBinaryPath = "/opt/odo/bin/supervisord"

	// SupervisordConfFile The supervisord configuration file inside the container volume mount
	SupervisordConfFile = "/opt/odo/conf/devfile-supervisor.conf"

	// OdoInitImageContents The path to the odo init image contents
	OdoInitImageContents = "/opt/odo-init/."

	// ENV variable to overwrite image used to bootstrap SupervisorD in S2I and Devfile builder Image
	bootstrapperImageEnvName = "ODO_BOOTSTRAPPER_IMAGE"

	// Default volume size for volumes defined in a devfile
	volumeSize = "5Gi"

	// EnvCheProjectsRoot is the env defined for /projects where component mountSources=true
	EnvCheProjectsRoot = "CHE_PROJECTS_ROOT"

	// EnvOdoCommandRunWorkingDir is the env defined in the runtime component container which holds the work dir for the run command
	EnvOdoCommandRunWorkingDir = "ODO_COMMAND_RUN_WORKING_DIR"

	// EnvOdoCommandRun is the env defined in the runtime component container which holds the run command to be executed
	EnvOdoCommandRun = "ODO_COMMAND_RUN"

	// ShellExecutable is the shell executable
	ShellExecutable = "/bin/sh"

	// SupervisordCtlSubCommand is the supervisord sub command ctl
	SupervisordCtlSubCommand = "ctl"
)

func isComponentSupported(component common.DevfileComponent) bool {
	// Currently odo only uses devfile components of type dockerimage, since most of the Che registry devfiles use it
	return component.Type == common.DevfileComponentTypeDockerimage
}

// GetBootstrapperImage returns the odo-init bootstrapper image
func GetBootstrapperImage() string {
	if env, ok := os.LookupEnv(bootstrapperImageEnvName); ok {
		return env
	}
	return defaultBootstrapperImage
}

// GetSupportedComponents iterates through the components in the devfile and returns a list of odo supported components
func GetSupportedComponents(data data.DevfileData) []common.DevfileComponent {
	var components []common.DevfileComponent
	// Only components with aliases are considered because without an alias commands cannot reference them
	for _, comp := range data.GetAliasedComponents() {
		if isComponentSupported(comp) {
			glog.V(3).Infof("Found component \"%v\" with alias \"%v\"\n", comp.Type, *comp.Alias)
			components = append(components, comp)
		}
	}
	return components
}

// GetVolumes iterates through the components in the devfile and returns a map of component alias to the devfile volumes
func GetVolumes(devfileObj devfileParser.DevfileObj) map[string][]DevfileVolume {
	// componentAliasToVolumes is a map of the Devfile Component Alias to the Devfile Component Volumes
	componentAliasToVolumes := make(map[string][]DevfileVolume)
	size := volumeSize
	for _, comp := range GetSupportedComponents(devfileObj.Data) {
		if comp.Volumes != nil {
			for _, volume := range comp.Volumes {
				vol := DevfileVolume{
					Name:          volume.Name,
					ContainerPath: volume.ContainerPath,
					Size:          &size,
				}
				componentAliasToVolumes[*comp.Alias] = append(componentAliasToVolumes[*comp.Alias], vol)
			}
		}
	}
	return componentAliasToVolumes
}

// IsEnvPresent checks if the env variable is present in an array of env variables
func IsEnvPresent(envVars []common.DockerimageEnv, envVarName string) bool {
	for _, envVar := range envVars {
		if *envVar.Name == envVarName {
			return true
		}
	}

	return false
}

// IsPortPresent checks if the port is present in the endpoints array
func IsPortPresent(endpoints []common.DockerimageEndpoint, port int) bool {
	for _, endpoint := range endpoints {
		if *endpoint.Port == int32(port) {
			return true
		}
	}

	return false
}

// IsComponentBuildRequired checks if a component build is required based on the push commands, it throws an error
// if the push commands does not meet the expected criteria
func IsComponentBuildRequired(pushDevfileCommands []common.DevfileCommand) (bool, error) {
	var buildRequired bool

	switch len(pushDevfileCommands) {
	case 1: // if there is one command, it is the mandatory run command. No need to build.
		buildRequired = false
	case 2:
		// if there are two commands, it is the optional build command and the mandatory run command, set buildRequired to true
		buildRequired = true
	default:
		return false, fmt.Errorf("error executing devfile commands - there should be at least 1 command or at most 2 commands, currently there are %d commands", len(pushDevfileCommands))
	}

	return buildRequired, nil
}