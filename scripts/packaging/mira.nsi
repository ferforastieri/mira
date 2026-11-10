; The installer of Mira Desktop for Windows, made by windows.sh with
; NSIS: makensis -DVERSION=… -DNUMERIC_VERSION=… -DSOURCE=<install folder>
; -DICON=<.ico> -DOUTFILE=<setup.exe> mira.nsi
;
; It installs for the current user, without administrator rights, into
; %LOCALAPPDATA%\Programs\Mira, adds Mira to the Start menu and to
; Settings › Apps, whose uninstaller removes the programs, the Start menu
; entry and the login item. The data folder, %LOCALAPPDATA%\Mira, is
; never touched. Run again, as an update does, it replaces the programs.
; /S installs, or uninstalls, silently.

Unicode true
ManifestDPIAware true
RequestExecutionLevel user
SetCompressor /SOLID lzma

!include "MUI2.nsh"
!include "FileFunc.nsh"

!define UNINSTALL_KEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\Mira"
!define RUN_KEY "Software\Microsoft\Windows\CurrentVersion\Run"
; The login item's value, as internal/desktop/loginitem names it.
!define LOGIN_ITEM "io.github.ferforastieri.mira-desktop"

Name "Mira"
OutFile "${OUTFILE}"
InstallDir "$LOCALAPPDATA\Programs\Mira"
BrandingText "Mira ${VERSION}"

VIProductVersion "${NUMERIC_VERSION}.0"
VIAddVersionKey /LANG=0 "ProductName" "Mira"
VIAddVersionKey /LANG=0 "ProductVersion" "${VERSION}"
VIAddVersionKey /LANG=0 "FileVersion" "${VERSION}"
VIAddVersionKey /LANG=0 "FileDescription" "Mira Desktop installer"
VIAddVersionKey /LANG=0 "LegalCopyright" "Mira contributors, MIT License"

!define MUI_ICON "${ICON}"
!define MUI_UNICON "${ICON}"
!define MUI_ABORTWARNING
!define MUI_FINISHPAGE_RUN "$INSTDIR\mira-desktop.exe"
!define MUI_FINISHPAGE_RUN_TEXT "$(StartMira)"

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"
!insertmacro MUI_LANGUAGE "French"

LangString StartMira ${LANG_ENGLISH} "Start Mira"
LangString StartMira ${LANG_FRENCH} "Lancer Mira"
LangString QuitFirst ${LANG_ENGLISH} "Mira is running. Quit it from its icon in the notification area, then choose Retry."
LangString QuitFirst ${LANG_FRENCH} "Mira est ouvert. Quittez-le depuis son icône dans la zone de notification, puis choisissez Recommencer."

; A program that runs cannot be replaced: Mira must be quit first.
!macro QuitMira
	retry:
	ClearErrors
	Delete "$INSTDIR\mira-desktop.exe"
	Delete "$INSTDIR\mira.exe"
	IfErrors 0 +3
		MessageBox MB_RETRYCANCEL|MB_ICONEXCLAMATION "$(QuitFirst)" /SD IDCANCEL IDRETRY retry
		Abort
!macroend

; The programs and their folders, never the data folder.
!macro RemovePrograms
	RMDir /r "$INSTDIR\postgresql"
	RMDir /r "$INSTDIR\jellyfin-web"
	RMDir /r "$INSTDIR\fonts"
	RMDir /r "$INSTDIR\licenses"
	Delete "$INSTDIR\mira.ico"
!macroend

Section
	SetShellVarContext current
	!insertmacro QuitMira
	!insertmacro RemovePrograms
	SetOutPath "$INSTDIR"
	File /r "${SOURCE}\*"
	WriteUninstaller "$INSTDIR\Uninstall.exe"
	CreateShortcut "$SMPROGRAMS\Mira.lnk" "$INSTDIR\mira-desktop.exe" "" "$INSTDIR\mira.ico"

	WriteRegStr HKCU "${UNINSTALL_KEY}" "DisplayName" "Mira"
	WriteRegStr HKCU "${UNINSTALL_KEY}" "DisplayVersion" "${VERSION}"
	WriteRegStr HKCU "${UNINSTALL_KEY}" "Publisher" "Mira"
	WriteRegStr HKCU "${UNINSTALL_KEY}" "URLInfoAbout" "https://github.com/ferforastieri/mira"
	WriteRegStr HKCU "${UNINSTALL_KEY}" "DisplayIcon" "$INSTDIR\mira.ico"
	WriteRegStr HKCU "${UNINSTALL_KEY}" "InstallLocation" "$INSTDIR"
	WriteRegStr HKCU "${UNINSTALL_KEY}" "UninstallString" '"$INSTDIR\Uninstall.exe"'
	WriteRegStr HKCU "${UNINSTALL_KEY}" "QuietUninstallString" '"$INSTDIR\Uninstall.exe" /S'
	WriteRegDWORD HKCU "${UNINSTALL_KEY}" "NoModify" 1
	WriteRegDWORD HKCU "${UNINSTALL_KEY}" "NoRepair" 1
	${GetSize} "$INSTDIR" "/S=0K" $0 $1 $2
	WriteRegDWORD HKCU "${UNINSTALL_KEY}" "EstimatedSize" $0
SectionEnd

Section "Uninstall"
	SetShellVarContext current
	!insertmacro QuitMira
	!insertmacro RemovePrograms
	Delete "$INSTDIR\Uninstall.exe"
	RMDir "$INSTDIR"
	Delete "$SMPROGRAMS\Mira.lnk"
	DeleteRegValue HKCU "${RUN_KEY}" "${LOGIN_ITEM}"
	DeleteRegKey HKCU "${UNINSTALL_KEY}"
SectionEnd
