#!/bin/bash
#
# test_examples.sh - Test OpenAI Terraform provider examples
#
# Usage:
#   ./test_examples.sh [command] [options]
#
# Commands:
#   plan [type/example]   - Run terraform plan (default: all examples)
#   apply [type/example]  - Run terraform apply with cleanup (default: all examples)
#   quick                 - Quick test with minimal resources
#   cleanup               - Remove all terraform artifacts
#
# Examples:
#   ./test_examples.sh plan                          # Plan all examples
#   ./test_examples.sh plan resources/               # Plan all resource examples
#   ./test_examples.sh plan resources/openai_image   # Plan only image example
#   ./test_examples.sh apply data-sources/           # Apply all data source examples
#   ./test_examples.sh apply                         # Apply all examples (WARNING: costs!)
#   ./test_examples.sh quick                         # Quick verification test
#   ./test_examples.sh cleanup                       # Clean all terraform files
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
    if ! terraform init -upgrade > /tmp/init_$(echo $example | tr '/' '_').log 2>&1; then
        echo -e "${RED}✗ Init failed${NC}"
        grep "Error:" /tmp/init_$(echo $example | tr '/' '_').log | head -5
        return 1
    fi
    
    # Set variables
    local tf_vars=""
    if [ -f "variables.tf" ]; then
        tf_vars="-var=openai_api_key=${OPENAI_API_KEY}"
        if grep -q "openai_admin_key" variables.tf 2>/dev/null && [ -n "${OPENAI_ADMIN_KEY:-}" ]; then
            tf_vars="$tf_vars -var=openai_admin_key=${OPENAI_ADMIN_KEY}"
        fi
        if grep -q "organization_id" variables.tf 2>/dev/null && [ -n "${OPENAI_ORGANIZATION_ID:-}" ]; then
            tf_vars="$tf_vars -var=organization_id=${OPENAI_ORGANIZATION_ID}"
        fi
    fi
    
    # Plan
    if ! terraform plan -input=false $tf_vars > /tmp/plan_$(echo $example | tr '/' '_').log 2>&1; then
        echo -e "${RED}✗ Plan failed${NC}"
        grep "Error:" /tmp/plan_$(echo $example | tr '/' '_').log | head -5
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
    if ! terraform init -upgrade > /tmp/init_$(echo $example | tr '/' '_').log 2>&1; then
        echo -e "${RED}  ✗ Init failed${NC}"
        return 1
    fi
    
    # Set variables
    local tf_vars=""
    if [ -f "variables.tf" ]; then
        tf_vars="-var=openai_api_key=${OPENAI_API_KEY}"
        if grep -q "openai_admin_key" variables.tf 2>/dev/null && [ -n "${OPENAI_ADMIN_KEY:-}" ]; then
            tf_vars="$tf_vars -var=openai_admin_key=${OPENAI_ADMIN_KEY}"
        fi
        if grep -q "organization_id" variables.tf 2>/dev/null && [ -n "${OPENAI_ORGANIZATION_ID:-}" ]; then
            tf_vars="$tf_vars -var=organization_id=${OPENAI_ORGANIZATION_ID}"
        fi
    fi
    
    # Plan
    echo "  Planning..."
    if ! terraform plan -input=false $tf_vars -out=tfplan > /tmp/plan_$(echo $example | tr '/' '_').log 2>&1; then
        echo -e "${RED}  ✗ Plan failed${NC}"
        return 1
    fi
    
    # Apply
    echo "  Applying..."
    if ! terraform apply -auto-approve tfplan > /tmp/apply_$(echo $example | tr '/' '_').log 2>&1; then
        echo -e "${RED}  ✗ Apply failed${NC}"
        # Try to destroy any partial resources
        terraform destroy -auto-approve $tf_vars > /dev/null 2>&1 || true
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
    if ! terraform destroy -auto-approve $tf_vars > /tmp/destroy_$(echo $example | tr '/' '_').log 2>&1; then
        echo -e "${RED}  ✗ Destroy failed - resources may still exist!${NC}"
        return 1
    fi
    
    echo -e "${GREEN}  ✓ Cleanup successful${NC}"
    
    # Final cleanup
    rm -rf .terraform* terraform.tfstate* tfplan 2>/dev/null || true
    return 0
}

# Get list of examples to test
get_examples() {
    local pattern=${1:-}
    
    if [ -z "$pattern" ] || [ "$pattern" = "all" ]; then
        # All examples
        find "$EXAMPLES_DIR" -type f -name "*.tf" | grep -E "(resource|data-source|provider)\.tf$" | \
            sed "s|$EXAMPLES_DIR/||" | sed 's|/[^/]*\.tf$||' | sort -u
    elif [ -d "$EXAMPLES_DIR/$pattern" ]; then
        # Directory pattern (e.g., resources/, data-sources/)
        find "$EXAMPLES_DIR/$pattern" -type f -name "*.tf" | grep -E "(resource|data-source|provider)\.tf$" | \
            sed "s|$EXAMPLES_DIR/||" | sed 's|/[^/]*\.tf$||' | sort -u
    else
        # Single example
        echo "$pattern"
    fi
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
    }
  }
}

provider "openai" {}

# Test with embeddings (cheap and fast)
resource "openai_embedding" "test" {
  model = "text-embedding-3-small"
  input = jsonencode(["Quick test"])
}

output "test_passed" {
  value = "✓ Provider working correctly"
}
EOF

    # Run test
    terraform init -upgrade > /dev/null 2>&1
    
    if terraform apply -auto-approve > /tmp/quick-test-apply.log 2>&1; then
        echo -e "${GREEN}✓ Quick test passed${NC}"
        terraform destroy -auto-approve > /dev/null 2>&1
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
    for tf_file in $(find "$EXAMPLES_DIR" -name "*.tf" -type f); do
        dir=$(dirname "$tf_file")
        if ls "$dir"/.terraform* "$dir"/terraform.tfstate* "$dir"/tfplan 2>/dev/null | grep -q .; then
            example=${dir#$EXAMPLES_DIR/}
            echo "  Cleaning $example"
            rm -rf "$dir"/.terraform* "$dir"/terraform.tfstate* "$dir"/tfplan 2>/dev/null || true
            ((cleaned++))
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
            local examples=($(get_examples "$target"))
            
            if [ ${#examples[@]} -eq 0 ]; then
                echo -e "${RED}No examples found for pattern: $target${NC}"
                exit 1
            fi
            
            echo -e "${BLUE}Testing ${#examples[@]} examples (plan only)${NC}"
            echo ""
            
            local passed=0
            local failed=0
            
            for example in "${examples[@]}"; do
                if test_plan "$example"; then
                    ((passed++))
                else
                    ((failed++))
                fi
            done
            
            echo ""
            echo -e "${BLUE}Summary:${NC} ${GREEN}$passed passed${NC}, ${RED}$failed failed${NC}"
            ;;
            
        apply)
            check_env
            local examples=($(get_examples "$target"))
            
            if [ ${#examples[@]} -eq 0 ]; then
                echo -e "${RED}No examples found for pattern: $target${NC}"
                exit 1
            fi
            
            echo -e "${YELLOW}WARNING: This will create real resources and may incur costs!${NC}"
            echo "Will test ${#examples[@]} examples"
            read -p "Continue? (y/N) " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                echo "Cancelled."
                exit 0
            fi
            
            echo -e "${BLUE}Testing ${#examples[@]} examples (with apply)${NC}"
            echo ""
            
            local passed=0
            local failed=0
            
            for example in "${examples[@]}"; do
                if test_apply "$example"; then
                    ((passed++))
                else
                    ((failed++))
                fi
                echo ""
            done
            
            echo -e "${BLUE}Summary:${NC} ${GREEN}$passed passed${NC}, ${RED}$failed failed${NC}"
            ;;
            
        quick)
            check_env
            quick_test
            ;;
            
        cleanup)
            cleanup_all
            ;;
            
        *)
            echo "Usage: $0 [plan|apply|quick|cleanup] [example_pattern|all]"
            echo ""
            echo "Commands:"
            echo "  plan [pattern]   - Run terraform plan"
            echo "  apply [pattern]  - Run terraform apply with cleanup"
            echo "  quick            - Quick verification test"
            echo "  cleanup          - Remove all terraform artifacts"
            echo ""
            echo "Patterns:"
            echo "  all                               - All examples (default)"
            echo "  resources/                        - All resource examples"
            echo "  data-sources/                     - All data source examples"
            echo "  resources/openai_chat_completion  - Specific example"
            exit 1
            ;;
    esac
}

main "$@"