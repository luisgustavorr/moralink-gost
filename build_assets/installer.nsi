!include "MUI2.nsh"
!include "x64.nsh"

;-----------------------------------------
; General
;-----------------------------------------
Name              "MoraLink GOst"
OutFile           "../dist/moralink-setup.exe"
InstallDir        "$PROGRAMFILES64\MoraLink\gost"
InstallDirRegKey  HKLM "Software\MoraLink\GOst" "InstallDir"
RequestExecutionLevel admin
SetCompressor     /SOLID lzma

;-----------------------------------------
; Version info (shows in file properties)
;-----------------------------------------
VIProductVersion  "0.0.1.0"
VIAddVersionKey   "ProductName"     "MoraLink GOst"
VIAddVersionKey   "CompanyName"     "ORBIS SOLUCOES TECNOLOGICAS"
VIAddVersionKey   "FileDescription" "MoraLink GOst Installer"
VIAddVersionKey   "FileVersion"     "0.0.1"
VIAddVersionKey   "LegalCopyright"  "© ORBIS SOLUCOES TECNOLOGICAS"

;-----------------------------------------
; MUI Settings
;-----------------------------------------
!define MUI_ABORTWARNING
!define MUI_ICON   "icon.ico"
!define MUI_UNICON "icon.ico"

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "Portuguese"

;-----------------------------------------
; Install Section
;-----------------------------------------
Section "Install" SecInstall

  SetOutPath "$INSTDIR"

  ; Stop the service if already running (upgrade flow)
  ExecWait 'sc.exe stop moralink-gost'
  Sleep 2000

  ; Copy binary and config
  File "..\dist\moralink-gost-windows-amd64.exe"
  Rename "$INSTDIR\moralink-gost-windows-amd64.exe" "$INSTDIR\moralink-gost.exe"

  ; Create ProgramData directories upfront with correct permissions
  ; so the service (SYSTEM account) can always write logs/config
  CreateDirectory "$APPDATA\..\..\..\ProgramData\moralink-gost"
  CreateDirectory "$APPDATA\..\..\..\ProgramData\moralink-gost\logs"

  ; Register & start the service
  ExecWait '"$INSTDIR\moralink-gost.exe" install'
  ExecWait 'sc.exe start moralink-gost'

  ; Write uninstall info to registry
  WriteRegStr   HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\MoraLinkGOst" \
                "DisplayName"     "MoraLink GOst"
  WriteRegStr   HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\MoraLinkGOst" \
                "UninstallString" '"$INSTDIR\uninstall.exe"'
  WriteRegStr   HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\MoraLinkGOst" \
                "DisplayVersion"  "0.0.1"
  WriteRegStr   HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\MoraLinkGOst" \
                "Publisher"       "ORBIS SOLUCOES TECNOLOGICAS"
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\MoraLinkGOst" \
                "NoModify" 1
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\MoraLinkGOst" \
                "NoRepair" 1

  WriteUninstaller "$INSTDIR\uninstall.exe"

SectionEnd

;-----------------------------------------
; Uninstall Section
;-----------------------------------------
Section "Uninstall"

  ExecWait 'sc.exe stop moralink-gost'
  Sleep 2000
  ExecWait '"$INSTDIR\moralink-gost.exe" uninstall'
  Sleep 1000

  Delete "$INSTDIR\moralink-gost.exe"
  Delete "$INSTDIR\uninstall.exe"
  RMDir  "$INSTDIR"

  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\MoraLinkGOst"
  DeleteRegKey HKLM "Software\MoraLink\GOst"

SectionEnd