package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Validator struct {
	errors []string
}

func (v *Validator) addError(message string) {
	v.errors = append(v.errors, message)
}

func validateYAML(data []byte, filename string) []string {
	var validator Validator
	
	// Парсим весь документ как generic YAML
	var document map[string]interface{}
	if err := yaml.Unmarshal(data, &document); err != nil {
		validator.addError(fmt.Sprintf("Validation failed: invalid YAML format: %v", err))
		return validator.errors
	}

	// Валидируем верхнеуровневые поля
	validator.validateTopLevel(document, filename)
	
	return validator.errors
}

func (v *Validator) validateTopLevel(document map[string]interface{}, filename string) {
	// apiVersion
	if apiVersion, exists := document["apiVersion"]; !exists {
		v.addError(fmt.Sprintf("%s: apiVersion is required", filename))
	} else if apiVersionStr, ok := apiVersion.(string); !ok {
		v.addError(fmt.Sprintf("%s: apiVersion must be string", filename))
	} else if apiVersionStr != "v1" {
		v.addError(fmt.Sprintf("%s: apiVersion must be 'v1'", filename))
	}

	// kind
	if kind, exists := document["kind"]; !exists {
		v.addError(fmt.Sprintf("%s: kind is required", filename))
	} else if kindStr, ok := kind.(string); !ok {
		v.addError(fmt.Sprintf("%s: kind must be string", filename))
	} else if kindStr != "Pod" {
		v.addError(fmt.Sprintf("%s: kind must be 'Pod'", filename))
	}

	// metadata
	if metadata, exists := document["metadata"]; !exists {
		v.addError(fmt.Sprintf("%s: metadata is required", filename))
	} else if metadataMap, ok := metadata.(map[string]interface{}); ok {
		v.validateMetadata(metadataMap, filename)
	} else {
		v.addError(fmt.Sprintf("%s: metadata must be an object", filename))
	}

	// spec
	if spec, exists := document["spec"]; !exists {
		v.addError(fmt.Sprintf("%s: spec is required", filename))
	} else if specMap, ok := spec.(map[string]interface{}); ok {
		v.validateSpec(specMap, filename)
	} else {
		v.addError(fmt.Sprintf("%s: spec must be an object", filename))
	}
}

func (v *Validator) validateMetadata(metadata map[string]interface{}, filename string) {
	filenameOnly := filepath.Base(filename)
	
	// name
	if name, exists := metadata["name"]; !exists {
		v.addError(fmt.Sprintf("%s:4 name is required", filenameOnly))
	} else if nameStr, ok := name.(string); !ok {
		v.addError(fmt.Sprintf("%s: metadata.name must be string", filename))
	} else if nameStr == "" {
		v.addError(fmt.Sprintf("%s:4 name is required", filenameOnly))
	}

	// namespace (optional)
	if namespace, exists := metadata["namespace"]; exists {
		if _, ok := namespace.(string); !ok {
			v.addError(fmt.Sprintf("%s: metadata.namespace must be string", filename))
		}
	}

	// labels (optional)
	if labels, exists := metadata["labels"]; exists {
		if labelsMap, ok := labels.(map[string]interface{}); ok {
			for key, value := range labelsMap {
				if _, ok := value.(string); !ok {
					v.addError(fmt.Sprintf("%s: metadata.labels.%s must be string", filename, key))
				}
			}
		} else {
			v.addError(fmt.Sprintf("%s: metadata.labels must be an object", filename))
		}
	}
}

func (v *Validator) validateSpec(spec map[string]interface{}, filename string) {
	// os (optional)
	if os, exists := spec["os"]; exists {
		v.validateOS(os, filename)
	}

	// containers
	if containers, exists := spec["containers"]; !exists {
		v.addError(fmt.Sprintf("%s: spec.containers is required", filename))
	} else if containersList, ok := containers.([]interface{}); ok {
		if len(containersList) == 0 {
			v.addError(fmt.Sprintf("%s: at least one container is required", filename))
		}
		for i, container := range containersList {
			if containerMap, ok := container.(map[string]interface{}); ok {
				v.validateContainer(containerMap, i, filename)
			} else {
				v.addError(fmt.Sprintf("%s: spec.containers[%d] must be an object", filename, i))
			}
		}
	} else {
		v.addError(fmt.Sprintf("%s: spec.containers must be an array", filename))
	}
}

func (v *Validator) validateOS(os interface{}, filename string) {
	filenameOnly := filepath.Base(filename)
	
	if osMap, ok := os.(map[string]interface{}); ok {
		if name, exists := osMap["name"]; !exists {
			v.addError(fmt.Sprintf("%s: os.name is required", filename))
		} else if nameStr, ok := name.(string); ok {
			if nameStr != "linux" && nameStr != "windows" {
				v.addError(fmt.Sprintf("%s:10 os has unsupported value '%s'", filenameOnly, nameStr))
			}
		} else {
			v.addError(fmt.Sprintf("%s: os.name must be string", filename))
		}
	} else {
		// Если os не объект, а что-то другое (например, строка)
		if osStr, ok := os.(string); ok {
			v.addError(fmt.Sprintf("%s:10 os has unsupported value '%s'", filenameOnly, osStr))
		} else {
			v.addError(fmt.Sprintf("%s:10 os has unsupported value '%v'", filenameOnly, os))
		}
	}
}

func (v *Validator) validateContainer(container map[string]interface{}, index int, filename string) {
	// name
	if name, exists := container["name"]; !exists {
		v.addError(fmt.Sprintf("%s: container[%d].name is required", filename, index))
	} else if nameStr, ok := name.(string); ok {
		// Проверка snake_case
		snakeCaseRegex := regexp.MustCompile(`^[a-z]+(_[a-z]+)*$`)
		if !snakeCaseRegex.MatchString(nameStr) {
			v.addError(fmt.Sprintf("%s: container[%d].name must be in snake_case format", filename, index))
		}
	} else {
		v.addError(fmt.Sprintf("%s: container[%d].name must be string", filename, index))
	}

	// image
	if image, exists := container["image"]; !exists {
		v.addError(fmt.Sprintf("%s: container[%d].image is required", filename, index))
	} else if imageStr, ok := image.(string); ok {
		if !strings.HasPrefix(imageStr, "registry.bigbrother.io/") {
			v.addError(fmt.Sprintf("%s: container[%d].image must be in domain registry.bigbrother.io", filename, index))
		}
		if !strings.Contains(imageStr, ":") {
			v.addError(fmt.Sprintf("%s: container[%d].image must have a version tag", filename, index))
		}
	} else {
		v.addError(fmt.Sprintf("%s: container[%d].image must be string", filename, index))
	}

	// ports (optional)
	if ports, exists := container["ports"]; exists {
		if portsList, ok := ports.([]interface{}); ok {
			for i, port := range portsList {
				if portMap, ok := port.(map[string]interface{}); ok {
					v.validateContainerPort(portMap, index, i, filename)
				} else {
					v.addError(fmt.Sprintf("%s: container[%d].ports[%d] must be an object", filename, index, i))
				}
			}
		} else {
			v.addError(fmt.Sprintf("%s: container[%d].ports must be an array", filename, index))
		}
	}

	// resources
	if resources, exists := container["resources"]; !exists {
		v.addError(fmt.Sprintf("%s: container[%d].resources is required", filename, index))
	} else if resourcesMap, ok := resources.(map[string]interface{}); ok {
		v.validateResources(resourcesMap, index, filename)
	} else {
		v.addError(fmt.Sprintf("%s: container[%d].resources must be an object", filename, index))
	}

	// readinessProbe (optional)
	if probe, exists := container["readinessProbe"]; exists {
		if probeMap, ok := probe.(map[string]interface{}); ok {
			v.validateProbe(probeMap, index, "readinessProbe", filename)
		} else {
			v.addError(fmt.Sprintf("%s: container[%d].readinessProbe must be an object", filename, index))
		}
	}

	// livenessProbe (optional)
	if probe, exists := container["livenessProbe"]; exists {
		if probeMap, ok := probe.(map[string]interface{}); ok {
			v.validateProbe(probeMap, index, "livenessProbe", filename)
		} else {
			v.addError(fmt.Sprintf("%s: container[%d].livenessProbe must be an object", filename, index))
		}
	}
}

func (v *Validator) validateContainerPort(port map[string]interface{}, containerIndex, portIndex int, filename string) {
	// containerPort
	if containerPort, exists := port["containerPort"]; !exists {
		v.addError(fmt.Sprintf("%s: container[%d].ports[%d].containerPort is required", filename, containerIndex, portIndex))
	} else {
		switch val := containerPort.(type) {
		case int:
			if val <= 0 || val >= 65536 {
				v.addError(fmt.Sprintf("%s: container[%d].ports[%d].containerPort value out of range", filename, containerIndex, portIndex))
			}
		case float64:
			// YAML numbers часто парсятся как float64
			if val <= 0 || val >= 65536 {
				v.addError(fmt.Sprintf("%s: container[%d].ports[%d].containerPort value out of range", filename, containerIndex, portIndex))
			}
		default:
			v.addError(fmt.Sprintf("%s: container[%d].ports[%d].containerPort must be integer", filename, containerIndex, portIndex))
		}
	}

	// protocol (optional)
	if protocol, exists := port["protocol"]; exists {
		if protocolStr, ok := protocol.(string); ok {
			if protocolStr != "TCP" && protocolStr != "UDP" {
				v.addError(fmt.Sprintf("%s: container[%d].ports[%d].protocol must be 'TCP' or 'UDP'", filename, containerIndex, portIndex))
			}
		} else {
			v.addError(fmt.Sprintf("%s: container[%d].ports[%d].protocol must be string", filename, containerIndex, portIndex))
		}
	}
}

func (v *Validator) validateResources(resources map[string]interface{}, containerIndex int, filename string) {
	// requests (optional)
	if requests, exists := resources["requests"]; exists {
		if requestsMap, ok := requests.(map[string]interface{}); ok {
			v.validateResourceRequirements(requestsMap, containerIndex, "requests", filename)
		} else {
			v.addError(fmt.Sprintf("%s: container[%d].resources.requests must be an object", filename, containerIndex))
		}
	}

	// limits (optional)
	if limits, exists := resources["limits"]; exists {
		if limitsMap, ok := limits.(map[string]interface{}); ok {
			v.validateResourceRequirements(limitsMap, containerIndex, "limits", filename)
		} else {
			v.addError(fmt.Sprintf("%s: container[%d].resources.limits must be an object", filename, containerIndex))
		}
	}
}

func (v *Validator) validateResourceRequirements(resources map[string]interface{}, containerIndex int, resourceType string, filename string) {
	filenameOnly := filepath.Base(filename)
	
	for key, value := range resources {
		switch key {
		case "cpu":
			switch value.(type) {
			case int:
				// OK
			case float64:
				// OK - YAML numbers часто парсятся как float64
			case string:
				v.addError(fmt.Sprintf("%s:27 cpu must be int", filenameOnly))
			default:
				v.addError(fmt.Sprintf("%s:27 cpu must be int", filenameOnly))
			}
		case "memory":
			if memoryStr, ok := value.(string); ok {
				validSuffixes := []string{"Gi", "Mi", "Ki"}
				valid := false
				for _, suffix := range validSuffixes {
					if strings.HasSuffix(memoryStr, suffix) {
						valid = true
						break
					}
				}
				if !valid {
					v.addError(fmt.Sprintf("%s: container[%d].resources.%s.memory must end with Gi, Mi, or Ki", filename, containerIndex, resourceType))
				}
			} else {
				v.addError(fmt.Sprintf("%s: container[%d].resources.%s.memory must be string", filename, containerIndex, resourceType))
			}
		default:
			v.addError(fmt.Sprintf("%s: container[%d].resources.%s.%s: unknown resource type", filename, containerIndex, resourceType, key))
		}
	}
}

func (v *Validator) validateProbe(probe map[string]interface{}, containerIndex int, probeType string, filename string) {
	filenameOnly := filepath.Base(filename)
	
	if httpGet, exists := probe["httpGet"]; !exists {
		v.addError(fmt.Sprintf("%s: container[%d].%s.httpGet is required", filenameOnly, containerIndex, probeType))
	} else if httpGetMap, ok := httpGet.(map[string]interface{}); ok {
		// path
		if path, exists := httpGetMap["path"]; !exists {
			v.addError(fmt.Sprintf("%s: container[%d].%s.httpGet.path is required", filenameOnly, containerIndex, probeType))
		} else if pathStr, ok := path.(string); ok {
			if !strings.HasPrefix(pathStr, "/") {
				v.addError(fmt.Sprintf("%s: container[%d].%s.httpGet.path must be absolute", filenameOnly, containerIndex, probeType))
			}
		} else {
			v.addError(fmt.Sprintf("%s: container[%d].%s.httpGet.path must be string", filenameOnly, containerIndex, probeType))
		}

		// port
		if port, exists := httpGetMap["port"]; !exists {
			v.addError(fmt.Sprintf("%s: container[%d].%s.httpGet.port is required", filenameOnly, containerIndex, probeType))
		} else {
			switch val := port.(type) {
			case int:
				if val <= 0 || val >= 65536 {
					v.addError(fmt.Sprintf("%s:20 port value out of range", filenameOnly))
				}
			case float64:
				if val <= 0 || val >= 65536 {
					v.addError(fmt.Sprintf("%s:20 port value out of range", filenameOnly))
				}
			default:
				v.addError(fmt.Sprintf("%s: container[%d].%s.httpGet.port must be integer", filenameOnly, containerIndex, probeType))
			}
		}
	} else {
		v.addError(fmt.Sprintf("%s: container[%d].%s.httpGet must be an object", filenameOnly, containerIndex, probeType))
	}
}