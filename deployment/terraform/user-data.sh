#!/bin/bash
# AWS Instance Benchmarks - EC2 User Data Script
# Prepares instances for reproducible benchmark execution

set -euo pipefail

# Configuration
S3_BUCKET="${s3_bucket}"
AWS_REGION="${region}"
BENCHMARK_HOME="/opt/aws-instance-benchmarks"
LOG_GROUP="/aws/benchmark/setup"

# Logging function
log() {
    echo "$(date -Iseconds) $*" | tee -a /var/log/benchmark-setup.log
    aws logs put-log-events \
        --region "$AWS_REGION" \
        --log-group-name "$LOG_GROUP" \
        --log-stream-name "$(ec2-metadata --instance-id | cut -d' ' -f2)" \
        --log-events "timestamp=$(date +%s000),message=$*" \
        --region "$AWS_REGION" || true
}

# System update and basic tools
log "Starting benchmark instance setup"
yum update -y
yum install -y \
    git \
    docker \
    htop \
    iotop \
    iftop \
    perf \
    sysstat \
    numactl \
    hwloc \
    lshw \
    dmidecode \
    pciutils \
    gcc \
    gcc-c++ \
    gfortran \
    make \
    cmake \
    autoconf \
    automake \
    libtool \
    pkgconfig \
    python3 \
    python3-pip \
    python3-devel \
    R \
    openmpi \
    openmpi-devel \
    blas \
    blas-devel \
    lapack \
    lapack-devel \
    fftw \
    fftw-devel \
    hdf5 \
    hdf5-devel \
    netcdf \
    netcdf-devel

# Enable and start services
systemctl enable docker
systemctl start docker
usermod -aG docker ec2-user

# Install AWS CLI v2
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
rm -rf aws awscliv2.zip

# Install CloudWatch agent
wget https://s3.amazonaws.com/amazoncloudwatch-agent/amazon_linux/amd64/latest/amazon-cloudwatch-agent.rpm
rpm -U ./amazon-cloudwatch-agent.rpm
rm -f ./amazon-cloudwatch-agent.rpm

# Create benchmark directories
mkdir -p "$BENCHMARK_HOME"
mkdir -p /var/log/benchmarks
mkdir -p /tmp/benchmark-results
chown ec2-user:ec2-user "$BENCHMARK_HOME" /var/log/benchmarks /tmp/benchmark-results

# Install Python packages
pip3 install --upgrade pip
pip3 install \
    numpy \
    scipy \
    pandas \
    matplotlib \
    seaborn \
    scikit-learn \
    tensorflow \
    torch \
    transformers \
    boto3 \
    awscli \
    psutil \
    py-cpuinfo \
    GPUtil \
    h5py \
    netcdf4 \
    xarray \
    mpi4py \
    dask \
    joblib

# Install R packages
R -e "install.packages(c('devtools', 'BiocManager', 'data.table', 'dplyr', 'ggplot2', 'parallel', 'doParallel', 'foreach'), repos='http://cran.us.r-project.org')"

# Clone benchmark repository
cd "$BENCHMARK_HOME"
git clone https://github.com/scttfrdmn/aws-instance-benchmarks.git .
chown -R ec2-user:ec2-user .

# Build containers
log "Building benchmark containers"
cd "$BENCHMARK_HOME"
sudo -u ec2-user bash ./scripts/build-containers.sh

# System tuning for benchmarks
log "Applying system tuning for benchmarks"

# Disable CPU frequency scaling
echo performance > /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor || true

# Disable CPU idle states
for state in /sys/devices/system/cpu/cpu*/cpuidle/state*/disable; do
    echo 1 > "$state" 2>/dev/null || true
done

# Set CPU affinity and NUMA policy
echo always > /sys/kernel/mm/transparent_hugepage/enabled

# Network optimizations
echo 'net.core.rmem_default = 262144' >> /etc/sysctl.conf
echo 'net.core.rmem_max = 16777216' >> /etc/sysctl.conf
echo 'net.core.wmem_default = 262144' >> /etc/sysctl.conf
echo 'net.core.wmem_max = 16777216' >> /etc/sysctl.conf
sysctl -p

# Configure CloudWatch agent
cat > /opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json << 'EOF'
{
    "metrics": {
        "namespace": "AWS/InstanceBenchmarks",
        "metrics_collected": {
            "cpu": {
                "measurement": ["cpu_usage_idle", "cpu_usage_iowait", "cpu_usage_user", "cpu_usage_system"],
                "metrics_collection_interval": 60
            },
            "disk": {
                "measurement": ["used_percent"],
                "metrics_collection_interval": 60,
                "resources": ["*"]
            },
            "mem": {
                "measurement": ["mem_used_percent"],
                "metrics_collection_interval": 60
            },
            "netstat": {
                "measurement": ["tcp_established", "tcp_time_wait"],
                "metrics_collection_interval": 60
            }
        }
    },
    "logs": {
        "logs_collected": {
            "files": {
                "collect_list": [
                    {
                        "file_path": "/var/log/benchmarks/*.log",
                        "log_group_name": "/aws/benchmark/execution",
                        "log_stream_name": "{instance_id}"
                    }
                ]
            }
        }
    }
}
EOF

/opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl \
    -a fetch-config \
    -m ec2 \
    -c file:/opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json \
    -s

# Create benchmark execution script
cat > /usr/local/bin/run-benchmark.sh << 'EOF'
#!/bin/bash
set -euo pipefail

SUITE=${1:-"all"}
INSTANCE_TYPE=$(ec2-metadata --instance-type | cut -d' ' -f2)
INSTANCE_ID=$(ec2-metadata --instance-id | cut -d' ' -f2)
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

export BENCHMARK_HOME="/opt/aws-instance-benchmarks"
export RESULTS_DIR="/tmp/benchmark-results"
export LOG_FILE="/var/log/benchmarks/benchmark-${SUITE}-${TIMESTAMP}.log"

# Setup logging
mkdir -p "$(dirname "$LOG_FILE")"
exec 1> >(tee -a "$LOG_FILE")
exec 2>&1

echo "Starting benchmark execution: $SUITE on $INSTANCE_TYPE"
echo "Instance ID: $INSTANCE_ID"
echo "Timestamp: $TIMESTAMP"

cd "$BENCHMARK_HOME"

# System information collection
echo "=== System Information ==="
uname -a
lscpu
lsmem
numactl --hardware
cat /proc/meminfo
lspci

echo "=== Starting Benchmarks ==="

# Run the appropriate benchmark suite
case "$SUITE" in
    "genomics")
        ./scripts/benchmark-runners/run-genomics-benchmarks.sh
        ;;
    "ml")
        ./scripts/benchmark-runners/run-ml-benchmarks.sh
        ;;
    "climate")
        ./scripts/benchmark-runners/run-climate-benchmarks.sh
        ;;
    "chemistry")
        ./scripts/benchmark-runners/run-chemistry-benchmarks.sh
        ;;
    "all")
        ./scripts/benchmark-runners/run-all-benchmarks.sh
        ;;
    *)
        echo "Unknown benchmark suite: $SUITE"
        echo "Available suites: genomics, ml, climate, chemistry, all"
        exit 1
        ;;
esac

echo "=== Uploading Results ==="
aws s3 sync "$RESULTS_DIR" "s3://${S3_BUCKET}/results/$(date +%Y-%m-%d)/" --region "${AWS_REGION}"

echo "Benchmark execution completed successfully"
EOF

chmod +x /usr/local/bin/run-benchmark.sh

# Create system monitoring script
cat > /usr/local/bin/monitor-benchmark.sh << 'EOF'
#!/bin/bash
# Continuous monitoring during benchmark execution

INTERVAL=${1:-5}
DURATION=${2:-3600}  # 1 hour default
OUTPUT_FILE="/tmp/benchmark-monitoring-$(date +%Y%m%d-%H%M%S).log"

echo "Starting system monitoring (interval: ${INTERVAL}s, duration: ${DURATION}s)"
echo "Output file: $OUTPUT_FILE"

END_TIME=$(($(date +%s) + DURATION))

while [ $(date +%s) -lt $END_TIME ]; do
    {
        echo "=== $(date -Iseconds) ==="
        echo "CPU Usage:"
        top -bn1 | grep "^%Cpu" | head -1
        
        echo "Memory Usage:"
        free -h
        
        echo "Network I/O:"
        cat /proc/net/dev | grep -E "(eth0|ens)" | head -1
        
        echo "Disk I/O:"
        iostat -x 1 1 | tail -n +4
        
        echo "Load Average:"
        uptime
        
        echo ""
    } >> "$OUTPUT_FILE"
    
    sleep "$INTERVAL"
done

echo "Monitoring completed. Log saved to: $OUTPUT_FILE"

# Upload monitoring data to S3
aws s3 cp "$OUTPUT_FILE" "s3://${S3_BUCKET}/monitoring/" --region "${AWS_REGION}"
EOF

chmod +x /usr/local/bin/monitor-benchmark.sh

# Signal completion
log "Benchmark instance setup completed successfully"
/opt/aws/bin/cfn-signal -e $? --stack $(ec2-metadata --instance-id | cut -d' ' -f2) --resource AutoScalingGroup --region "$AWS_REGION" || true

log "Instance is ready for benchmark execution"
echo "Instance setup complete. Ready for benchmarking." > /tmp/setup-complete