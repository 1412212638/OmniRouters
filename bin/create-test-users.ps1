[CmdletBinding()]
param(
    [string]$BaseUrl = "http://127.0.0.1:3000",
    [string]$AdminToken = $env:OMNIROUTERS_ADMIN_TOKEN,
    [int]$AdminUserId = 0,
    [int]$Count = 10000,
    [string]$Prefix = "loadtest_",
    [string]$OutputFile = "",
    [int]$MaxAttempts = 5,
    [int]$DelayMilliseconds = 25
)

$ErrorActionPreference = "Stop"

if ([string]::IsNullOrWhiteSpace($AdminToken)) {
    throw "请通过 -AdminToken 或 OMNIROUTERS_ADMIN_TOKEN 提供管理员访问令牌。"
}
if ($AdminUserId -le 0) {
    throw "请通过 -AdminUserId 提供管理员用户 ID。"
}
if ($Count -lt 1 -or $Count -gt 100000) {
    throw "Count 必须在 1 到 100000 之间。"
}
if ($Prefix.Length -gt 9) {
    throw "Prefix 最长为 9 个字符，以保证随机用户名不超过 20 个字符。"
}
if ($MaxAttempts -lt 1 -or $MaxAttempts -gt 10) {
    throw "MaxAttempts 必须在 1 到 10 之间。"
}
if ($DelayMilliseconds -lt 0 -or $DelayMilliseconds -gt 60000) {
    throw "DelayMilliseconds 必须在 0 到 60000 之间。"
}

$baseUri = [Uri]$BaseUrl.TrimEnd('/')
$createUri = [Uri]::new($baseUri, "/api/user/")

if ([string]::IsNullOrWhiteSpace($OutputFile)) {
    $stamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $OutputFile = Join-Path (Get-Location) "created-test-users-$stamp.txt"
}
$outputParent = Split-Path -Parent $OutputFile
if ($outputParent -and -not (Test-Path -LiteralPath $outputParent)) {
    New-Item -ItemType Directory -Path $outputParent -Force | Out-Null
}

# The file contains usernames only. Passwords are generated per request and
# deliberately discarded after the request so this script does not create a
# credential export for temporary test accounts.
$created = [System.Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)
if (Test-Path -LiteralPath $OutputFile) {
    foreach ($line in Get-Content -LiteralPath $OutputFile) {
        $username = $line.Trim()
        if ($username) { [void]$created.Add($username) }
    }
}

$alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
$random = [System.Security.Cryptography.RandomNumberGenerator]::Create()

function New-RandomString {
    param([int]$Length)

    $bytes = [byte[]]::new($Length)
    $random.GetBytes($bytes)
    $chars = [char[]]::new($Length)
    for ($i = 0; $i -lt $Length; $i++) {
        $chars[$i] = $alphabet[$bytes[$i] % $alphabet.Length]
    }
    return -join $chars
}

function Get-ErrorText {
    param($Exception)
    $response = $Exception.Exception.Response
    if ($response -and $response.GetResponseStream()) {
        $reader = [System.IO.StreamReader]::new($response.GetResponseStream())
        try { return $reader.ReadToEnd() } finally { $reader.Dispose() }
    }
    return $Exception.Exception.Message
}

try {
    while ($created.Count -lt $Count) {
        $username = $null
        $payload = $null
        $succeeded = $false

        for ($attempt = 1; $attempt -le $MaxAttempts -and -not $succeeded; $attempt++) {
            $username = $Prefix + (New-RandomString -Length 10)
            if ($created.Contains($username)) { continue }

            # Password is intentionally kept only in this request payload.
            $payload = @{
                username     = $username
                display_name = $username
                password     = New-RandomString -Length 20
                role         = 1
            } | ConvertTo-Json -Compress

            try {
                $response = Invoke-RestMethod `
                    -Uri $createUri `
                    -Method Post `
                    -Headers @{
                        Authorization = "Bearer $AdminToken"
                        "New-Api-User" = [string]$AdminUserId
                    } `
                    -ContentType "application/json" `
                    -Body $payload `
                    -TimeoutSec 60

                if ($response.success -ne $true) {
                    throw "接口返回 success=false"
                }

                [void]$created.Add($username)
                Add-Content -LiteralPath $OutputFile -Value $username -Encoding UTF8
                $succeeded = $true
            } catch {
                $errorText = Get-ErrorText $_
                if ($attempt -eq $MaxAttempts) {
                    throw "创建 $username 失败（第 $attempt 次）：$errorText"
                }
                Start-Sleep -Milliseconds ([Math]::Min(5000, 250 * $attempt))
            } finally {
                $payload = $null
            }
        }

        if (-not $succeeded) {
            throw "无法创建下一个测试用户。"
        }

        if (($created.Count % 100) -eq 0 -or $created.Count -eq $Count) {
            Write-Host ("已创建 {0}/{1} 个测试账号。" -f $created.Count, $Count)
        }
        if ($DelayMilliseconds -gt 0) {
            Start-Sleep -Milliseconds $DelayMilliseconds
        }
    }
} finally {
    $random.Dispose()
}

Write-Host "创建完成。用户名清单：$OutputFile"
