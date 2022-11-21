$ErrorActionPreference = 'Stop';

$toolsDir   = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$url        = '{{DownloadUrl}}'
$url64      = 'https://download.microsoft.com/download/d/4/4/d4403a51-2ab7-4ea8-b850-d2710c5e1323/sqlcmd_0.8.1-1.msi'

$packageArgs = @{
  packageName   = $env:ChocolateyPackageName
  unzipLocation = $toolsDir
  fileType      = 'MSI'
  url           = $url
  url64bit      = $url64

  softwareName  = 'sqlcmd*'

  checksum      = '{{Checksum}}'
  checksumType  = '{{ChecksumType}}'
  checksum64    = '03587762932D5A66ACFE15D306FE14645D53BC61162B4DA0D9AF29B4A8A1550D'
  checksumType64= 'sha256'

  silentArgs    = "/qn /norestart /l*v `"$($env:TEMP)\$($packageName).$($env:chocolateyPackageVersion).MsiInstall.log`""
  validExitCodes= @(0, 3010, 1641)
}

Install-ChocolateyPackage @packageArgs










    









