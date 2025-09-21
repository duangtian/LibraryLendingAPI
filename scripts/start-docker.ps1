Param(
    [switch]$Rebuild,
    [switch]$ResetDb,
    [switch]$Logs,
    [switch]$Wait
)

$ErrorActionPreference = 'Stop'

function Ensure-Docker() {
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        Write-Error 'Docker CLI not found. Install Docker Desktop first.'
    }
}

function Compose() {
    param([Parameter(ValueFromRemainingArguments=$true)][string[]]$Args)
    try {
        docker compose @Args
    } catch {
        # Fallback for older setups
        if (Get-Command docker-compose -ErrorAction SilentlyContinue) {
            docker-compose @Args
        } else { throw }
    }
}

function Wait-Health() {
    $max = 40 # ~40 * 0.5s = 20s
    for ($i=1; $i -le $max; $i++) {
        try {
            $resp = Invoke-WebRequest -Uri 'http://localhost:8080/v1/healthz' -UseBasicParsing -TimeoutSec 3
            if ($resp.StatusCode -eq 200) { Write-Host "Health OK (attempt $i)."; return }
        } catch {}
        Start-Sleep -Milliseconds 500
    }
    Write-Warning 'Health endpoint not ready after timeout.'
}

# Normalize working directory to project root (one level up from script dir)
$ProjectRoot = (Resolve-Path "$PSScriptRoot\..").ProviderPath
Set-Location $ProjectRoot

Ensure-Docker

if ($ResetDb) {
    Write-Host 'Resetting database volumes...' -ForegroundColor Yellow
    Compose down -v
}

if ($Rebuild) {
    Write-Host 'Building images (no cache)...' -ForegroundColor Cyan
    Compose build --no-cache
} else {
    Write-Host 'Building images (cached)...' -ForegroundColor Cyan
    Compose build
}

Write-Host 'Starting stack...' -ForegroundColor Green
Compose up -d

if ($Wait) { Wait-Health }

if ($Logs) {
    Write-Host 'Tailing API logs (Ctrl+C to exit)...' -ForegroundColor Magenta
    Compose logs -f api
}

Write-Host 'Done. API -> http://localhost:8080  Health: /v1/healthz'
