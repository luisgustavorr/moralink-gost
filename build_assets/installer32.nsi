!include "MUI2.nsh"
!include "x64.nsh"

;-----------------------------------------
; General
;-----------------------------------------
Name              "MoraLink GOst (32-bit)"
OutFile           "../dist/moralink-setup-x86.exe"   ; <-- different output
InstallDir        "$PROGRAMFILES32\MoraLink\gost"     ; <-- 32-bit program files
InstallDirRegKey  HKLM "Software\MoraLink\GOst" "InstallDir"
RequestExecutionLevel admin
SetCompressor     /SOLID lzma

; ... version info and MUI settings identical to installer.nsi ...

Section "Install" SecInstall

  SetOutPath "$INSTDIR"

  ExecWait 'sc.exe stop moralink-gost'
  Sleep 2000

  ; Always use 32-bit binary — no detection needed
  File "..\dist\moralink-gost-windows-386.exe"        ; <-- only this binary
  Rename "$INSTDIR\moralink-gost-windows-386.exe" "$INSTDIR\moralink-gost.exe"

  CreateDirectory "$APPDATA\..\..\..\ProgramData\moralink-gost"
  CreateDirectory "$APPDATA\..\..\..\ProgramData\moralink-gost\logs"

  ExecWait '"$INSTDIR\moralink-gost.exe" install'
  ExecWait 'sc.exe start moralink-gost'

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