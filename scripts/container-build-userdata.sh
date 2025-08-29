#!/bin/bash

# EC2 User Data Script for Container Building Instances
# Prepares EC2 instance for native container building with Docker and development tools

set -euo pipefail

# Log setup
exec > >(tee /var/log/container-build-setup.log)
exec 2>&1

echo "=== Container Build Instance Setup Started ==="
echo "Timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo "Instance ID: $(curl -s http://169.254.169.254/latest/meta-data/instance-id)"
echo "Instance Type: $(curl -s http://169.254.169.254/latest/meta-data/instance-type)"
echo "Architecture: $(uname -m)"

# Update system
echo "Updating system packages..."
yum update -y

# Install Docker and development tools
echo "Installing Docker and development tools..."
yum install -y \
    docker \
    git \
    wget \
    curl \
    gcc \
    gcc-c++ \
    make \
    cmake \
    python3 \
    python3-pip

# Start and enable Docker
echo "Starting Docker service..."
systemctl start docker
systemctl enable docker

# Add ec2-user to docker group
usermod -a -G docker ec2-user

# Install Docker Compose
echo "Installing Docker Compose..."
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" \
    -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Install AWS CLI v2
echo "Installing AWS CLI v2..."
cd /tmp
curl "https://awscli.amazonaws.com/awscli-exe-linux-$(uname -m).zip" -o "awscliv2.zip"
unzip awscliv2.zip
./aws/install

# Configure git for container builds
echo "Configuring git..."
git config --global user.name "AWS Container Builder"
git config --global user.email "container-builder@aws.local"
git config --global init.defaultBranch main

# Create workspace directory
echo "Creating build workspace..."
mkdir -p /opt/container-build
chown ec2-user:ec2-user /opt/container-build

# Pre-pull common base images
echo "Pre-pulling common base images..."
docker pull ubuntu:24.04
docker pull ubuntu:22.04
docker pull amazonlinux:2023

# Install architecture-specific tools
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        echo "Installing x86_64 specific tools..."
        # Install Intel tools if needed for Intel-optimized builds
        if curl -s --max-time 5 "https://apt.repos.intel.com/intel-gpg-keys/GPG-PUB-KEY-INTEL-SW-PRODUCTS.PUB" >/dev/null; then
            echo "Intel repository accessible, preparing for potential Intel builds"
        fi
        ;;
    aarch64)
        echo "Installing ARM64 specific tools..."
        # Install ARM-specific development packages
        yum install -y gcc-aarch64-linux-gnu
        ;;
esac

# Create performance monitoring script
cat > /opt/container-build/monitor.sh << 'EOF'
#!/bin/bash
# Monitor system resources during container build
while true; do
    echo "$(date): CPU: $(top -bn1 | grep "Cpu(s)" | awk '{print $2}') | Memory: $(free -h | grep Mem | awk '{print $3"/"$2}') | Disk: $(df -h / | tail -1 | awk '{print $5}')"
    sleep 30
done
EOF
chmod +x /opt/container-build/monitor.sh

# Signal CloudFormation/SSM that setup is complete
echo "Container build instance setup complete"
echo "Docker version: $(docker --version)"
echo "Git version: $(git --version)"
echo "AWS CLI version: $(aws --version)"
echo "=== Container Build Instance Setup Complete ==="

# Touch completion file
touch /opt/container-build/setup-complete