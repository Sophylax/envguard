$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
$tmpRepo = Join-Path ([System.IO.Path]::GetTempPath()) ("envguard-smoke-" + [System.Guid]::NewGuid().ToString("N"))

try {
    New-Item -ItemType Directory -Path $tmpRepo | Out-Null
    Set-Location $tmpRepo

    git init -b main | Out-Null
    git config user.name "envguard-ci"
    git config user.email "envguard-ci@example.com"

    @'
const key = "AKIA1234567890ABCDEF";
'@ | Set-Content -Path secret.js -NoNewline

    git add secret.js

    $binDir = Join-Path $repoRoot "bin"
    $env:PATH = "$binDir;$env:PATH"
    & (Join-Path $binDir "envguard.exe") install | Out-Null

    $commitOutput = & git commit -m blocked 2>&1
    if ($LASTEXITCODE -eq 0) {
        throw "expected commit with staged secret to be blocked"
    }
    if (-not ($commitOutput | Out-String).Contains("Commit blocked.")) {
        throw "expected blocked commit output to mention 'Commit blocked.'"
    }

    $jsonOutput = & (Join-Path $binDir "envguard.exe") check secret.js --json 2>&1
    if ($LASTEXITCODE -eq 0) {
        throw "expected JSON check to report findings before allowlisting"
    }

    $findings = $jsonOutput | ConvertFrom-Json
    if ($findings.Count -lt 1) {
        throw "expected at least one finding in JSON output"
    }

    $fingerprint = $findings[0].fingerprint
    if (-not $fingerprint) {
        throw "failed to extract finding fingerprint from JSON output"
    }

    & (Join-Path $binDir "envguard.exe") allow $fingerprint | Out-Null
    git add .envguard-ignore
    & git commit -m "allow known fixture" *> $null

    Add-Content -Path .git/hooks/pre-commit -Value "`necho foreign-hook"
    & (Join-Path $binDir "envguard.exe") uninstall | Out-Null

    $hookContent = Get-Content .git/hooks/pre-commit -Raw
    if (-not $hookContent.Contains("foreign-hook")) {
        throw "expected uninstall to preserve foreign hook content"
    }
    if ($hookContent.Contains("envguard check")) {
        throw "envguard uninstall left hook content behind"
    }
}
finally {
    Set-Location $repoRoot
    if (Test-Path $tmpRepo) {
        Remove-Item -Recurse -Force $tmpRepo
    }
}
