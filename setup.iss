#define AppName "Android Remote Control"
#define AppNameNoSpace "AndroidRemoteControl"
#define AppNameShort "goscrcpy"
#define AppExe "goscrcpy.exe"

[Setup]
AppName={#AppName}
AppVersion=1.1
WizardStyle=modern
DefaultDirName={autopf}\{#AppNameNoSpace}
DefaultGroupName={#AppName}
UninstallDisplayIcon={app}\{#AppExe}
Compression=lzma2
SolidCompression=yes
;Compression=none
OutputDir=.
OutputBaseFilename={#AppNameShort}-{#SetupSetting("AppVersion")}
SetupIconFile=icon.ico

[Run]
Filename: "{app}\{#AppExe}"; Description: "Launch application"; Flags: postinstall nowait skipifsilent

[Tasks]
Name: desktopicon; Description: "Create a &desktop icon"; GroupDescription: "Additional icons:"

[Files]
Source: "{#AppExe}"; DestDir: "{app}"
Source: "scrcpy\*"; DestDir: "{app}\scrcpy"

[Icons]
Name: "{group}\{#AppName}"; Filename: "{app}\{#AppExe}"
Name: "{commondesktop}\{#AppName}"; Filename: "{app}\{#AppExe}"; Tasks: desktopicon
