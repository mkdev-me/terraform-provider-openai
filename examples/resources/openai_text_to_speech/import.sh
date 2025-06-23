#!/bin/bash
# Text-to-speech resources cannot be imported

echo "Text-to-speech is a one-time operation, not a persistent resource."
echo "It cannot be imported because there's no state to track in OpenAI."
echo ""
echo "To use text-to-speech:"
echo "1. Create a resource configuration (see resource.tf)"
echo "2. Run 'terraform apply' to generate the audio file"