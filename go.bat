@echo off
setlocal EnableExtensions EnableDelayedExpansion

set "GOEXE=C:\Program Files\Go\bin\go.exe"
set "ORIGINAL_ARGS=%*"
if not defined GOTMPDIR set "GOTMPDIR=%~dp0.gotmp"
if not exist "%GOTMPDIR%" mkdir "%GOTMPDIR%"

if /I not "%~1"=="test" goto passthrough

shift
set "MODE=passthrough"
set "TEST_FLAGS="
set "BINARY_FLAGS="

:scan_args
if "%~1"=="" goto dispatch
if /I "%~1"=="./testing" set "MODE=testing_only" & shift & goto scan_args
if /I "%~1"==".\testing" set "MODE=testing_only" & shift & goto scan_args
if /I "%~1"=="./testing/..." set "MODE=testing_only" & shift & goto scan_args
if /I "%~1"==".\testing\..." set "MODE=testing_only" & shift & goto scan_args
if /I "%~1"=="./..." set "MODE=all_with_stable_testing" & shift & goto scan_args
if /I "%~1"==".\..." set "MODE=all_with_stable_testing" & shift & goto scan_args

if /I "%~1"=="-v" (
 set "TEST_FLAGS=!TEST_FLAGS! -v"
 set "BINARY_FLAGS=!BINARY_FLAGS! -test.v"
 shift
 goto scan_args
)
if /I "%~1"=="-short" (
 set "TEST_FLAGS=!TEST_FLAGS! -short"
 set "BINARY_FLAGS=!BINARY_FLAGS! -test.short"
 shift
 goto scan_args
)
if /I "%~1"=="-failfast" (
 set "TEST_FLAGS=!TEST_FLAGS! -failfast"
 set "BINARY_FLAGS=!BINARY_FLAGS! -test.failfast"
 shift
 goto scan_args
)
if /I "%~1"=="-run" (
 set "TEST_FLAGS=!TEST_FLAGS! -run %~2"
 set "BINARY_FLAGS=!BINARY_FLAGS! -test.run %~2"
 shift
 shift
 goto scan_args
)
if /I "%~1"=="-list" (
 set "TEST_FLAGS=!TEST_FLAGS! -list %~2"
 set "BINARY_FLAGS=!BINARY_FLAGS! -test.list %~2"
 shift
 shift
 goto scan_args
)
if /I "%~1"=="-timeout" (
 set "TEST_FLAGS=!TEST_FLAGS! -timeout %~2"
 set "BINARY_FLAGS=!BINARY_FLAGS! -test.timeout %~2"
 shift
 shift
 goto scan_args
)

set "ARG=%~1"
if /I "!ARG:~0,5!"=="-run=" (
 set "TEST_FLAGS=!TEST_FLAGS! %~1"
 set "BINARY_FLAGS=!BINARY_FLAGS! -test.run=!ARG:~5!"
 shift
 goto scan_args
)
if /I "!ARG:~0,6!"=="-list=" (
 set "TEST_FLAGS=!TEST_FLAGS! %~1"
 set "BINARY_FLAGS=!BINARY_FLAGS! -test.list=!ARG:~6!"
 shift
 goto scan_args
)
if /I "!ARG:~0,9!"=="-timeout=" (
 set "TEST_FLAGS=!TEST_FLAGS! %~1"
 set "BINARY_FLAGS=!BINARY_FLAGS! -test.timeout=!ARG:~9!"
 shift
 goto scan_args
)

set "TEST_FLAGS=!TEST_FLAGS! %~1"

shift
goto scan_args

:dispatch
if /I "%MODE%"=="testing_only" goto run_testing_only
if /I "%MODE%"=="all_with_stable_testing" goto run_all_with_stable_testing
goto passthrough

:run_all_with_stable_testing
set "PACKAGE_ARGS="
for /f "delims=" %%A in ('""%GOEXE%" list ./..."') do (
 if /I not "%%A"=="gut/testing" set "PACKAGE_ARGS=!PACKAGE_ARGS! %%A"
)
if defined PACKAGE_ARGS (
 "%GOEXE%" test!TEST_FLAGS! !PACKAGE_ARGS!
 if errorlevel 1 exit /b !ERRORLEVEL!
)
goto run_stable_testing_binary

:run_testing_only
goto run_stable_testing_binary

:run_stable_testing_binary
set "STABLE_DIR=%~dp0.artifacts\go-test"
if not exist "%STABLE_DIR%" mkdir "%STABLE_DIR%"
set "STABLE_BIN=%STABLE_DIR%\gut-testing.exe"
"%GOEXE%" test -c ./testing -o "%STABLE_BIN%"
if errorlevel 1 exit /b !ERRORLEVEL!
call "%STABLE_BIN%"!BINARY_FLAGS!
exit /b !ERRORLEVEL!

:passthrough
"%GOEXE%" %ORIGINAL_ARGS%
exit /b !ERRORLEVEL!
