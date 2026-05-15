# deploy-sidecars.ps1
# PowerShell script to deploy DroidRun and Open Computer Use sidecar services

$RepoRoot = $PSScriptRoot
Set-Location $RepoRoot

Write-Host "Starting deployment of sidecar microservices..." -ForegroundColor Cyan

# 1. Deploy DroidRun
Write-Host "`n[1/2] Deploying DroidRun sidecar..." -ForegroundColor Yellow
Set-Location "$RepoRoot/droidrun-deploy"
& "C:\Users\mikek\.fly\bin\flyctl.exe" deploy --app koola10-droidrun
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: DroidRun deployment failed with exit code $LASTEXITCODE" -ForegroundColor Red
    exit $LASTEXITCODE
}

# 2. Deploy Open Computer Use
Write-Host "`n[2/2] Deploying Open Computer Use sidecar..." -ForegroundColor Yellow
Set-Location "$RepoRoot/desktop-deploy"
& "C:\Users\mikek\.fly\bin\flyctl.exe" deploy --app koola10-desktop
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Open Computer Use deployment failed with exit code $LASTEXITCODE" -ForegroundColor Red
    exit $LASTEXITCODE
}

Set-Location $RepoRoot
Write-Host "`nSuccess: Both sidecar microservices have been deployed successfully!" -ForegroundColor Green
