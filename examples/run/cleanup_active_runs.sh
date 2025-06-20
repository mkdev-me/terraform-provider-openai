#!/bin/bash

# Script to clean up active runs before testing
# This helps when previous terraform apply failed and left active runs

echo "Cleaning up active runs from previous attempts..."

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed"
    exit 1
fi

# Check API key
if [ -z "$OPENAI_API_KEY" ]; then
    echo "Error: OPENAI_API_KEY not set"
    exit 1
fi

# Function to cancel runs for a thread
cancel_runs_for_thread() {
    local thread_id=$1
    echo "Checking thread: $thread_id"
    
    # List runs for the thread
    runs=$(curl -s https://api.openai.com/v1/threads/$thread_id/runs \
        -H "Authorization: Bearer $OPENAI_API_KEY" \
        -H "OpenAI-Beta: assistants=v2")
    
    # Check each run
    echo "$runs" | jq -r '.data[]? | select(.status == "in_progress" or .status == "queued" or .status == "requires_action") | .id' | while read run_id; do
        if [ ! -z "$run_id" ]; then
            echo "  Cancelling active run: $run_id"
            curl -s -X POST https://api.openai.com/v1/threads/$thread_id/runs/$run_id/cancel \
                -H "Authorization: Bearer $OPENAI_API_KEY" \
                -H "OpenAI-Beta: assistants=v2" > /dev/null
        fi
    done
}

# Get thread IDs from terraform state if it exists
if [ -f terraform.tfstate ]; then
    echo "Found terraform.tfstate, extracting thread IDs..."
    
    # Extract thread IDs from state
    thread_ids=$(cat terraform.tfstate | jq -r '.resources[]? | select(.type == "openai_thread") | .instances[].attributes.id' 2>/dev/null)
    
    if [ ! -z "$thread_ids" ]; then
        echo "$thread_ids" | while read thread_id; do
            if [ ! -z "$thread_id" ]; then
                cancel_runs_for_thread "$thread_id"
            fi
        done
    else
        echo "No threads found in state"
    fi
else
    echo "No terraform.tfstate found"
fi

echo "Cleanup complete!"