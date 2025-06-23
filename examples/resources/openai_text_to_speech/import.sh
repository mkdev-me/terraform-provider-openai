#!/bin/bash
###############################################################################
# OpenAI Text-to-Speech Resource Note
###############################################################################
# Text-to-speech operations are not persistent resources in the OpenAI API.
# They are one-time operations that generate audio files from text input.
#
# Therefore, this resource cannot be imported as it doesn't have a persistent
# state in OpenAI's system.
#
# To use this resource:
# 1. Create a new resource configuration in your Terraform files
# 2. Run 'terraform apply' to generate the speech audio
# 3. The audio file will be saved to the specified output path
#
# Example usage:
#   See resource.tf for a complete example configuration
#
###############################################################################

echo "Text-to-speech resources cannot be imported."
echo ""
echo "Text-to-speech operations are one-time requests that:"
echo "- Send text to OpenAI's TTS API"
echo "- Generate an audio file"
echo "- Save the file locally"
echo "- Do not maintain persistent state in OpenAI"
echo ""
echo "To generate speech, create a new resource configuration and apply it."
echo "See resource.tf for an example."