package enclave_tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"gopkg.in/yaml.v2"
)

const (
	// ArgDeliveryPort is the vsock port for batcher arg delivery. Must match NC_PORT in enclave-entrypoint.bash.
	ArgDeliveryPort uint16 = 8337
	// ReadinessPort is the vsock port for the readiness handshake. Must match READY_PORT in enclave-entrypoint.bash.
	ReadinessPort uint16 = 8338
)

type EnclaveMeasurements struct {
	PCR0 string `json:"PCR0"`
	PCR1 string `json:"PCR1"`
	PCR2 string `json:"PCR2"`
}

type EnclaverManifestSources struct {
	App string `yaml:"app"`
}

type EnclaverManifestDefaults struct {
	CpuCount uint `yaml:"cpu_count"`
	MemoryMb uint `yaml:"memory_mb"`
}

type EnclaverManifestEgress struct {
	Allow     []string `yaml:"allow"`
	Deny      []string `yaml:"deny"`
	ProxyPort uint16   `yaml:"proxy_port,omitempty"`
}

type EnclaverManifestIngress struct {
	ListenPort uint16 `yaml:"listen_port"`
}

type EnclaverManifest struct {
	Version  string                    `yaml:"version"`
	Name     string                    `yaml:"name"`
	Target   string                    `yaml:"target"`
	Sources  *EnclaverManifestSources  `yaml:"sources,omitempty"`
	Defaults *EnclaverManifestDefaults `yaml:"defaults,omitempty"`
	Egress   *EnclaverManifestEgress   `yaml:"egress,omitempty"`
	Ingress  []EnclaverManifestIngress `yaml:"ingress"`
}

func DefaultManifest(name string, target string, source string, cpuCount uint, memoryMb uint) EnclaverManifest {
	return EnclaverManifest{
		Version: "v1",
		Name:    name,
		Target:  target,
		Sources: &EnclaverManifestSources{
			App: source,
		},
		Defaults: &EnclaverManifestDefaults{
			CpuCount: cpuCount,
			MemoryMb: memoryMb,
		},
		Egress: &EnclaverManifestEgress{
			ProxyPort: 10000,
			Allow:     []string{"0.0.0.0/0", "**", "::/0"},
		},
		Ingress: []EnclaverManifestIngress{
			{ListenPort: ArgDeliveryPort}, // batcher arg delivery
			{ListenPort: ReadinessPort},   // readiness handshake
		},
	}
}

// BuildEifFromImage builds an EIF image by wrapping a pre-built app Docker image with Enclaver.
func BuildEifFromImage(ctx context.Context, appImage string, eifTag string, cpuCount uint, memoryMb uint) (EnclaveMeasurements, error) {
	return buildEnclave(ctx, DefaultManifest("op-batcher", eifTag, appImage, cpuCount, memoryMb))
}

type enclaverBuildOutput struct {
	Measurements EnclaveMeasurements `json:"Measurements"`
}

// buildEnclave builds an enclaver EIF image using the provided manifest. If build is successful,
// it returns the image's Measurements.
func buildEnclave(ctx context.Context, manifest EnclaverManifest) (EnclaveMeasurements, error) {
	tempfile, err := os.CreateTemp("", "enclaver-manifest")
	if err != nil {
		return EnclaveMeasurements{}, err
	}
	defer os.Remove(tempfile.Name())

	if err := yaml.NewEncoder(tempfile).Encode(manifest); err != nil {
		return EnclaveMeasurements{}, err
	}

	var stdout bytes.Buffer
	cmd := exec.CommandContext(
		ctx,
		"enclaver",
		"build",
		"--file",
		tempfile.Name(),
	)
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return EnclaveMeasurements{}, err
	}

	// Find measurements in the output
	re := regexp.MustCompile(`\{[\s\S]*"Measurements"[\s\S]*\}`)
	jsonMatch := re.Find(stdout.Bytes())
	if jsonMatch == nil {
		return EnclaveMeasurements{}, fmt.Errorf("could not find measurements JSON in output")
	}

	var output enclaverBuildOutput
	if err := json.Unmarshal(jsonMatch, &output); err != nil {
		return EnclaveMeasurements{}, fmt.Errorf("failed to parse measurements JSON: %w", err)
	}

	return output.Measurements, nil
}
