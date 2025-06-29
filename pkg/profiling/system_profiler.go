// Package profiling provides comprehensive system topology discovery and profiling
// capabilities for cloud instances across different providers.
package profiling

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// SystemProfiler provides comprehensive system topology discovery
type SystemProfiler struct {
	containerized bool
	privileged    bool
	hostAccess    bool
}

// SystemTopology represents the complete hardware and software topology of a system
type SystemTopology struct {
	SchemaVersion         string                `json:"schema_version"`
	CollectionTimestamp   time.Time             `json:"collection_timestamp"`
	InstanceMetadata      InstanceMetadata      `json:"instance_metadata"`
	CPUTopology          CPUTopology           `json:"cpu_topology"`
	CacheHierarchy       CacheHierarchy        `json:"cache_hierarchy"`
	MemoryTopology       MemoryTopology        `json:"memory_topology"`
	VirtualizationDetails VirtualizationDetails `json:"virtualization_details"`
	BenchmarkEnvironment BenchmarkEnvironment  `json:"benchmark_environment"`
}

// InstanceMetadata contains cloud instance identification information
type InstanceMetadata struct {
	InstanceType       string `json:"instance_type"`
	InstanceID         string `json:"instance_id"`
	Region             string `json:"region"`
	AvailabilityZone   string `json:"availability_zone"`
	InstanceFamily     string `json:"instance_family"`
	VirtualizationType string `json:"virtualization_type"`
	Hypervisor         string `json:"hypervisor"`
}

// CPUTopology contains detailed CPU hardware information
type CPUTopology struct {
	Identification   CPUIdentification `json:"identification"`
	PhysicalLayout   PhysicalLayout    `json:"physical_layout"`
	Frequency        FrequencyInfo     `json:"frequency"`
}

// CPUIdentification contains CPU model and feature information
type CPUIdentification struct {
	Vendor          string            `json:"vendor"`
	ModelName       string            `json:"model_name"`
	Family          int               `json:"family"`
	Model           int               `json:"model"`
	Stepping        int               `json:"stepping"`
	Microcode       string            `json:"microcode"`
	Architecture    string            `json:"architecture"`
	InstructionSets []string          `json:"instruction_sets"`
	Features        map[string]bool   `json:"features"`
}

// PhysicalLayout describes the physical CPU topology
type PhysicalLayout struct {
	Sockets                int                `json:"sockets"`
	CoresPerSocket         int                `json:"cores_per_socket"`
	ThreadsPerCore         int                `json:"threads_per_core"`
	TotalLogicalCPUs       int                `json:"total_logical_cpus"`
	TotalPhysicalCores     int                `json:"total_physical_cores"`
	HyperthreadingEnabled  bool               `json:"hyperthreading_enabled"`
	CPUList                string             `json:"cpu_list"`
	CoreSiblings           map[int][]int      `json:"core_siblings"`
}

// FrequencyInfo contains CPU frequency and power management information
type FrequencyInfo struct {
	BaseFrequencyMHz    float64           `json:"base_frequency_mhz"`
	MaxTurboFrequencyMHz float64          `json:"max_turbo_frequency_mhz"`
	CurrentFrequencies  map[string]float64 `json:"current_frequencies"`
	Governor            string            `json:"governor"`
	ScalingDriver       string            `json:"scaling_driver"`
	FrequencyRange      FrequencyRange    `json:"frequency_range"`
	TurboEnabled        bool              `json:"turbo_enabled"`
	CStates             CStateInfo        `json:"c_states"`
}

// FrequencyRange defines the available frequency range
type FrequencyRange struct {
	MinMHz float64 `json:"min_mhz"`
	MaxMHz float64 `json:"max_mhz"`
}

// CStateInfo contains CPU idle state information
type CStateInfo struct {
	Available     []string `json:"available"`
	CurrentPolicy string   `json:"current_policy"`
}

// CacheHierarchy describes the complete cache topology
type CacheHierarchy struct {
	L1Data         CacheLevel       `json:"l1_data"`
	L1Instruction  CacheLevel       `json:"l1_instruction"`
	L2Unified      CacheLevel       `json:"l2_unified"`
	L3Unified      CacheLevel       `json:"l3_unified"`
	CacheCoherency CacheCoherency   `json:"cache_coherency"`
}

// CacheLevel describes a single cache level
type CacheLevel struct {
	SizeKB             int    `json:"size_kb"`
	Associativity      int    `json:"associativity"`
	LineSizeBytes      int    `json:"line_size_bytes"`
	Sets               int    `json:"sets"`
	Type               string `json:"type"`
	Level              int    `json:"level"`
	SharedCPUList      string `json:"shared_cpu_list"`
	WritePolicy        string `json:"write_policy,omitempty"`
	ReplacementPolicy  string `json:"replacement_policy,omitempty"`
}

// CacheCoherency describes cache coherency protocol information
type CacheCoherency struct {
	Protocol                 string `json:"protocol"`
	SnoopLatencyCycles      int    `json:"snoop_latency_cycles"`
	CrossSocketLatencyCycles int    `json:"cross_socket_latency_cycles"`
}

// MemoryTopology describes the complete memory subsystem
type MemoryTopology struct {
	TotalMemoryGB     float64          `json:"total_memory_gb"`
	AvailableMemoryGB float64          `json:"available_memory_gb"`
	MemoryLayout      []DIMMInfo       `json:"memory_layout"`
	NUMATopology      NUMATopology     `json:"numa_topology"`
	MemoryController  MemoryController `json:"memory_controller"`
}

// DIMMInfo describes individual memory modules
type DIMMInfo struct {
	DIMMSlot     int    `json:"dimm_slot"`
	SizeGB       int    `json:"size_gb"`
	Type         string `json:"type"`
	SpeedMHz     int    `json:"speed_mhz"`
	Manufacturer string `json:"manufacturer"`
	PartNumber   string `json:"part_number"`
	BankLabel    string `json:"bank_label"`
	Locator      string `json:"locator"`
}

// NUMATopology describes Non-Uniform Memory Access topology
type NUMATopology struct {
	Nodes            []NUMANode `json:"nodes"`
	InterleavePolicy string     `json:"interleave_policy"`
	MemoryPolicy     string     `json:"memory_policy"`
}

// NUMANode describes a single NUMA node
type NUMANode struct {
	NodeID     int                `json:"node_id"`
	CPUs       string             `json:"cpus"`
	MemoryGB   float64            `json:"memory_gb"`
	Distances  map[string]int     `json:"distances"`
	Hugepages  HugepageInfo       `json:"hugepages"`
}

// HugepageInfo describes hugepage availability
type HugepageInfo struct {
	MB2Total  int `json:"2mb_total"`
	MB2Free   int `json:"2mb_free"`
	GB1Total  int `json:"1gb_total"`
	GB1Free   int `json:"1gb_free"`
}

// MemoryController describes the memory controller configuration
type MemoryController struct {
	Channels         int     `json:"channels"`
	DIMMs           int     `json:"dimms_per_channel"`
	MaxBandwidthGBs float64 `json:"max_bandwidth_gb_s"`
	ECCEnabled      bool    `json:"ecc_enabled"`
	RefreshRate     string  `json:"refresh_rate"`
}

// VirtualizationDetails describes the virtualization environment
type VirtualizationDetails struct {
	Hypervisor           string                 `json:"hypervisor"`
	CPUStealTimePercent  float64                `json:"cpu_steal_time_percent"`
	MemoryBallooning     bool                   `json:"memory_ballooning"`
	SRIOVEnabled         bool                   `json:"sr_iov_enabled"`
	PCIPassthrough       bool                   `json:"pci_passthrough"`
	NestedVirtualization bool                   `json:"nested_virtualization"`
	Paravirtualization   ParavirtualizationInfo `json:"paravirtualization"`
}

// ParavirtualizationInfo describes paravirtualization features
type ParavirtualizationInfo struct {
	ClockSource    string `json:"clock_source"`
	BalloonDriver  string `json:"balloon_driver"`
}

// BenchmarkEnvironment describes the optimal benchmark execution environment
type BenchmarkEnvironment struct {
	ThreadingConfiguration ThreadingConfiguration `json:"threading_configuration"`
	MemoryConfiguration    MemoryConfiguration    `json:"memory_configuration"`
	SystemState           SystemState            `json:"system_state"`
}

// ThreadingConfiguration describes optimal threading setup
type ThreadingConfiguration struct {
	AffinityPolicy      string              `json:"affinity_policy"`
	CPUPinning          CPUPinningConfig    `json:"cpu_pinning"`
	NUMABinding         NUMABindingConfig   `json:"numa_binding"`
	HyperthreadingUsage HyperthreadingConfig `json:"hyperthreading_usage"`
}

// CPUPinningConfig describes CPU affinity configuration
type CPUPinningConfig struct {
	Enabled bool              `json:"enabled"`
	Mapping map[string]string `json:"mapping"`
}

// NUMABindingConfig describes NUMA binding configuration
type NUMABindingConfig struct {
	Enabled          bool   `json:"enabled"`
	Policy           string `json:"policy"`
	MemoryAllocation string `json:"memory_allocation"`
}

// HyperthreadingConfig describes hyperthreading usage strategy
type HyperthreadingConfig struct {
	Enabled   bool   `json:"enabled"`
	Strategy  string `json:"strategy"`
	Isolation string `json:"isolation"`
}

// MemoryConfiguration describes optimal memory setup
type MemoryConfiguration struct {
	AllocationPolicy           string `json:"allocation_policy"`
	TransparentHugepages       string `json:"transparent_hugepages"`
	SwapEnabled                bool   `json:"swap_enabled"`
	MemoryCompaction           string `json:"memory_compaction"`
	OOMKiller                  string `json:"oom_killer"`
	DropCachesBeforeBenchmark  bool   `json:"drop_caches_before_benchmark"`
}

// SystemState describes current system configuration
type SystemState struct {
	CPUGovernor                       string `json:"cpu_governor"`
	TurboBoost                        string `json:"turbo_boost"`
	CStates                           string `json:"c_states"`
	AddressSpaceLayoutRandomization   string `json:"address_space_layout_randomization"`
	InterruptAffinity                 string `json:"interrupt_affinity"`
	KernelPreemption                  string `json:"kernel_preemption"`
	TickRateHz                        int    `json:"tick_rate_hz"`
}

// NewSystemProfiler creates a new system profiler instance
func NewSystemProfiler() *SystemProfiler {
	return &SystemProfiler{
		containerized: isRunningInContainer(),
		privileged:    hasPrivilegedAccess(),
		hostAccess:    hasHostAccess(),
	}
}

// ProfileSystem performs comprehensive system topology discovery
func (sp *SystemProfiler) ProfileSystem(ctx context.Context) (*SystemTopology, error) {
	topology := &SystemTopology{
		SchemaVersion:       "2.0.0",
		CollectionTimestamp: time.Now(),
	}
	
	// Profile each subsystem
	var err error
	
	topology.InstanceMetadata, err = sp.profileInstanceMetadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to profile instance metadata: %w", err)
	}
	
	topology.CPUTopology, err = sp.profileCPUTopology(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to profile CPU topology: %w", err)
	}
	
	topology.CacheHierarchy, err = sp.profileCacheHierarchy(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to profile cache hierarchy: %w", err)
	}
	
	topology.MemoryTopology, err = sp.profileMemoryTopology(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to profile memory topology: %w", err)
	}
	
	topology.VirtualizationDetails, err = sp.profileVirtualization(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to profile virtualization: %w", err)
	}
	
	topology.BenchmarkEnvironment, err = sp.profileBenchmarkEnvironment(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to profile benchmark environment: %w", err)
	}
	
	return topology, nil
}

// profileInstanceMetadata discovers cloud instance metadata
func (sp *SystemProfiler) profileInstanceMetadata(ctx context.Context) (InstanceMetadata, error) {
	metadata := InstanceMetadata{}
	
	// Try to get AWS instance metadata
	if instanceData, err := sp.getAWSInstanceMetadata(ctx); err == nil {
		metadata = instanceData
	}
	
	// If AWS metadata not available, try to determine from other sources
	if metadata.InstanceType == "" {
		metadata.InstanceType = "unknown"
		metadata.InstanceFamily = "unknown"
		metadata.Region = "unknown"
	}
	
	// Determine virtualization type from system
	virtType, hypervisor := sp.detectVirtualization()
	metadata.VirtualizationType = virtType
	metadata.Hypervisor = hypervisor
	
	return metadata, nil
}

// profileCPUTopology discovers CPU hardware topology
func (sp *SystemProfiler) profileCPUTopology(ctx context.Context) (CPUTopology, error) {
	topology := CPUTopology{}
	
	// Parse /proc/cpuinfo
	cpuInfo, err := sp.parseCPUInfo()
	if err != nil {
		return topology, fmt.Errorf("failed to parse CPU info: %w", err)
	}
	
	topology.Identification = cpuInfo
	
	// Get physical layout using lscpu
	layout, err := sp.getCPULayout(ctx)
	if err != nil {
		return topology, fmt.Errorf("failed to get CPU layout: %w", err)
	}
	
	topology.PhysicalLayout = layout
	
	// Get frequency information
	frequency, err := sp.getCPUFrequency(ctx)
	if err != nil {
		return topology, fmt.Errorf("failed to get CPU frequency: %w", err)
	}
	
	topology.Frequency = frequency
	
	return topology, nil
}

// profileCacheHierarchy discovers cache topology
func (sp *SystemProfiler) profileCacheHierarchy(ctx context.Context) (CacheHierarchy, error) {
	hierarchy := CacheHierarchy{}
	
	// Use sysfs to get cache information
	caches, err := sp.parseCacheTopology()
	if err != nil {
		return hierarchy, fmt.Errorf("failed to parse cache topology: %w", err)
	}
	
	// Organize by cache level and type
	for _, cache := range caches {
		switch cache.Level {
		case 1:
			if cache.Type == "Data" {
				hierarchy.L1Data = cache
			} else if cache.Type == "Instruction" {
				hierarchy.L1Instruction = cache
			}
		case 2:
			hierarchy.L2Unified = cache
		case 3:
			hierarchy.L3Unified = cache
		}
	}
	
	// Get cache coherency information
	coherency, err := sp.getCacheCoherency(ctx)
	if err == nil {
		hierarchy.CacheCoherency = coherency
	}
	
	return hierarchy, nil
}

// profileMemoryTopology discovers memory subsystem topology
func (sp *SystemProfiler) profileMemoryTopology(ctx context.Context) (MemoryTopology, error) {
	topology := MemoryTopology{}
	
	// Get total memory from /proc/meminfo
	memInfo, err := sp.parseMemInfo()
	if err != nil {
		return topology, fmt.Errorf("failed to parse memory info: %w", err)
	}
	
	topology.TotalMemoryGB = memInfo.TotalGB
	topology.AvailableMemoryGB = memInfo.AvailableGB
	
	// Get DIMM information using dmidecode (if privileged)
	if sp.privileged {
		dimms, err := sp.getDIMMInfo(ctx)
		if err == nil {
			topology.MemoryLayout = dimms
		}
	}
	
	// Get NUMA topology
	numaTopology, err := sp.getNUMATopology(ctx)
	if err != nil {
		return topology, fmt.Errorf("failed to get NUMA topology: %w", err)
	}
	
	topology.NUMATopology = numaTopology
	
	// Get memory controller information
	controller, err := sp.getMemoryController(ctx)
	if err == nil {
		topology.MemoryController = controller
	}
	
	return topology, nil
}

// profileVirtualization discovers virtualization environment details
func (sp *SystemProfiler) profileVirtualization(ctx context.Context) (VirtualizationDetails, error) {
	details := VirtualizationDetails{}
	
	// Detect hypervisor and virtualization type
	_, hypervisor := sp.detectVirtualization()
	details.Hypervisor = hypervisor
	
	// Get CPU steal time from /proc/stat
	stealTime, err := sp.getCPUStealTime()
	if err == nil {
		details.CPUStealTimePercent = stealTime
	}
	
	// Check for SR-IOV and other virtualization features
	details.SRIOVEnabled = sp.checkSRIOV()
	details.PCIPassthrough = sp.checkPCIPassthrough()
	details.NestedVirtualization = sp.checkNestedVirtualization()
	
	// Get paravirtualization info
	paraInfo := sp.getParavirtualizationInfo()
	details.Paravirtualization = paraInfo
	
	return details, nil
}

// profileBenchmarkEnvironment determines optimal benchmark configuration
func (sp *SystemProfiler) profileBenchmarkEnvironment(ctx context.Context) (BenchmarkEnvironment, error) {
	env := BenchmarkEnvironment{}
	
	// Configure optimal threading for benchmarks
	threadingConfig, err := sp.getThreadingConfiguration(ctx)
	if err != nil {
		return env, fmt.Errorf("failed to get threading configuration: %w", err)
	}
	
	env.ThreadingConfiguration = threadingConfig
	
	// Configure optimal memory settings
	memoryConfig, err := sp.getMemoryConfiguration(ctx)
	if err != nil {
		return env, fmt.Errorf("failed to get memory configuration: %w", err)
	}
	
	env.MemoryConfiguration = memoryConfig
	
	// Get current system state
	systemState, err := sp.getSystemState(ctx)
	if err != nil {
		return env, fmt.Errorf("failed to get system state: %w", err)
	}
	
	env.SystemState = systemState
	
	return env, nil
}

// Helper functions for system discovery

// isRunningInContainer detects if running inside a container
func isRunningInContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		return strings.Contains(string(data), "docker") || strings.Contains(string(data), "containerd")
	}
	
	return false
}

// hasPrivilegedAccess checks if the process has privileged access
func hasPrivilegedAccess() bool {
	return os.Geteuid() == 0
}

// hasHostAccess checks if the process can access host information
func hasHostAccess() bool {
	// Check if /proc/cpuinfo is accessible
	_, err := os.Stat("/proc/cpuinfo")
	return err == nil
}

// getAWSInstanceMetadata retrieves AWS EC2 instance metadata
func (sp *SystemProfiler) getAWSInstanceMetadata(ctx context.Context) (InstanceMetadata, error) {
	metadata := InstanceMetadata{}
	
	// Try IMDSv2 first
	token, err := sp.getIMDSv2Token(ctx)
	if err != nil {
		return metadata, err
	}
	
	// Get instance type
	instanceType, err := sp.getMetadataWithToken(ctx, "instance-type", token)
	if err != nil {
		return metadata, err
	}
	metadata.InstanceType = instanceType
	
	// Extract instance family
	if parts := strings.Split(instanceType, "."); len(parts) > 0 {
		metadata.InstanceFamily = parts[0]
	}
	
	// Get instance ID
	instanceID, err := sp.getMetadataWithToken(ctx, "instance-id", token)
	if err == nil {
		metadata.InstanceID = instanceID
	}
	
	// Get region and AZ
	az, err := sp.getMetadataWithToken(ctx, "placement/availability-zone", token)
	if err == nil {
		metadata.AvailabilityZone = az
		if len(az) > 0 {
			metadata.Region = az[:len(az)-1] // Remove the zone letter
		}
	}
	
	return metadata, nil
}

// getIMDSv2Token gets an IMDSv2 token for AWS metadata access
func (sp *SystemProfiler) getIMDSv2Token(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "curl", "-X", "PUT", 
		"http://169.254.169.254/latest/api/token",
		"-H", "X-aws-ec2-metadata-token-ttl-seconds: 21600",
		"--connect-timeout", "5")
	
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	
	return strings.TrimSpace(string(output)), nil
}

// getMetadataWithToken retrieves metadata using IMDSv2 token
func (sp *SystemProfiler) getMetadataWithToken(ctx context.Context, path, token string) (string, error) {
	cmd := exec.CommandContext(ctx, "curl",
		"-H", fmt.Sprintf("X-aws-ec2-metadata-token: %s", token),
		fmt.Sprintf("http://169.254.169.254/latest/meta-data/%s", path),
		"--connect-timeout", "5")
	
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	
	return strings.TrimSpace(string(output)), nil
}

// detectVirtualization detects the virtualization type and hypervisor
func (sp *SystemProfiler) detectVirtualization() (string, string) {
	// Check /proc/cpuinfo for hypervisor flag
	if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		content := string(data)
		if strings.Contains(content, "hypervisor") {
			// Try to determine hypervisor type
			if strings.Contains(content, "QEMU") {
				return "hvm", "QEMU/KVM"
			}
			return "hvm", "unknown"
		}
	}
	
	// Check DMI information for cloud providers
	if data, err := os.ReadFile("/sys/class/dmi/id/sys_vendor"); err == nil {
		vendor := strings.TrimSpace(string(data))
		switch {
		case strings.Contains(vendor, "Amazon"):
			return "hvm", "AWS Nitro"
		case strings.Contains(vendor, "Google"):
			return "hvm", "Google Compute Engine"
		case strings.Contains(vendor, "Microsoft"):
			return "hvm", "Hyper-V"
		}
	}
	
	return "unknown", "unknown"
}

// parseCPUInfo parses /proc/cpuinfo for CPU identification
func (sp *SystemProfiler) parseCPUInfo() (CPUIdentification, error) {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return CPUIdentification{}, err
	}
	defer file.Close()
	
	cpuInfo := CPUIdentification{
		Features:        make(map[string]bool),
		InstructionSets: []string{},
	}
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			
			switch key {
			case "vendor_id":
				cpuInfo.Vendor = value
			case "model name":
				cpuInfo.ModelName = value
			case "cpu family":
				if family, err := strconv.Atoi(value); err == nil {
					cpuInfo.Family = family
				}
			case "model":
				if model, err := strconv.Atoi(value); err == nil {
					cpuInfo.Model = model
				}
			case "stepping":
				if stepping, err := strconv.Atoi(value); err == nil {
					cpuInfo.Stepping = stepping
				}
			case "microcode":
				cpuInfo.Microcode = value
			case "flags":
				flags := strings.Fields(value)
				for _, flag := range flags {
					cpuInfo.Features[flag] = true
					
					// Identify instruction sets
					if isInstructionSet(flag) {
						cpuInfo.InstructionSets = append(cpuInfo.InstructionSets, strings.ToUpper(flag))
					}
				}
			}
		}
	}
	
	cpuInfo.Architecture = "x86_64" // Could be detected more dynamically
	
	return cpuInfo, scanner.Err()
}

// getCPULayout discovers physical CPU layout using lscpu
func (sp *SystemProfiler) getCPULayout(ctx context.Context) (PhysicalLayout, error) {
	layout := PhysicalLayout{}
	
	// Use lscpu command
	cmd := exec.CommandContext(ctx, "lscpu", "-p=CPU,CORE,SOCKET,NODE")
	output, err := cmd.Output()
	if err != nil {
		return layout, err
	}
	
	lines := strings.Split(string(output), "\n")
	cpuToCore := make(map[int]int)
	coreToSocket := make(map[int]int)
	coreSiblings := make(map[int][]int)
	
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		
		fields := strings.Split(line, ",")
		if len(fields) >= 3 {
			cpu, _ := strconv.Atoi(fields[0])
			core, _ := strconv.Atoi(fields[1])
			socket, _ := strconv.Atoi(fields[2])
			
			cpuToCore[cpu] = core
			coreToSocket[core] = socket
			
			// Build core siblings map
			if _, exists := coreSiblings[core]; !exists {
				coreSiblings[core] = []int{}
			}
			coreSiblings[core] = append(coreSiblings[core], cpu)
		}
	}
	
	// Calculate layout statistics
	sockets := make(map[int]bool)
	cores := make(map[int]bool)
	
	for core, socket := range coreToSocket {
		sockets[socket] = true
		cores[core] = true
	}
	
	layout.Sockets = len(sockets)
	layout.TotalPhysicalCores = len(cores)
	layout.TotalLogicalCPUs = len(cpuToCore)
	if layout.TotalPhysicalCores > 0 {
		layout.ThreadsPerCore = layout.TotalLogicalCPUs / layout.TotalPhysicalCores
	}
	if layout.Sockets > 0 {
		layout.CoresPerSocket = layout.TotalPhysicalCores / layout.Sockets
	}
	layout.HyperthreadingEnabled = layout.ThreadsPerCore > 1
	layout.CoreSiblings = coreSiblings
	
	// Build CPU list string
	cpuList := make([]string, 0, layout.TotalLogicalCPUs)
	for cpu := 0; cpu < layout.TotalLogicalCPUs; cpu++ {
		cpuList = append(cpuList, strconv.Itoa(cpu))
	}
	layout.CPUList = strings.Join(cpuList, ",")
	
	return layout, nil
}

// getCPUFrequency gets CPU frequency information
func (sp *SystemProfiler) getCPUFrequency(ctx context.Context) (FrequencyInfo, error) {
	freq := FrequencyInfo{
		CurrentFrequencies: make(map[string]float64),
	}
	
	// Get base and max frequencies from /proc/cpuinfo
	if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		content := string(data)
		if matches := regexp.MustCompile(`cpu MHz\s*:\s*(\d+\.?\d*)`).FindStringSubmatch(content); len(matches) > 1 {
			if baseFreq, err := strconv.ParseFloat(matches[1], 64); err == nil {
				freq.BaseFrequencyMHz = baseFreq
			}
		}
	}
	
	// Try to get scaling info from sysfs
	if gov, err := os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor"); err == nil {
		freq.Governor = strings.TrimSpace(string(gov))
	}
	
	if driver, err := os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_driver"); err == nil {
		freq.ScalingDriver = strings.TrimSpace(string(driver))
	}
	
	// Get current frequencies for each CPU
	cpuDirs, err := filepath.Glob("/sys/devices/system/cpu/cpu[0-9]*")
	if err == nil {
		for _, cpuDir := range cpuDirs {
			cpuName := filepath.Base(cpuDir)
			freqFile := filepath.Join(cpuDir, "cpufreq/scaling_cur_freq")
			if freqData, err := os.ReadFile(freqFile); err == nil {
				if freqKHz, err := strconv.ParseFloat(strings.TrimSpace(string(freqData)), 64); err == nil {
					freq.CurrentFrequencies[cpuName] = freqKHz / 1000 // Convert to MHz
				}
			}
		}
	}
	
	return freq, nil
}

// Additional helper functions would be implemented for the remaining methods...
// This is a comprehensive starting point that can be extended with the full implementation

// isInstructionSet determines if a CPU flag represents an instruction set
func isInstructionSet(flag string) bool {
	instructionSets := map[string]bool{
		"sse": true, "sse2": true, "sse3": true, "ssse3": true, "sse4_1": true, "sse4_2": true,
		"avx": true, "avx2": true, "avx512f": true, "avx512cd": true, "avx512vl": true,
		"avx512bw": true, "avx512dq": true, "avx512ifma": true, "avx512vbmi": true,
		"fma": true, "fma3": true, "fma4": true,
		"aes": true, "pclmul": true, "rdrand": true, "rdseed": true,
		"bmi1": true, "bmi2": true, "adx": true, "mpx": true,
		"sha": true, "vaes": true, "vpclmul": true,
		"neon": true, "asimd": true, "sve": true, "sve2": true,
	}
	
	return instructionSets[strings.ToLower(flag)]
}

// Placeholder implementations for remaining methods
func (sp *SystemProfiler) parseCacheTopology() ([]CacheLevel, error) {
	// TODO: Implement cache topology parsing from sysfs
	return []CacheLevel{}, nil
}

func (sp *SystemProfiler) getCacheCoherency(ctx context.Context) (CacheCoherency, error) {
	// TODO: Implement cache coherency detection
	return CacheCoherency{Protocol: "MESI"}, nil
}

type MemInfo struct {
	TotalGB     float64
	AvailableGB float64
}

func (sp *SystemProfiler) parseMemInfo() (MemInfo, error) {
	// TODO: Implement /proc/meminfo parsing
	return MemInfo{}, nil
}

func (sp *SystemProfiler) getDIMMInfo(ctx context.Context) ([]DIMMInfo, error) {
	// TODO: Implement dmidecode parsing for DIMM info
	return []DIMMInfo{}, nil
}

func (sp *SystemProfiler) getNUMATopology(ctx context.Context) (NUMATopology, error) {
	// TODO: Implement NUMA topology discovery
	return NUMATopology{}, nil
}

func (sp *SystemProfiler) getMemoryController(ctx context.Context) (MemoryController, error) {
	// TODO: Implement memory controller detection
	return MemoryController{}, nil
}

func (sp *SystemProfiler) getCPUStealTime() (float64, error) {
	// TODO: Implement CPU steal time calculation from /proc/stat
	return 0.0, nil
}

func (sp *SystemProfiler) checkSRIOV() bool {
	// TODO: Implement SR-IOV detection
	return false
}

func (sp *SystemProfiler) checkPCIPassthrough() bool {
	// TODO: Implement PCI passthrough detection
	return false
}

func (sp *SystemProfiler) checkNestedVirtualization() bool {
	// TODO: Implement nested virtualization detection
	return false
}

func (sp *SystemProfiler) getParavirtualizationInfo() ParavirtualizationInfo {
	// TODO: Implement paravirtualization info detection
	return ParavirtualizationInfo{}
}

func (sp *SystemProfiler) getThreadingConfiguration(ctx context.Context) (ThreadingConfiguration, error) {
	// TODO: Implement threading configuration optimization
	return ThreadingConfiguration{}, nil
}

func (sp *SystemProfiler) getMemoryConfiguration(ctx context.Context) (MemoryConfiguration, error) {
	// TODO: Implement memory configuration optimization
	return MemoryConfiguration{}, nil
}

func (sp *SystemProfiler) getSystemState(ctx context.Context) (SystemState, error) {
	// TODO: Implement system state detection
	return SystemState{}, nil
}