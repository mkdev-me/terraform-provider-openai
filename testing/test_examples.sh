#!/bin/bash
#
# test_examples.sh - Test OpenAI Terraform provider examples
#
# Usage:
#   ./test_examples.sh [command] [options]
#
# Commands:
#   plan [example]    - Run terraform plan (default: all examples)
#   apply [example]   - Run terraform apply with cleanup (default: all examples)
#   quick             - Quick test with minimal resources
#   cleanup           - Remove all terraform artifacts
#
# Examples:
#   ./test_examples.sh plan              # Plan all examples
#   ./test_examples.sh plan image        # Plan only image example
#   ./test_examples.sh apply chat        # Apply and destroy chat example
#   ./test_examples.sh apply             # Apply all examples (WARNING: costs!)
#   ./test_examples.sh quick             # Quick verification test
#   ./test_examples.sh cleanup           # Clean all terraform files
#
# Environment Variables Required:
#   OPENAI_API_KEY    - Project API key (sk-proj-...)
#   OPENAI_ADMIN_KEY  - Admin API key (sk-admin-...) for organization resources
#

set -euo pipefail

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
EXAMPLES_DIR="$PROJECT_ROOT/examples"

# Check environment variables
check_env() {
    if [ -z "${OPENAI_API_KEY:-}" ]; then
        echo -e "${RED}Error: OPENAI_API_KEY not set${NC}"
        echo ""
        echo "Please set environment variables:"
        echo "  export OPENAI_API_KEY=sk-proj-..."
        echo "  export OPENAI_ADMIN_KEY=sk-admin-..."
        echo ""
        echo "Or source from .env file:"
        echo "  source .env"
        exit 1
    fi
    
    echo "Environment:"
    echo "  API Key: ${OPENAI_API_KEY:0:20}..."
    [ -n "${OPENAI_ADMIN_KEY:-}" ] && echo "  Admin Key: ${OPENAI_ADMIN_KEY:0:20}..."
    echo ""
}

# Test a single example with plan only
test_plan() {
    local example=$1
    local example_path="$EXAMPLES_DIR/$example"
    
    if [ ! -d "$example_path" ]; then
        echo -e "${RED}Error: Example '$example' not found${NC}"
        return 1
    fi
    
    echo -e "${YELLOW}Testing: $example${NC}"
    cd "$example_path"
    
    # Clean previous state
    rm -rf .terraform* terraform.tfstate* tfplan 2>/dev/null || true
    
    # Initialize
    if ! terraform init -upgrade > /tmp/init_$example.log 2>&1; then
        echo -e "${RED}✗ Init failed${NC}"
        grep "Error:" /tmp/init_$example.log | head -5
        return 1
    fi
    
    # Plan
    if ! terraform plan -input=false > /tmp/plan_$example.log 2>&1; then
        echo -e "${RED}✗ Plan failed${NC}"
        grep "Error:" /tmp/plan_$example.log | head -5
        return 1
    fi
    
    echo -e "${GREEN}✓ Plan successful${NC}"
    return 0
}

# Test a single example with apply and destroy
test_apply() {
    local example=$1
    local example_path="$EXAMPLES_DIR/$example"
    
    if [ ! -d "$example_path" ]; then
        echo -e "${RED}Error: Example '$example' not found${NC}"
        return 1
    fi
    
    echo -e "${YELLOW}Testing: $example (with apply)${NC}"
    cd "$example_path"
    
    # Clean previous state
    rm -rf .terraform* terraform.tfstate* tfplan 2>/dev/null || true
    
    # Initialize
    echo "  Initializing..."
    if ! terraform init -upgrade > /tmp/init_$example.log 2>&1; then
        echo -e "${RED}  ✗ Init failed${NC}"
        return 1
    fi
    
    # Plan
    echo "  Planning..."
    if ! terraform plan -input=false -out=tfplan > /tmp/plan_$example.log 2>&1; then
        echo -e "${RED}  ✗ Plan failed${NC}"
        return 1
    fi
    
    # Apply
    echo "  Applying..."
    if ! terraform apply -auto-approve tfplan > /tmp/apply_$example.log 2>&1; then
        echo -e "${RED}  ✗ Apply failed${NC}"
        # Try to destroy any partial resources
        terraform destroy -auto-approve > /dev/null 2>&1 || true
        return 1
    fi
    
    echo -e "${GREEN}  ✓ Apply successful${NC}"
    
    # Show outputs (if any)
    if terraform output -json 2>/dev/null | jq -e '. | length > 0' > /dev/null; then
        echo "  Outputs:"
        terraform output 2>/dev/null | head -5 | sed 's/^/    /'
    fi
    
    # Destroy
    echo "  Destroying..."
    if ! terraform destroy -auto-approve > /tmp/destroy_$example.log 2>&1; then
        echo -e "${RED}  ✗ Destroy failed - resources may still exist!${NC}"
        return 1
    fi
    
    echo -e "${GREEN}  ✓ Cleanup successful${NC}"
    
    # Final cleanup
    rm -rf .terraform* terraform.tfstate* tfplan 2>/dev/null || true
    return 0
}

# Quick verification test
quick_test() {
    echo -e "${BLUE}Running quick verification test...${NC}"
    
    local TEST_DIR="$SCRIPT_DIR/temp-quick-test-$$"
    mkdir -p "$TEST_DIR"
    cd "$TEST_DIR"
    
    cat > main.tf << 'EOF'
terraform {
  required_providers {
    openai = {
      source  = "mkdev-me/openai"
      version = "1.0.0"
    }
  }
}

provider "openai" {}

# Test project key
resource "openai_embedding" "test" {
  model = "text-embedding-ada-002"
  input = jsonencode(["Quick test"])
}

# Test admin key (if available)
resource "openai_project" "test" {
  count = var.has_admin_key ? 1 : 0
  name  = "Quick test project"
}

variable "has_admin_key" {
  type    = bool
  default = false
}

output "test_passed" {
  value = "✓ Provider working correctly"
}
EOF

    # Run test
    terraform init -upgrade > /dev/null 2>&1
    
    local admin_flag=""
    if [ -n "${OPENAI_ADMIN_KEY:-}" ]; then
        admin_flag="-var=has_admin_key=true"
    fi
    
    if terraform apply -auto-approve $admin_flag > /tmp/quick-test-apply.log 2>&1; then
        echo -e "${GREEN}✓ Quick test passed${NC}"
        terraform destroy -auto-approve $admin_flag > /dev/null 2>&1
    else
        echo -e "${RED}✗ Quick test failed${NC}"
        echo "Error details:"
        tail -20 /tmp/quick-test-apply.log | grep -E "(Error:|error:|failed)" || tail -10 /tmp/quick-test-apply.log
        cd "$SCRIPT_DIR"
        rm -rf "$TEST_DIR"
        return 1
    fi
    
    cd "$SCRIPT_DIR"
    rm -rf "$TEST_DIR"
    return 0
}

# Clean all terraform artifacts
cleanup_all() {
    echo -e "${BLUE}Cleaning all terraform artifacts...${NC}"
    
    local cleaned=0
    for example_dir in "$EXAMPLES_DIR"/*; do
        if [ -d "$example_dir" ]; then
            if ls "$example_dir"/.terraform* "$example_dir"/terraform.tfstate* "$example_dir"/tfplan 2>/dev/null | grep -q .; then
                example=$(basename "$example_dir")
                echo "  Cleaning $example"
                rm -rf "$example_dir"/.terraform* "$example_dir"/terraform.tfstate* "$example_dir"/tfplan 2>/dev/null || true
                ((cleaned++))
            fi
        fi
    done
    
    # Clean temp files
    rm -f /tmp/init_*.log /tmp/plan_*.log /tmp/apply_*.log /tmp/destroy_*.log 2>/dev/null || true
    
    echo -e "${GREEN}✓ Cleaned $cleaned directories${NC}"
}

# Main logic
main() {
    local command=${1:-plan}
    local target=${2:-all}
    
    case "$command" in
        plan)
            check_env
            if [ "$target" = "all" ]; then
                echo -e "${BLUE}Testing all examples (plan only)${NC}"
                echo ""
                
                local passed=0
                local failed=0
                
                for example_dir in "$EXAMPLES_DIR"/*; do
                    if [ -d "$example_dir" ] && [ -f "$example_dir/main.tf" ]; then
                        example=$(basename "$example_dir")
                        if test_plan "$example"; then
                            ((passed++))
                        else
                            ((failed++))
                        fi
                    fi
                done
                
                echo ""
                echo -e "${BLUE}Summary:${NC} ${GREEN}$passed passed${NC}, ${RED}$failed failed${NC}"
            else
                test_plan "$target"
            fi
            ;;
            
        apply)
            check_env
            if [ "$target" = "all" ]; then
                echo -e "${YELLOW}WARNING: This will create real resources and may incur costs!${NC}"
                read -p "Continue? (y/N) " -n 1 -r
                echo
                if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                    echo "Cancelled."
                    exit 0
                fi
                
                echo -e "${BLUE}Testing all examples (with apply)${NC}"
                echo ""
                
                local passed=0
                local failed=0
                
                for example_dir in "$EXAMPLES_DIR"/*; do
                    if [ -d "$example_dir" ] && [ -f "$example_dir/main.tf" ]; then
                        example=$(basename "$example_dir")
                        if test_apply "$example"; then
                            ((passed++))
                        else
                            ((failed++))
                        fi
                        echo ""
                    fi
                done
                
                echo -e "${BLUE}Summary:${NC} ${GREEN}$passed passed${NC}, ${RED}$failed failed${NC}"
            else
                test_apply "$target"
            fi
            ;;
            
        quick)
            check_env
            quick_test
            ;;
            
        cleanup)
            cleanup_all
            ;;
            
        *)
            echo "Usage: $0 [plan|apply|quick|cleanup] [example_name|all]"
            echo ""
            echo "Commands:"
            echo "  plan [example]   - Run terraform plan"
            echo "  apply [example]  - Run terraform apply with cleanup"
            echo "  quick            - Quick verification test"
            echo "  cleanup          - Remove all terraform artifacts"
            exit 1
            ;;
    esac
}

main "$@"