[Setup]
AppName=Android Remote Control
AppVersion=1.0
WizardStyle=modern
DefaultDirName={autopf}\AndroidRemoteControl
DefaultGroupName=Android Remote Control
UninstallDisplayIcon={app}\goscrcpy.exe
Compression=lzma2
SolidCompression=yes
OutputDir=.
OutputBaseFilename=goscrcpy-{#SetupSetting("AppVersion")}

[Files]
Source: "goscrcpy.exe"; DestDir: "{app}"
Source: "scrcpy\*"; DestDir: "{app}\scrcpy"

[Icons]
Name: "{group}\Android Remote Control"; Filename: "{app}\goscrcpy.exe"
