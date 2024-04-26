package config

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

// ServerConfig defines server-specific configurations.
//
// swagger:model ServerConfig
type ServerConfig struct {
	// Label of the server configuration used for identification.
	//
	// required: true
	Label string `yaml:"label"`

	// HTTP method used for sending data to the destination server.
	//
	// required: true
	DestinationHTTPMethod string `yaml:"destinationHttpMethod"`

	// Header name for the file resource.
	//
	// required: false
	FileHeader string `yaml:"fileHeader,omitempty"`

	// Header name for additional arguments passed to the command.
	//
	// required: false
	ArgHeader string `yaml:"argHeader,omitempty"`

	// Indicates whether the authentication header should be forwarded.
	//
	// required: false
	ForwardAuth bool `yaml:"forwardAuth,omitempty"`

	// List of MIME types allowed for processing.
	//
	// required: true
	AllowedMimeTypes []string `yaml:"allowedMimeTypes"`

	// Commands and arguments ran by MIME type.
	//
	// required: true
	CmdByMimeType map[string]Command `yaml:"cmdByMimeType"`
}

// Command describes the command and arguments to execute for a specific MIME type.
//
// swagger:model Command
type Command struct {
	// Command to execute.
	//
	// required: true
	Cmd string `yaml:"cmd"`

	// Arguments for the command.
	//
	// required: false
	Args []string `yaml:"args"`
}

func IsAllowedMimeType(mimetype string, allowedFormats []string) bool {
	for _, format := range allowedFormats {
		if format == mimetype {
			return true
		}
		if strings.HasSuffix(format, "/*") {
			// Check wildcard MIME type
			prefix := strings.TrimSuffix(format, "*")
			if strings.HasPrefix(mimetype, prefix) {
				return true
			}
		}
	}
	return false
}

func ReadConfig(yp string) (*ServerConfig, error) {
	var (
		y   []byte
		err error
	)
	yml := os.Getenv("SCYLLARIDAE_YML")
	if yml != "" {
		y = []byte(yml)
	} else {
		y, err = os.ReadFile(yp)
		if err != nil {
			return nil, err
		}
	}

	var c ServerConfig
	err = yaml.Unmarshal(y, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func BuildExecCommand(mimetype, addtlArgs string, c *ServerConfig) (*exec.Cmd, error) {
	var cmdConfig Command
	var exists bool
	if IsAllowedMimeType(mimetype, c.AllowedMimeTypes) {
		cmdConfig, exists = c.CmdByMimeType[mimetype]
		if !exists || (len(cmdConfig.Cmd) == 0) {
			// Fallback to default if specific MIME type not configured or if command is empty
			cmdConfig = c.CmdByMimeType["default"]
		}
	} else {
		return nil, fmt.Errorf("undefined mimetype: %s", mimetype)
	}

	args := []string{}
	for _, a := range cmdConfig.Args {
		if a == "%s" && addtlArgs != "" {
			args = append(args, addtlArgs)
		} else {
			args = append(args, a)
		}
	}

	cmd := exec.Command(cmdConfig.Cmd, args...)

	return cmd, nil
}
