; NET PORTAL installer (Inno Setup 6).
; Build with setup\build.bat after build.bat; it passes AppVersion.

#ifndef AppVersion
  #define AppVersion "0.0.0"
#endif

[Setup]
AppId={{6F1E3C2A-8B7D-4C5E-9A1F-2D3B4C5E6F70}
AppName=NET PORTAL
AppVersion={#AppVersion}
AppVerName=NET PORTAL {#AppVersion}
AppPublisher=ZipFileZ
AppPublisherURL=https://github.com/ZipFileZ/netportal
AppSupportURL=https://github.com/ZipFileZ/netportal/issues
AppUpdatesURL=https://github.com/ZipFileZ/netportal/releases
VersionInfoVersion={#AppVersion}
; The manager keeps its data (tunnel configurations, log) in
; Program Files\NetPortal\Data, so the program lives in the same folder.
DefaultDirName={autopf}\NetPortal
DisableDirPage=yes
DefaultGroupName=NET PORTAL
DisableProgramGroupPage=yes
PrivilegesRequired=admin
ArchitecturesAllowed=x86compatible x64compatible arm64
ArchitecturesInstallIn64BitMode=x64compatible arm64
MinVersion=10.0
OutputDir=out
OutputBaseFilename=NetPortal-Setup-{#AppVersion}
SetupIconFile=..\webui\icon\netportal.ico
UninstallDisplayIcon={app}\netportal.exe
UninstallDisplayName=NET PORTAL
WizardStyle=modern
Compression=lzma2/max
SolidCompression=yes
CloseApplications=no
LicenseFile=..\COPYING

[Languages]
Name: "ru"; MessagesFile: "compiler:Languages\Russian.isl"
Name: "en"; MessagesFile: "compiler:Default.isl"

[CustomMessages]
ru.DesktopIcon=Создать значок на рабочем столе
en.DesktopIcon=Create a desktop shortcut
ru.LaunchApp=Запустить NET PORTAL
en.LaunchApp=Launch NET PORTAL
ru.RemoveData=Удалить также туннели, журнал и настройки NET PORTAL?%n%nЕсли вы собираетесь установить программу снова, нажмите «Нет» — туннели сохранятся.
en.RemoveData=Also delete NET PORTAL tunnels, log and settings?%n%nChoose No if you plan to install it again — your tunnels will be kept.
ru.StoppingServices=Останавливаем службы NET PORTAL…
en.StoppingServices=Stopping NET PORTAL services…

[Tasks]
Name: "desktopicon"; Description: "{cm:DesktopIcon}"; Flags: unchecked

[Files]
; One executable for the machine's architecture.
Source: "..\arm64\netportal.exe"; DestDir: "{app}"; Flags: ignoreversion; Check: IsArm64
Source: "..\amd64\netportal.exe"; DestDir: "{app}"; Flags: ignoreversion; Check: IsX64OS
Source: "..\x86\netportal.exe";   DestDir: "{app}"; Flags: ignoreversion; Check: IsX86OS
Source: "..\COPYING"; DestDir: "{app}"; DestName: "LICENSE.txt"; Flags: ignoreversion

[Icons]
Name: "{autoprograms}\NET PORTAL"; Filename: "{app}\netportal.exe"
Name: "{autodesktop}\NET PORTAL"; Filename: "{app}\netportal.exe"; Tasks: desktopicon

[Run]
; Registers and starts the manager service, which then opens the window.
Filename: "{app}\netportal.exe"; Parameters: "/installmanagerservice"; Description: "{cm:LaunchApp}"; Flags: postinstall nowait

[UninstallDelete]
; Left over by self-updates (the replaced executable).
Type: files; Name: "{app}\netportal.exe.old"
Type: files; Name: "{app}\netportal.exe.new"

[Code]
// Stops (and with Delete removes) only NET PORTAL's own services and
// processes: NetPortalManager, NetPortalTunnel$<name> and netportal.exe.
// AmneziaWG's services have different names and are never touched.
procedure StopNetPortal(Delete: Boolean);
var
  Script: String;
  ResultCode: Integer;
begin
  Script := 'Get-Service -Name ''NetPortalManager'',''NetPortalTunnel$*'' -ErrorAction SilentlyContinue | ForEach-Object { Stop-Service -Name $_.Name -Force -ErrorAction SilentlyContinue';
  if Delete then
    Script := Script + '; & sc.exe delete $_.Name | Out-Null';
  Script := Script + ' }; Get-Process -Name netportal -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue';
  Exec(ExpandConstant('{sys}\WindowsPowerShell\v1.0\powershell.exe'),
    '-NoProfile -NonInteractive -ExecutionPolicy Bypass -Command "' + Script + '"',
    '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
end;

// Before installing over an existing copy, stop it so its files can be
// replaced. The manager comes back when the installer finishes.
function PrepareToInstall(var NeedsRestart: Boolean): String;
begin
  WizardForm.StatusLabel.Caption := CustomMessage('StoppingServices');
  StopNetPortal(False);
  Result := '';
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usUninstall then
    StopNetPortal(True);
  if CurUninstallStep = usPostUninstall then
  begin
    if not UninstallSilent and
       (MsgBox(CustomMessage('RemoveData'), mbConfirmation, MB_YESNO or MB_DEFBUTTON2) = IDYES) then
    begin
      DelTree(ExpandConstant('{app}\Data'), True, True, True);
      DelTree(ExpandConstant('{userappdata}\NetPortal'), True, True, True);
      DelTree(ExpandConstant('{localappdata}\NetPortal'), True, True, True);
    end;
    RemoveDir(ExpandConstant('{app}'));
  end;
end;
