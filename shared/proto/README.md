# Protocol Buffer Definitions

This directory contains the Protocol Buffer definitions for inter-service communication.

## Files

- `auth.proto` - Authentication service interfaces
- `analytics.proto` - ML service interfaces  
- `common.proto` - Shared message types and enums

## Code Generation

Run `pnpm proto:generate` from the root directory to generate client/server stubs for all languages.

## Generated Code Locations

- Go: `shared/proto/gen/go/`
- TypeScript: `shared/proto/gen/ts/`
- Python: `shared/proto/gen/python/`