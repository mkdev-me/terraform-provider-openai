#!/bin/bash
###############################################################################
# OpenAI Audio Transcription Resource Note
###############################################################################
# Audio transcriptions are not persistent resources in the OpenAI API.
# They are one-time operations that return a transcription result.
#
# Therefore, this resource cannot be imported as it doesn't have a persistent
# state in OpenAI's system.
#
# To use this resource:
# 1. Create a new resource configuration in your Terraform files
# 2. Run 'terraform apply' to perform the transcription
# 3. The transcription result will be available in the resource's outputs
#
# Example usage:
#   See resource.tf for a complete example configuration
#
###############################################################################

echo "Audio transcription resources cannot be imported."
echo ""
echo "Audio transcriptions are one-time operations that:"
echo "- Upload an audio file to OpenAI"
echo "- Return a transcription result"
echo "- Do not maintain persistent state in OpenAI"
echo ""
echo "To perform a transcription, create a new resource configuration and apply it."
echo "See resource.tf for an example."