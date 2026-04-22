Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$RepoSlug = 'chaitin/chaitin-cli'
$InstallName = 'chaitin-cli.exe'

function Write-Log {
    param([string]$Message)

    [Console]::Error.WriteLine($Message)
}

function Get-IsAdmin {
    $identity = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($identity)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Get-GoArch {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()

    switch ($arch) {
        'x64' { return 'amd64' }
        'arm64' { return 'arm64' }
        default { throw "unsupported architecture: $arch" }
    }
}

function Normalize-Tag {
    param([string]$Version)

    if ([string]::IsNullOrWhiteSpace($Version)) {
        throw 'release version is empty'
    }

    if ($Version.StartsWith('v')) {
        return $Version
    }

    return "v$Version"
}

function Get-LatestTag {
    $headers = @{ 'User-Agent' = 'chaitin-cli-installer' }
    $release = Invoke-RestMethod -Headers $headers -Uri "https://api.github.com/repos/$RepoSlug/releases/latest"

    if ([string]::IsNullOrWhiteSpace($release.tag_name)) {
        throw 'failed to parse latest release tag'
    }

    return $release.tag_name
}

function Get-PathEntries {
    param([string]$Value)

    if ([string]::IsNullOrWhiteSpace($Value)) {
        return @()
    }

    return @($Value -split ';' | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
}

function Test-PathContains {
    param(
        [string[]]$Entries,
        [string]$Directory
    )

    $normalizedTarget = $Directory.TrimEnd('\\')
    foreach ($entry in $Entries) {
        if ($entry.TrimEnd('\\').Equals($normalizedTarget, [System.StringComparison]::OrdinalIgnoreCase)) {
            return $true
        }
    }

    return $false
}

function Add-ToPath {
    param(
        [string]$Directory,
        [System.EnvironmentVariableTarget]$Scope
    )

    $persistentValue = [Environment]::GetEnvironmentVariable('Path', $Scope)
    $persistentEntries = Get-PathEntries $persistentValue
    if (-not (Test-PathContains -Entries $persistentEntries -Directory $Directory)) {
        $newValue = if ([string]::IsNullOrWhiteSpace($persistentValue)) {
            $Directory
        } else {
            "$persistentValue;$Directory"
        }

        [Environment]::SetEnvironmentVariable('Path', $newValue, $Scope)
        Write-Log "updated $Scope PATH"
    }

    $sessionEntries = Get-PathEntries $env:Path
    if (-not (Test-PathContains -Entries $sessionEntries -Directory $Directory)) {
        $env:Path = if ([string]::IsNullOrWhiteSpace($env:Path)) {
            $Directory
        } else {
            "$Directory;$env:Path"
        }
    }
}

function Get-InstallTargets {
    $targets = @()

    if (-not [string]::IsNullOrWhiteSpace($env:CHAITIN_CLI_INSTALL_DIR)) {
        $targets += [pscustomobject]@{
            Directory = $env:CHAITIN_CLI_INSTALL_DIR
            Scope = [System.EnvironmentVariableTarget]::User
        }
    }

    if ((Get-IsAdmin) -and -not [string]::IsNullOrWhiteSpace($env:ProgramFiles)) {
        $targets += [pscustomobject]@{
            Directory = Join-Path $env:ProgramFiles 'chaitin-cli\bin'
            Scope = [System.EnvironmentVariableTarget]::Machine
        }
    }

    if (-not [string]::IsNullOrWhiteSpace($env:LOCALAPPDATA)) {
        $targets += [pscustomobject]@{
            Directory = Join-Path $env:LOCALAPPDATA 'Programs\chaitin-cli\bin'
            Scope = [System.EnvironmentVariableTarget]::User
        }
    }

    if (-not [string]::IsNullOrWhiteSpace($env:USERPROFILE)) {
        $targets += [pscustomobject]@{
            Directory = Join-Path $env:USERPROFILE 'bin'
            Scope = [System.EnvironmentVariableTarget]::User
        }
    }

    return $targets
}

function Install-Binary {
    param(
        [string]$SourcePath,
        [string]$DestinationDirectory,
        [System.EnvironmentVariableTarget]$Scope
    )

    New-Item -ItemType Directory -Path $DestinationDirectory -Force | Out-Null
    $destinationPath = Join-Path $DestinationDirectory $InstallName
    Copy-Item -LiteralPath $SourcePath -Destination $destinationPath -Force
    Add-ToPath -Directory $DestinationDirectory -Scope $Scope
    return $destinationPath
}

function Install-WithFallback {
    param([string]$SourcePath)

    $errors = New-Object System.Collections.Generic.List[string]
    foreach ($target in Get-InstallTargets) {
        try {
            return Install-Binary -SourcePath $SourcePath -DestinationDirectory $target.Directory -Scope $target.Scope
        } catch {
            $errors.Add("$($target.Directory): $($_.Exception.Message)")
        }
    }

    throw "failed to install chaitin-cli. attempts: $($errors -join '; ')"
}

function Download-ReleaseBinary {
    param(
        [string]$GoArch,
        [string]$WorkDir
    )

    $tag = if ([string]::IsNullOrWhiteSpace($env:CHAITIN_CLI_VERSION)) {
        Normalize-Tag (Get-LatestTag)
    } else {
        Normalize-Tag $env:CHAITIN_CLI_VERSION
    }

    $headers = @{ 'User-Agent' = 'chaitin-cli-installer' }
    $assetName = "chaitin-cli_${tag}_windows_${GoArch}.zip"
    $archivePath = Join-Path $WorkDir $assetName
    $downloadUrl = "https://github.com/$RepoSlug/releases/download/$tag/$assetName"

    Write-Log "downloading $downloadUrl"
    Invoke-WebRequest -Headers $headers -Uri $downloadUrl -OutFile $archivePath

    $extractDir = Join-Path $WorkDir 'extract'
    Expand-Archive -LiteralPath $archivePath -DestinationPath $extractDir -Force

    $binary = Get-ChildItem -Path $extractDir -Recurse -File | Where-Object { $_.Name -ieq $InstallName } | Select-Object -First 1
    if ($null -eq $binary) {
        throw "failed to locate $InstallName in downloaded archive"
    }

    return $binary.FullName
}

function Main {
    $existing = Get-Command chaitin-cli -ErrorAction SilentlyContinue
    if ($null -ne $existing) {
        return $existing.Source
    }

    $goArch = Get-GoArch
    $workDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString('N'))

    try {
        New-Item -ItemType Directory -Path $workDir -Force | Out-Null
        $sourceBinary = Download-ReleaseBinary -GoArch $goArch -WorkDir $workDir
        $installedPath = Install-WithFallback -SourcePath $sourceBinary
        Write-Log "installed chaitin-cli to $installedPath"
        return $installedPath
    } finally {
        if (Test-Path -LiteralPath $workDir) {
            Remove-Item -LiteralPath $workDir -Recurse -Force
        }
    }
}

$installed = Main
Write-Output $installed
