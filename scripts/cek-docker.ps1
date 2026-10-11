# Menjalankan tes API di dalam container backend, yaitu lingkungan yang sama
# dengan deploy: Postgres, MinIO, dan LibreOffice ada semua di sana. Inilah
# satu-satunya cara tes persetujuan perjanjian (DOCX ke PDF) ikut jalan.
#
# Pakai:  .\scripts\cek-docker.ps1
#         .\scripts\cek-docker.ps1 -Tes Perjanjian -Rinci
#
# Butuh container backend, postgres, dan minio hidup (docker compose up -d).
# Tes membuat ulang database legal_riu_test dan memakai bucket legal-riu-test
# di Postgres dan MinIO milik Docker; data aplikasi tidak disentuh.

param(
    [string]$Tes,   # hanya tes yang namanya cocok, mis. -Tes Perjanjian
    [switch]$Rinci  # tampilkan tiap tes, termasuk yang dilewati
)

$ErrorActionPreference = 'Stop'
$argTes = @()
if ($Tes) { $argTes += '-test.run', $Tes }
if ($Rinci) { $argTes += '-test.v' }
$container = 'legal_riu_backend'
$root = Split-Path -Parent $PSScriptRoot
$tmp = Join-Path ([IO.Path]::GetTempPath()) 'alea-tes-docker'

# Docker bisa terpasang langsung di Windows atau hanya di dalam WSL.
$lewatWsl = -not (Get-Command docker -ErrorAction SilentlyContinue)
function DockerCmd {
    if ($lewatWsl) { & wsl.exe -e docker @args } else { & docker @args }
}

Write-Host "== Menyusun program tes untuk Linux ==" -ForegroundColor Cyan
if (Test-Path $tmp) { Remove-Item -Recurse -Force $tmp }
New-Item -ItemType Directory (Join-Path $tmp 'testdata') | Out-Null
Push-Location (Join-Path $root 'backend')
try {
    $env:GOOS = 'linux'; $env:GOARCH = 'amd64'; $env:CGO_ENABLED = '0'
    go test -c ./internal/router -o (Join-Path $tmp 'router.test')
} finally {
    Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED
    Pop-Location
}
if ($LASTEXITCODE -ne 0) { Write-Host "`nGAGAL: program tes tidak bisa disusun" -ForegroundColor Red; exit 1 }
Copy-Item (Join-Path $root 'backend\internal\router\testdata\*') (Join-Path $tmp 'testdata')

$sumber = $tmp
if ($lewatWsl) { $sumber = (& wsl.exe -e wslpath -a $tmp).Trim() }

$kode = 1
try {
    Write-Host "`n== Menjalankan tes di container $container ==" -ForegroundColor Cyan
    DockerCmd exec $container rm -rf /tmp/alea-test
    if ($LASTEXITCODE -ne 0) { Write-Host "`nGAGAL: container $container tidak bisa dijangkau" -ForegroundColor Red; exit 1 }
    DockerCmd cp $sumber "${container}:/tmp/alea-test"
    DockerCmd exec $container chmod +x /tmp/alea-test/router.test
    DockerCmd exec -w /tmp/alea-test $container ./router.test @argTes
    $kode = $LASTEXITCODE
} finally {
    DockerCmd exec $container rm -rf /tmp/alea-test
    Remove-Item -Recurse -Force $tmp
}

if ($kode -ne 0) { Write-Host "`nGAGAL: tes di container" -ForegroundColor Red; exit 1 }
Write-Host "`nTes di container lulus." -ForegroundColor Green
