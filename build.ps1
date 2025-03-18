param (
    [string]$versionTag
)

if (-not $versionTag) {
    Write-Host "Usage: .\build.ps1 <versionTag>"
    exit 1
}

# Build for Linux
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -ldflags="-w -s -X main.AppVersion=$versionTag" -o "./build/bptvnftester-linux-amd64"

