# Testing Scripts

This directory contains the testing script for the OpenAI Terraform provider.

## test_examples.sh

A unified script for testing provider examples with multiple commands:

### Usage

```bash
./test_examples.sh [command] [target]
```

### Commands

- **`plan`** - Run terraform plan (default)
- **`apply`** - Run terraform apply with automatic cleanup
- **`quick`** - Quick verification test
- **`cleanup`** - Remove all terraform artifacts

### Examples

```bash
# Quick test
./test_examples.sh quick

# Test all examples (plan only)
./test_examples.sh plan

# Test specific example
./test_examples.sh plan image

# Apply specific example
./test_examples.sh apply chat_completion

# Apply all (WARNING: creates resources!)
./test_examples.sh apply

# Clean everything
./test_examples.sh cleanup
```

## Environment Setup

Before running tests, set your API keys:

```bash
export OPENAI_API_KEY="sk-proj-..."
export OPENAI_ADMIN_KEY="sk-admin-..."
```

Or use a .env file:

```bash
source ../../.env
```

See [TESTING.md](../../TESTING.md) for detailed documentation.