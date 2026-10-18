[CmdletBinding()]
param(
    [switch]$NoPush
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Invoke-Git {
    param(
        [Parameter(Mandatory, ValueFromRemainingArguments)]
        [string[]]$Args
    )

    & git @Args
    $exitCode = $LASTEXITCODE
    if ($exitCode -ne 0) {
        throw "git $($Args -join ' ') failed with exit code $exitCode."
    }
}

function Get-GitOutput {
    param(
        [Parameter(Mandatory, ValueFromRemainingArguments)]
        [string[]]$Args
    )

    $output = & git @Args
    $exitCode = $LASTEXITCODE
    if ($exitCode -ne 0) {
        throw "git $($Args -join ' ') failed with exit code $exitCode."
    }

    return ($output -join [Environment]::NewLine).Trim()
}

function Require-CleanWorktree {
    $status = Get-GitOutput status --porcelain
    if ($status) {
        throw "Working tree is not clean. Commit or stash your changes before running this script."
    }
}

function Require-Remote {
    param(
        [Parameter(Mandatory)]
        [string]$Name
    )

    $remotes = (Get-GitOutput remote) -split "\r?\n"
    if ($Name -notin $remotes) {
        throw "Missing required git remote '$Name'."
    }
}

function Sync-OriginMain {
    $counts = (Get-GitOutput rev-list --left-right --count main...origin/main) -split "\s+"
    $localOnly = [int]$counts[0]
    $remoteOnly = [int]$counts[1]

    if ($localOnly -gt 0 -and $remoteOnly -gt 0) {
        throw "Local main and origin/main have diverged. Reconcile them manually before syncing upstream."
    }

    if ($remoteOnly -gt 0) {
        Write-Host "Fast-forwarding local main from origin/main..."
        Invoke-Git merge --ff-only origin/main
        return
    }

    if ($localOnly -gt 0) {
        Write-Host "Local main is ahead of origin/main; keeping local commits."
        return
    }

    Write-Host "Local main already matches origin/main."
}

try {
    $repoRoot = Get-GitOutput rev-parse --show-toplevel
    Set-Location $repoRoot

    Require-CleanWorktree
    Require-Remote origin
    Require-Remote upstream

    Write-Host "Fetching latest refs from origin and upstream..."
    Invoke-Git fetch origin main
    Invoke-Git fetch upstream main nightly

    Write-Host "Switching to main..."
    Invoke-Git switch main

    Sync-OriginMain

    Write-Host "Merging upstream/nightly into main..."
    Invoke-Git merge --no-edit upstream/nightly

    Write-Host "Merging upstream/main into main..."
    Invoke-Git merge --no-edit upstream/main

    if ($NoPush) {
        Write-Host "Skipping push because -NoPush was provided."
        exit 0
    }

    Write-Host "Pushing main to origin..."
    Invoke-Git push origin main

    Write-Host "Sync complete."
}
catch {
    Write-Error $_.Exception.Message

    $mergeHead = Join-Path (Get-GitOutput rev-parse --git-dir) "MERGE_HEAD"
    if (Test-Path $mergeHead) {
        Write-Host ""
        Write-Host "A merge stopped for conflicts."
        Write-Host "Resolve the files, then run:"
        Write-Host "  git add ."
        Write-Host "  git commit"
        Write-Host "  git push origin main"
    }

    exit 1
}
