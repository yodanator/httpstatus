#!/bin/bash

# Cross-platform build script with ARM and FreeBSD support
set -e

# Initialize Go module if needed
if [ ! -f go.mod ]; then
    echo "Initializing Go module..."
    go mod init httpstatus
fi

# Install dependencies
go mod tidy

VERSION=${1:-1.0.0}
OUT_DIR="dist"
mkdir -p $OUT_DIR

# Install UPX if available
install_upx() {
    if ! command -v upx &> /dev/null; then
        echo "Installing UPX..."
        case "$(uname -s)" in
            Linux*) 
                if command -v apt &> /dev/null; then
                    sudo apt update && sudo apt install -y upx
                elif command -v yum &> /dev/null; then
                    sudo yum install -y upx
                else
                    echo "Warning: Package manager not found. UPX won't be installed."
                fi
                ;;
            Darwin*) 
                if command -v brew &> /dev/null; then
                    brew install upx
                else
                    echo "Warning: Homebrew not found. UPX won't be installed."
                fi
                ;;
        esac
    fi
}

# Supported platforms
PLATFORMS=(
    "darwin/amd64"        # Intel Mac
    "darwin/arm64"        # Apple Silicon
    "linux/amd64"         # 64-bit Linux
    "linux/386"           # 32-bit Linux
    "linux/arm"           # ARMv6 (Raspberry Pi Zero)
    "linux/arm64"         # ARM64 (Raspberry Pi 3/4)
    "linux/arm/v7"        # ARMv7 (Raspberry Pi 2/3)
    "windows/amd64"       # 64-bit Windows
    "windows/386"         # 32-bit Windows
    "windows/arm64"       # ARM64 Windows
    "freebsd/amd64"       # FreeBSD
)

build_for_platform() {
    local os=$1
    local arch=$2
    local version=$3
    
    # Handle ARM variants
    local goarm=""
    local arm_suffix=""
    if [[ $arch == arm* ]]; then
        if [[ $arch == "arm" ]]; then
            goarm=6
            arm_suffix="v6"
        elif [[ $arch == "arm/v7" ]]; then
            goarm=7
            arch="arm"
            arm_suffix="v7"
        fi
    fi
    
    # Set output name
    local binary_name="httpstatus"
    if [ "$os" = "windows" ]; then
        binary_name="$binary_name.exe"
    fi
    
    echo "Building $os/$arch v$version..."
    
    # Build flags
    local build_cmd="CGO_ENABLED=0 GOOS=$os GOARCH=$arch"
    if [ -n "$goarm" ]; then
        build_cmd+=" GOARM=$goarm"
    fi
    build_cmd+=" go build -ldflags=\"-s -w\" -trimpath -o $OUT_DIR/$binary_name"
    
    # Execute build
    eval $build_cmd
    
    # Create archive name
    local archive_name="httpstatus-${os}-${arch}"
    if [ -n "$arm_suffix" ]; then
        archive_name="httpstatus-${os}-${arm_suffix}"
    fi
    archive_name+="-v$version"
    
    # Skip UPX for unsupported platforms
    local skip_upx=0
    case "$os/$arch" in
        darwin/*|windows/arm64|freebsd/*)
            skip_upx=1
            echo "Skipping UPX for $os/$arch"
            ;;
    esac
    
    # Apply UPX compression if available and supported
    if command -v upx &> /dev/null && [ $skip_upx -eq 0 ]; then
        echo "Compressing with UPX..."
        upx --best -q "$OUT_DIR/$binary_name" || true
    fi
    
    # Create archive
    if [ "$os" = "windows" ]; then
        zip -j "$OUT_DIR/$archive_name.zip" "$OUT_DIR/$binary_name" > /dev/null
    else
        tar -czf "$OUT_DIR/$archive_name.tar.gz" -C "$OUT_DIR" "$binary_name" > /dev/null
    fi
    
    # Cleanup binary
    rm "$OUT_DIR/$binary_name"
}

# Main build process
install_upx
for platform in "${PLATFORMS[@]}"; do
    platform_split=(${platform//\// })
    os=${platform_split[0]}
    arch=${platform_split[1]}
    
    # Handle ARM variants
    if [[ ${#platform_split[@]} -eq 3 ]]; then
        arch="$arch/${platform_split[2]}"
    fi
    
    build_for_platform $os $arch $VERSION
done

# Generate checksums
cd $OUT_DIR
sha256sum * > SHA256SUMS
cd ..

echo -e "\nBuild complete! Results in $OUT_DIR/:"
ls -lh $OUT_DIR
