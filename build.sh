#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default settings
BUILD_ALL=false
PLATFORMS=()
OUTPUT_DIR="dist"
BINARY_NAME="ccc"

# Available platforms (using plain variables for bash compatibility)
PLATFORMS_INFO=(
    "darwin-amd64|macOS x86_64"
    "darwin-arm64|macOS ARM64 (Apple Silicon)"
    "linux-amd64|Linux x86_64"
    "linux-arm64|Linux ARM64"
    "windows-amd64|Windows x86_64"
)

print_help() {
    echo -e "${BLUE}Usage:${NC} $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -a, --all           Build for all supported platforms"
    echo "  -p, --platforms     Build for specific platforms (comma-separated)"
    echo "                      Available: darwin-amd64, darwin-arm64, linux-amd64, linux-arm64, windows-amd64"
    echo "  -o, --output        Output directory (default: dist)"
    echo "  -n, --name          Binary name (default: ccc)"
    echo "  -h, --help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                          # Build for current platform only"
    echo "  $0 --all                    # Build for all platforms"
    echo "  $0 -p darwin-arm64,linux-amd64  # Build for specific platforms"
    echo "  $0 -a -o ./build            # Build all platforms to ./build directory"
}

# Detect current platform
detect_current_platform() {
    local os=$(go env GOOS)
    local arch=$(go env GOARCH)
    echo "${os}-${arch}"
}

# Build for a specific platform
build_platform() {
    local os=$1
    local arch=$2
    local platform_key="${os}-${arch}"
    local output_name="${BINARY_NAME}-${platform_key}"
    local output_path="${OUTPUT_DIR}/${output_name}"

    # Add .exe extension for Windows
    if [ "$os" = "windows" ]; then
        output_path="${output_path}.exe"
    fi

    echo -e "${YELLOW}Building for ${BLUE}${os}/${arch}${NC}..."

    # Create output directory
    mkdir -p "${OUTPUT_DIR}"

    # Build
    CGO_ENABLED=0 GOOS="${os}" GOARCH="${arch}" go build -ldflags="-s -w" -o "${output_path}" main.go

    # Make executable (not needed for Windows)
    if [ "$os" != "windows" ]; then
        chmod +x "${output_path}"
    fi

    echo -e "${GREEN}  ✓${NC} ${output_path}"
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -a|--all)
            BUILD_ALL=true
            shift
            ;;
        -p|--platforms)
            IFS=',' read -ra PLATFORMS <<< "$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -n|--name)
            BINARY_NAME="$2"
            shift 2
            ;;
        -h|--help)
            print_help
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            print_help
            exit 1
            ;;
    esac
done

# Main logic
echo -e "${BLUE}Building ccc command line tool...${NC}"
echo ""

# If no platforms specified, build current platform only
if [ "$BUILD_ALL" = false ] && [ ${#PLATFORMS[@]} -eq 0 ]; then
    current_platform=$(detect_current_platform)
    echo -e "${YELLOW}Building for current platform: ${BLUE}${current_platform}${NC}"
    build_platform "$(echo "$current_platform" | cut -d'-' -f1)" "$(echo "$current_platform" | cut -d'-' -f2)"
else
    # Build for specified or all platforms
    if [ "$BUILD_ALL" = true ]; then
        PLATFORMS=("darwin-amd64" "darwin-arm64" "linux-amd64" "linux-arm64" "windows-amd64")
    fi

    echo -e "${YELLOW}Building for ${#PLATFORMS[@]} platform(s):${NC}"
    for platform in "${PLATFORMS[@]}"; do
        os=$(echo "$platform" | cut -d'-' -f1)
        arch=$(echo "$platform" | cut -d'-' -f2)

        # Validate platform
        valid=false
        for info in "${PLATFORMS_INFO[@]}"; do
            if [[ "$info" == "${platform}"* ]]; then
                valid=true
                break
            fi
        done
        if [ "$valid" = false ]; then
            echo -e "${RED}  ✗ Invalid platform: ${platform}${NC}"
            echo "  Available platforms: darwin-amd64, darwin-arm64, linux-amd64, linux-arm64, windows-amd64"
            exit 1
        fi

        build_platform "$os" "$arch"
    done
fi

echo ""
echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "${BLUE}Output directory: ${OUTPUT_DIR}${NC}"
echo ""
echo "Contents:"
ls -lh "${OUTPUT_DIR}"/ccc-* 2>/dev/null || echo "  (no binaries found)"
echo ""
echo "To install a specific binary, run:"
echo "  sudo cp ${OUTPUT_DIR}/ccc-<platform> /usr/local/bin/ccc"
echo "Example:"
echo "  sudo cp ${OUTPUT_DIR}/ccc-darwin-amd64 /usr/local/bin/ccc"
