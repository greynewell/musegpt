# Clean script to wipe out all transient state

# Remove build directory
Remove-Item -Path "build" -Recurse -Force -ErrorAction SilentlyContinue

# Remove models directory
Remove-Item -Path "models" -Recurse -Force -ErrorAction SilentlyContinue

# Remove .env directory
Remove-Item -Path ".env" -Recurse -Force -ErrorAction SilentlyContinue

Write-Host "Cleanup completed."