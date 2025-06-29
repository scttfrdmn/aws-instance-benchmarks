# Comprehensive System Profiling

## Overview

To provide meaningful benchmark comparisons and enable precise workload optimization, the AWS Instance Benchmarks project captures comprehensive system topology and configuration details. This includes CPU microarchitecture, cache hierarchy, NUMA topology, memory configuration, and threading behavior.

## Current System Information Gaps

### Current Capture (Limited)
- Basic instance type and architecture
- Container configuration
- Performance results only

### Missing Critical Details
- **CPU Microarchitecture**: Model, stepping, microcode version
- **Clock Speeds**: Base, boost, and actual frequencies during benchmarks
- **Cache Hierarchy**: L1/L2/L3 sizes, associativity, latencies
- **NUMA Topology**: Node count, CPU-to-memory mapping, cross-socket bandwidth
- **Threading Configuration**: Hyper-threading status, core pinning strategy
- **Memory Details**: Type, speed, channels, timings, topology

## Enhanced System Profiling Architecture

### 1. Multi-Stage System Discovery
```
System Profiling Pipeline:
├── Hardware Detection
│   ├── CPU Identification (/proc/cpuinfo, cpuid, lscpu)
│   ├── Cache Topology (sysfs, hwloc)
│   ├── Memory Configuration (dmidecode, lsmem)
│   └── NUMA Layout (numactl, lstopo)
├── Runtime Configuration
│   ├── Governor Settings (cpufreq)
│   ├── Thread Affinity (taskset, cgroups)
│   ├── Memory Policy (numactl)
│   └── Container Limits (cgroups v1/v2)
└── Dynamic Profiling
    ├── Frequency Monitoring (turbostat, perf)
    ├── Cache Performance (perf cache-misses)
    ├── Memory Latency (lat_mem_rd)
    └── Thread Scaling (parallel execution)
```

### 2. Enhanced Data Structure

#### System Topology Schema (`system_topology.json`)
```json
{
  "schema_version": "2.0.0",
  "collection_timestamp": "2024-06-29T18:14:04Z",
  "instance_metadata": {
    "instance_type": "m7i.large",
    "instance_id": "i-0b7ba1acb4e2c4999",
    "region": "us-east-1",
    "availability_zone": "us-east-1c",
    "instance_family": "m7i",
    "virtualization_type": "hvm",
    "hypervisor": "nitro"
  },
  "cpu_topology": {
    "identification": {
      "vendor": "GenuineIntel",
      "model_name": "Intel(R) Xeon(R) Platinum 8488C",
      "family": 6,
      "model": 143,
      "stepping": 8,
      "microcode": "0x2b0001b0",
      "architecture": "x86_64",
      "instruction_sets": ["SSE4_2", "AVX", "AVX2", "AVX512F", "AVX512CD", "AVX512VL"],
      "features": {
        "fpu": true,
        "vme": true,
        "de": true,
        "pse": true,
        "tsc": true,
        "msr": true,
        "pae": true,
        "mce": true,
        "cx8": true,
        "apic": true,
        "sep": true,
        "mtrr": true,
        "pge": true,
        "mca": true,
        "cmov": true,
        "pat": true,
        "pse36": true,
        "clflush": true,
        "mmx": true,
        "fxsr": true,
        "sse": true,
        "sse2": true,
        "ss": true,
        "ht": true,
        "tm": true,
        "ia64": false,
        "pbe": true
      }
    },
    "physical_layout": {
      "sockets": 1,
      "cores_per_socket": 1,
      "threads_per_core": 2,
      "total_logical_cpus": 2,
      "total_physical_cores": 1,
      "hyperthreading_enabled": true,
      "cpu_list": "0-1",
      "core_siblings": {
        "0": [0, 1],
        "1": [0, 1]
      }
    },
    "frequency": {
      "base_frequency_mhz": 2400,
      "max_turbo_frequency_mhz": 4100,
      "current_frequencies": {
        "cpu0": 2847.123,
        "cpu1": 2834.891
      },
      "governor": "performance",
      "scaling_driver": "intel_pstate",
      "frequency_range": {
        "min_mhz": 800,
        "max_mhz": 4100
      },
      "turbo_enabled": true,
      "c_states": {
        "available": ["C1", "C1E", "C6"],
        "current_policy": "menu"
      }
    }
  },
  "cache_hierarchy": {
    "l1_data": {
      "size_kb": 48,
      "associativity": 12,
      "line_size_bytes": 64,
      "sets": 64,
      "type": "data",
      "level": 1,
      "shared_cpu_list": "0",
      "write_policy": "write-back",
      "replacement_policy": "lru"
    },
    "l1_instruction": {
      "size_kb": 32,
      "associativity": 8,
      "line_size_bytes": 64,
      "sets": 64,
      "type": "instruction",
      "level": 1,
      "shared_cpu_list": "0"
    },
    "l2_unified": {
      "size_kb": 2048,
      "associativity": 16,
      "line_size_bytes": 64,
      "sets": 2048,
      "type": "unified",
      "level": 2,
      "shared_cpu_list": "0,1",
      "write_policy": "write-back",
      "replacement_policy": "lru"
    },
    "l3_unified": {
      "size_kb": 52224,
      "associativity": 12,
      "line_size_bytes": 64,
      "sets": 69632,
      "type": "unified", 
      "level": 3,
      "shared_cpu_list": "0,1",
      "write_policy": "write-back",
      "replacement_policy": "lru"
    },
    "cache_coherency": {
      "protocol": "MESI",
      "snoop_latency_cycles": 85,
      "cross_socket_latency_cycles": 140
    }
  },
  "memory_topology": {
    "total_memory_gb": 8,
    "available_memory_gb": 7.6,
    "memory_layout": [
      {
        "dimm_slot": 0,
        "size_gb": 8,
        "type": "DDR4",
        "speed_mhz": 3200,
        "manufacturer": "AWS",
        "part_number": "AWS-DDR4-3200",
        "bank_label": "BANK 0",
        "locator": "DIMM_A1"
      }
    ],
    "numa_topology": {
      "nodes": [
        {
          "node_id": 0,
          "cpus": "0-1",
          "memory_gb": 8.0,
          "distances": {
            "node_0": 10
          },
          "hugepages": {
            "2mb_total": 0,
            "2mb_free": 0,
            "1gb_total": 0,
            "1gb_free": 0
          }
        }
      ],
      "interleave_policy": "default",
      "memory_policy": "default"
    },
    "memory_controller": {
      "channels": 2,
      "dimms_per_channel": 1,
      "max_bandwidth_gb_s": 51.2,
      "ecc_enabled": true,
      "refresh_rate": "7.8us"
    }
  },
  "virtualization_details": {
    "hypervisor": "AWS Nitro",
    "cpu_steal_time_percent": 0.02,
    "memory_ballooning": false,
    "sr_iov_enabled": true,
    "pci_passthrough": false,
    "nested_virtualization": false,
    "paravirtualization": {
      "clock_source": "kvm-clock",
      "balloon_driver": "virtio_balloon"
    }
  },
  "benchmark_environment": {
    "threading_configuration": {
      "affinity_policy": "physical_cores_only",
      "cpu_pinning": {
        "enabled": true,
        "mapping": {
          "thread_0": "cpu_0",
          "thread_1": "cpu_1"
        }
      },
      "numa_binding": {
        "enabled": true,
        "policy": "bind_to_node_0",
        "memory_allocation": "local_only"
      },
      "hyperthreading_usage": {
        "enabled": true,
        "strategy": "sibling_threads_separate_workloads",
        "isolation": "core_isolation_disabled"
      }
    },
    "memory_configuration": {
      "allocation_policy": "numa_local",
      "transparent_hugepages": "madvise",
      "swap_enabled": false,
      "memory_compaction": "disabled_during_benchmark",
      "oom_killer": "disabled",
      "drop_caches_before_benchmark": true
    },
    "system_state": {
      "cpu_governor": "performance",
      "turbo_boost": "enabled",
      "c_states": "enabled",
      "address_space_layout_randomization": "disabled",
      "interrupt_affinity": "optimized",
      "kernel_preemption": "voluntary",
      "tick_rate_hz": 250
    }
  }
}
```

### 3. System Profiling Implementation

#### System Profiler (`pkg/profiling/system_profiler.go`)
```go
package profiling

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "os"
    "os/exec"
    "regexp"
    "strconv"
    "strings"
    "time"
)

type SystemProfiler struct {
    containerized bool
    privileged    bool
    hostAccess    bool
}

type SystemTopology struct {
    SchemaVersion         string                 `json:"schema_version"`
    CollectionTimestamp   time.Time             `json:"collection_timestamp"`
    InstanceMetadata      InstanceMetadata      `json:"instance_metadata"`
    CPUTopology          CPUTopology           `json:"cpu_topology"`
    CacheHierarchy       CacheHierarchy        `json:"cache_hierarchy"`
    MemoryTopology       MemoryTopology        `json:"memory_topology"`
    VirtualizationDetails VirtualizationDetails `json:"virtualization_details"`
    BenchmarkEnvironment BenchmarkEnvironment  `json:"benchmark_environment"`
}

func NewSystemProfiler() *SystemProfiler {
    return &SystemProfiler{
        containerized: isRunningInContainer(),
        privileged:    hasPrivilegedAccess(),
        hostAccess:    hasHostAccess(),
    }
}

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
            if cache.Type == "data" {
                hierarchy.L1Data = cache
            } else if cache.Type == "instruction" {
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

// Utility methods for parsing system information

func (sp *SystemProfiler) parseCPUInfo() (CPUIdentification, error) {
    file, err := os.Open("/proc/cpuinfo")
    if err != nil {
        return CPUIdentification{}, err
    }
    defer file.Close()
    
    cpuInfo := CPUIdentification{
        Features: make(map[string]bool),
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
    layout.ThreadsPerCore = layout.TotalLogicalCPUs / layout.TotalPhysicalCores
    layout.CoresPerSocket = layout.TotalPhysicalCores / layout.Sockets
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

// Additional helper methods would be implemented for:
// - getCPUFrequency() - Parse /sys/devices/system/cpu/cpu*/cpufreq/
// - parseCacheTopology() - Parse /sys/devices/system/cpu/cpu*/cache/
// - getNUMATopology() - Use numactl --hardware
// - getDIMMInfo() - Use dmidecode for memory module details
// - getThreadingConfiguration() - Configure CPU affinity and NUMA binding
// - getMemoryConfiguration() - Configure memory policies and hugepages
// - getSystemState() - Get governor, turbo, c-states, etc.
```

### 4. Enhanced Benchmark Integration

#### Benchmark Runner with System Profiling
```go
func (orchestrator *Orchestrator) RunBenchmarkWithProfiling(ctx context.Context, config BenchmarkConfig) (*EnhancedInstanceResult, error) {
    // Profile system before benchmark
    profiler := profiling.NewSystemProfiler()
    systemTopology, err := profiler.ProfileSystem(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to profile system: %w", err)
    }
    
    // Configure optimal benchmark environment
    if err := orchestrator.configureBenchmarkEnvironment(systemTopology); err != nil {
        return nil, fmt.Errorf("failed to configure benchmark environment: %w", err)
    }
    
    // Run benchmark with system monitoring
    benchmarkResult, err := orchestrator.runBenchmarkWithMonitoring(ctx, config, systemTopology)
    if err != nil {
        return nil, fmt.Errorf("benchmark execution failed: %w", err)
    }
    
    // Include system topology in results
    enhancedResult := &EnhancedInstanceResult{
        BasicResult:     benchmarkResult,
        SystemTopology:  systemTopology,
        RuntimeMetrics:  benchmarkResult.RuntimeMetrics,
    }
    
    return enhancedResult, nil
}

func (orchestrator *Orchestrator) configureBenchmarkEnvironment(topology *SystemTopology) error {
    // Configure CPU affinity for optimal performance
    if topology.CPUTopology.PhysicalLayout.HyperthreadingEnabled {
        // Pin benchmark threads to physical cores only
        return orchestrator.configureCPUAffinity(topology)
    }
    
    // Configure NUMA binding for memory-intensive benchmarks
    if len(topology.MemoryTopology.NUMATopology.Nodes) > 1 {
        return orchestrator.configureNUMABinding(topology)
    }
    
    // Configure memory policies
    return orchestrator.configureMemoryPolicies(topology)
}
```

### 5. Thread Affinity and NUMA Optimization

#### STREAM Benchmark with NUMA Awareness
```bash
#!/bin/bash
# Enhanced STREAM benchmark with system-aware configuration

# Get system topology
NUMA_NODES=$(numactl --hardware | grep "available:" | awk '{print $2}')
PHYSICAL_CORES=$(lscpu | grep "Core(s) per socket:" | awk '{print $4}')
THREADS_PER_CORE=$(lscpu | grep "Thread(s) per core:" | awk '{print $4}')
TOTAL_CORES=$((PHYSICAL_CORES * $(lscpu | grep "Socket(s):" | awk '{print $2}')))

echo "System Topology Detected:"
echo "  NUMA Nodes: $NUMA_NODES"
echo "  Physical Cores: $TOTAL_CORES"
echo "  Threads per Core: $THREADS_PER_CORE"

# Configure optimal thread affinity
if [ "$THREADS_PER_CORE" -gt 1 ]; then
    echo "Hyperthreading detected - using physical cores only"
    # Generate physical core list (every other core)
    CORE_LIST=""
    for ((i=0; i<TOTAL_CORES; i++)); do
        CORE_LIST="$CORE_LIST,$((i * THREADS_PER_CORE))"
    done
    CORE_LIST=${CORE_LIST:1}  # Remove leading comma
    
    # Run STREAM with physical core affinity
    taskset -c $CORE_LIST ./stream_benchmark
else
    echo "No hyperthreading - using all cores"
    ./stream_benchmark
fi

# Run NUMA-aware tests if multiple nodes
if [ "$NUMA_NODES" -gt 1 ]; then
    echo "Running NUMA-specific tests..."
    
    # Test local memory access
    numactl --membind=0 --cpunodebind=0 ./stream_benchmark > stream_numa_local.out
    
    # Test cross-node memory access
    numactl --membind=1 --cpunodebind=0 ./stream_benchmark > stream_numa_remote.out
    
    # Test interleaved memory access
    numactl --interleave=all ./stream_benchmark > stream_numa_interleaved.out
fi
```

## Implementation Plan

### Phase 1: Core System Profiling (Week 1)
- [ ] Implement system profiler with CPU topology detection
- [ ] Add cache hierarchy analysis using sysfs
- [ ] Integrate NUMA topology discovery
- [ ] Enhance benchmark containers with profiling tools

### Phase 2: Memory and Threading Optimization (Week 2)
- [ ] Add memory configuration analysis
- [ ] Implement thread affinity optimization
- [ ] Add NUMA-aware benchmark execution
- [ ] Create performance tuning recommendations

### Phase 3: Runtime Monitoring (Week 3)
- [ ] Add frequency monitoring during benchmarks
- [ ] Implement cache performance monitoring
- [ ] Add memory latency profiling
- [ ] Create dynamic performance adjustment

### Phase 4: Data Integration (Week 4)
- [ ] Update Git data schema with system topology
- [ ] Enhance GitHub Pages with system details
- [ ] Add system-aware ComputeCompass integration
- [ ] Create performance correlation analysis

This comprehensive system profiling will provide the detailed hardware information needed for precise performance analysis and optimal workload placement recommendations.