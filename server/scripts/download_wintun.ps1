$ErrorActionPreference = "Stop"

$WintunVersion = "0.14.1"
$WintunUrl = "https://www.wintun.net/builds/wintun-$WintunVersion.zip"
$ZipFile = "wintun.zip"
$ExtractPath = "wintun_temp"
$TargetDll = "wintun.dll"

Write-Host "Downloading Wintun $WintunVersion..."
Invoke-WebRequest -Uri $WintunUrl -OutFile $ZipFile

Write-Host "Extracting..."
Expand-Archive -Path $ZipFile -DestinationPath $ExtractPath -Force

# Determine architecture
if ([IntPtr]::Size -eq 8) {
    $Arch = "amd64"
} else {
    $Arch = "x86"
}

$SourceDll = "$ExtractPath\wintun\bin\$Arch\wintun.dll"
Write-Host "Copying $Arch $TargetDll..."
Copy-Item -Path $SourceDll -Destination "..\$TargetDll" -Force

# Cleanup
Write-Host "Cleaning up..."
Remove-Item -Path $ZipFile -Force
Remove-Item -Path $ExtractPath -Recurse -Force

Write-Host "Success! wintun.dll has been downloaded."
