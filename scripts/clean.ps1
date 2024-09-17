#!/usr/bin/env pwsh

# Wipe out all transient state
Remove-Item -Path 'build', 'models', '.env' -Recurse -Force
