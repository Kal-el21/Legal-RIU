# Menjalankan semua pemeriksaan sebelum push. Berhenti di kegagalan pertama.
# Pakai: .\scripts\cek.ps1 (dari folder mana pun)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot

function Langkah($nama, $folder, [scriptblock]$perintah) {
    Write-Host "`n== $nama ==" -ForegroundColor Cyan
    Push-Location (Join-Path $root $folder)
    try { & $perintah } finally { Pop-Location }
    if ($LASTEXITCODE -ne 0) {
        Write-Host "`nGAGAL: $nama" -ForegroundColor Red
        exit 1
    }
}

Langkah 'Backend: go vet'  'backend'  { go vet ./... }
# -count=1: hasil tes yang tersimpan tidak memperhatikan perubahan di frontend/src
# (TestAlamatFrontendAda) maupun isi database, jadi tes selalu dijalankan ulang.
Langkah 'Backend: go test' 'backend'  { go test -count=1 ./... }
Langkah 'Frontend: tsc'    'frontend' { npx tsc --noEmit }
Langkah 'Frontend: lint'   'frontend' { npm run lint }
Langkah 'Frontend: build'  'frontend' { npm run build }

Write-Host "`nSemua pemeriksaan lulus." -ForegroundColor Green
