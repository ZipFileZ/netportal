@echo off
rem Builds setup\out\NetPortal-Setup-<version>.exe from amd64\, x86\ and
rem arm64\netportal.exe. Run build.bat first. Needs Inno Setup 6
rem (winget install JRSoftware.InnoSetup).

setlocal
cd /d %~dp0 || exit /b 1

for /f "tokens=3" %%a in ('findstr /r "Number.*=.*[0-9.]*" ..\version\version.go') do set VERSION=%%a
set VERSION=%VERSION:"=%

set ISCC=%LOCALAPPDATA%\Programs\Inno Setup 6\ISCC.exe
if not exist "%ISCC%" set ISCC=%ProgramFiles(x86)%\Inno Setup 6\ISCC.exe
if not exist "%ISCC%" (
	echo [-] Inno Setup 6 not found. Install it with: winget install JRSoftware.InnoSetup
	exit /b 1
)

for %%a in (amd64 x86 arm64) do if not exist ..\%%a\netportal.exe (
	echo [-] ..\%%a\netportal.exe is missing. Run build.bat first.
	exit /b 1
)

echo [+] Building installer %VERSION%
"%ISCC%" /Q /DAppVersion=%VERSION% netportal.iss || exit /b 1
echo [+] Done: setup\out\NetPortal-Setup-%VERSION%.exe
