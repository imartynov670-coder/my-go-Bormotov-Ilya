package main

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Структуры для парсинга YAML
type Pod struct {
	APIVersion string     `yaml:"apiVersion"`
	Kind       string     `yaml:"kind"`
	Metadata   ObjectMeta `yaml:"metadata"`
	Spec       PodSpec    `yaml:"spec"`
}

type ObjectMeta struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace,omitempty"`
	Labels    map[string]string `yaml:"labels,omitempty"`
}

type PodSpec struct {
	OS        *PodOS      `yaml:"os,omitempty"`
	Containers []Container `yaml:"containers"`
}

type PodOS struct {
	Name string `yaml:"name"`
}

type Container struct {
	Name          string               `yaml:"name"`
	Image         string               `yaml:"image"`
	Ports         []ContainerPort      `yaml:"ports,omitempty"`
	ReadinessProbe *Probe              `yaml:"readinessProbe,omitempty"`
	LivenessProbe  *Probe              `yaml:"livenessProbe,omitempty"`
	Resources     ResourceRequirements `yaml:"resources"`
}

type ContainerPort struct {
	ContainerPort int    `yaml:"containerPort"`
	Protocol     string `yaml:"protocol,omitempty"`
}

type Probe struct {
	HTTPGet HTTPGetAction `yaml:"httpGet"`
}

type HTTPGetAction struct {
	Path string `yaml:"path"`
	Port int    `yaml:"port"`
}

type ResourceRequirements struct {
	Requests map[string]interface{} `yaml:"requests,omitempty"`
	Limits   map[string]interface{} `yaml:"limits,omitempty"`
}

func validateYAML(data []byte) error {
	var pod Pod
	if err := yaml.Unmarshal(data, &pod); err != nil {
		return fmt.Errorf("invalid YAML format: %v", err)
	}

	// Валидация полей верхнего уровня
	if pod.APIVersion != "v1" {
		return fmt.Errorf("apiVersion must be 'v1'")
	}
	if pod.Kind != "Pod" {
		return fmt.Errorf("kind must be 'Pod'")
	}

	// Валидация metadata
	if pod.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}

	// Валидация spec
	if len(pod.Spec.Containers) == 0 {
		return fmt.Errorf("at least one container is required")
	}

	// Валидация PodOS если указан
	if pod.Spec.OS != nil {
		if pod.Spec.OS.Name != "linux" && pod.Spec.OS.Name != "windows" {
			return fmt.Errorf("os.name must be 'linux' or 'windows'")
		}
	}

	// Валидация контейнеров
	for i, container := range pod.Spec.Containers {
		if err := validateContainer(container, i); err != nil {
			return err
		}
	}

	return nil
}

func validateContainer(container Container, index int) error {
	// Проверка имени контейнера
	if container.Name == "" {
		return fmt.Errorf("container[%d].name is required", index)
	}
	
	// Проверка формата имени (snake_case)
	snakeCaseRegex := regexp.MustCompile(`^[a-z]+(_[a-z]+)*$`)
	if !snakeCaseRegex.MatchString(container.Name) {
		return fmt.Errorf("container[%d].name must be in snake_case format", index)
	}

	// Проверка image
	if container.Image == "" {
		return fmt.Errorf("container[%d].image is required", index)
	}
	if !strings.HasPrefix(container.Image, "registry.bigbrother.io/") {
		return fmt.Errorf("container[%d].image must be in domain registry.bigbrother.io", index)
	}
	if !strings.Contains(container.Image, ":") {
		return fmt.Errorf("container[%d].image must have a version tag", index)
	}

	// Проверка ports
	for j, port := range container.Ports {
		if port.ContainerPort <= 0 || port.ContainerPort >= 65536 {
			return fmt.Errorf("container[%d].ports[%d].containerPort must be between 1 and 65535", index, j)
		}
		if port.Protocol != "" && port.Protocol != "TCP" && port.Protocol != "UDP" {
			return fmt.Errorf("container[%d].ports[%d].protocol must be 'TCP' or 'UDP'", index, j)
		}
	}

	// Проверка probes
	if container.ReadinessProbe != nil {
		if err := validateProbe(container.ReadinessProbe, index, "readinessProbe"); err != nil {
			return err
		}
	}
	if container.LivenessProbe != nil {
		if err := validateProbe(container.LivenessProbe, index, "livenessProbe"); err != nil {
			return err
		}
	}

	// Проверка resources
	return validateResources(container.Resources, index)
}

func validateProbe(probe *Probe, containerIndex int, probeType string) error {
	if probe.HTTPGet.Path == "" {
		return fmt.Errorf("container[%d].%s.httpGet.path is required", containerIndex, probeType)
	}
	if !strings.HasPrefix(probe.HTTPGet.Path, "/") {
		return fmt.Errorf("container[%d].%s.httpGet.path must be absolute", containerIndex, probeType)
	}
	if probe.HTTPGet.Port <= 0 || probe.HTTPGet.Port >= 65536 {
		return fmt.Errorf("container[%d].%s.httpGet.port must be between 1 and 65535", containerIndex, probeType)
	}
	return nil
}

func validateResources(resources ResourceRequirements, containerIndex int) error {
	// Валидация requests
	for key, value := range resources.Requests {
		if err := validateResource(key, value, containerIndex, "requests"); err != nil {
			return err
		}
	}

	// Валидация limits
	for key, value := range resources.Limits {
		if err := validateResource(key, value, containerIndex, "limits"); err != nil {
			return err
		}
	}

	return nil
}

func validateResource(key string, value interface{}, containerIndex int, resourceType string) error {
	switch key {
	case "cpu":
		// cpu должен быть integer
		if cpu, ok := value.(int); !ok {
			return fmt.Errorf("container[%d].resources.%s.cpu must be an integer", containerIndex, resourceType)
		} else if cpu <= 0 {
			return fmt.Errorf("container[%d].resources.%s.cpu must be positive", containerIndex, resourceType)
		}
	case "memory":
		// memory должен быть string с суффиксом Gi, Mi, Ki
		if memory, ok := value.(string); !ok {
			return fmt.Errorf("container[%d].resources.%s.memory must be a string", containerIndex, resourceType)
		} else {
			validSuffixes := []string{"Gi", "Mi", "Ki"}
			valid := false
			for _, suffix := range validSuffixes {
				if strings.HasSuffix(memory, suffix) {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("container[%d].resources.%s.memory must end with Gi, Mi, or Ki", containerIndex, resourceType)
			}
		}
	default:
		return fmt.Errorf("container[%d].resources.%s.%s: unknown resource type", containerIndex, resourceType, key)
	}
	return nil
}