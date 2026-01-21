#!/bin/bash
# Unified Demo Runner
# Runs all NeuronIP demos in sequence or individually

set -e

API_URL="${API_URL:-http://localhost:8082/api/v1}"
API_KEY="${API_KEY:-your-api-key-here}"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEMO_DIR="${SCRIPT_DIR}"

usage() {
  echo "Usage: $0 [OPTIONS] [DEMO_NUMBER]"
  echo ""
  echo "Options:"
  echo "  -a, --all          Run all demos in sequence"
  echo "  -1, --support      Run Demo 1: Support Memory Hub"
  echo "  -2, --warehouse    Run Demo 2: Warehouse Q&A"
  echo "  -3, --semantic     Run Demo 3: Semantic Search"
  echo "  -h, --help         Show this help message"
  echo ""
  echo "Environment Variables:"
  echo "  API_URL           API endpoint (default: http://localhost:8082/api/v1)"
  echo "  API_KEY           API key for authentication"
  echo ""
  echo "Examples:"
  echo "  $0 --all                    # Run all demos"
  echo "  $0 --support                 # Run demo 1 only"
  echo "  $0 1                        # Run demo 1 (short form)"
  echo "  API_KEY=xxx $0 --all        # Run with custom API key"
}

check_api() {
  echo -e "${BLUE}Checking API connectivity...${NC}"
  if curl -s -f "${API_URL%/api/v1}/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ API is reachable${NC}"
    return 0
  else
    echo -e "${RED}✗ API is not reachable at ${API_URL}${NC}"
    echo "Please ensure NeuronIP is running and API_URL is correct"
    return 1
  fi
}

run_demo() {
  local demo_num=$1
  local demo_file=""
  local demo_name=""
  
  case $demo_num in
    1)
      demo_file="${DEMO_DIR}/demo-1-support-memory.sh"
      demo_name="Support Memory Hub"
      ;;
    2)
      demo_file="${DEMO_DIR}/demo-2-warehouse-qa.sh"
      demo_name="Warehouse Q&A"
      ;;
    3)
      demo_file="${DEMO_DIR}/demo-3-semantic-search.sh"
      demo_name="Semantic Search"
      ;;
    *)
      echo -e "${RED}Invalid demo number: ${demo_num}${NC}"
      echo "Valid options: 1, 2, 3"
      return 1
      ;;
  esac
  
  if [ ! -f "$demo_file" ]; then
    echo -e "${RED}Demo file not found: ${demo_file}${NC}"
    return 1
  fi
  
  if [ ! -x "$demo_file" ]; then
    chmod +x "$demo_file"
  fi
  
  echo ""
  echo "=========================================="
  echo "Running Demo ${demo_num}: ${demo_name}"
  echo "=========================================="
  echo ""
  
  export API_URL API_KEY
  bash "$demo_file"
  
  local exit_code=$?
  if [ $exit_code -eq 0 ]; then
    echo -e "${GREEN}✓ Demo ${demo_num} completed successfully${NC}"
  else
    echo -e "${RED}✗ Demo ${demo_num} failed with exit code ${exit_code}${NC}"
  fi
  
  return $exit_code
}

run_all_demos() {
  echo "=========================================="
  echo "NeuronIP Demo Suite"
  echo "Running All Demos"
  echo "=========================================="
  echo ""
  
  if ! check_api; then
    exit 1
  fi
  
  echo ""
  local failed=0
  
  for i in 1 2 3; do
    if ! run_demo $i; then
      failed=$((failed + 1))
    fi
    echo ""
    if [ $i -lt 3 ]; then
      echo -e "${YELLOW}Waiting 5 seconds before next demo...${NC}"
      sleep 5
      echo ""
    fi
  done
  
  echo "=========================================="
  echo "Demo Suite Complete"
  echo "=========================================="
  echo ""
  
  if [ $failed -eq 0 ]; then
    echo -e "${GREEN}✓ All demos completed successfully${NC}"
    return 0
  else
    echo -e "${RED}✗ ${failed} demo(s) failed${NC}"
    return 1
  fi
}

# Parse arguments
if [ $# -eq 0 ]; then
  usage
  exit 0
fi

case "$1" in
  -a|--all)
    run_all_demos
    ;;
  -1|--support)
    check_api && run_demo 1
    ;;
  -2|--warehouse)
    check_api && run_demo 2
    ;;
  -3|--semantic)
    check_api && run_demo 3
    ;;
  -h|--help)
    usage
    ;;
  1|2|3)
    check_api && run_demo "$1"
    ;;
  *)
    echo -e "${RED}Unknown option: $1${NC}"
    echo ""
    usage
    exit 1
    ;;
esac
