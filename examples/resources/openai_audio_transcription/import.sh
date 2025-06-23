#!/bin/bash
# Audio transcription resources cannot be imported

echo "Audio transcription is a one-time operation, not a persistent resource."
echo "It cannot be imported because there's no state to track in OpenAI."
echo ""
echo "To use audio transcription:"
echo "1. Create a resource configuration (see resource.tf)"
echo "2. Run 'terraform apply' to perform the transcription"