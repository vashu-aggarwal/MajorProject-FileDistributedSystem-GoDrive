@echo off
setlocal

echo ========================================
echo   GoDrive Pipeline Microservice
echo   LZW + AES-256-GCM + Base64
echo ========================================
echo.

:: Detect the Python executable available on this machine
set "PYTHON="
where python >nul 2>&1 && set "PYTHON=python"
if not defined PYTHON (
    where python3 >nul 2>&1 && set "PYTHON=python3"
)
if not defined PYTHON (
    where py >nul 2>&1 && set "PYTHON=py"
)
if not defined PYTHON (
    echo [ERROR] Python not found in PATH.
    echo Please install Python 3.9+ from https://www.python.org/downloads/
    echo Make sure to check "Add Python to PATH" during installation.
    pause
    exit /b 1
)

echo [OK] Using Python: %PYTHON%
%PYTHON% --version
echo.

:: Read port from config.yaml (default 8000)
set "PORT=8000"
set "AES_KEY="
set "CONFIG=..\config\config.yaml"

if exist "%CONFIG%" (
    for /f "tokens=1,* delims=: " %%A in ('findstr /i "port:" "%CONFIG%"') do (
        if /i "%%A"=="port" set "PORT=%%B"
    )
    for /f "tokens=1,* delims=: " %%A in ('findstr /i "aes_key_hex:" "%CONFIG%"') do (
        if /i "%%A"=="aes_key_hex" set "AES_KEY=%%B"
    )
    echo [OK] Config loaded from %CONFIG%
) else (
    echo [WARN] config.yaml not found, using defaults
)

:: Strip any trailing spaces or quotes from values
set PORT=%PORT: =%
set AES_KEY=%AES_KEY: =%
set AES_KEY=%AES_KEY:"=%

echo [Config] PORT    = %PORT%
if defined AES_KEY (
    echo [Config] AES KEY = loaded from config.yaml
) else (
    echo [Config] AES KEY = using built-in default
)
echo.

:: Install dependencies
echo Installing Python dependencies...
%PYTHON% -m pip install -q -r requirements.txt
if errorlevel 1 (
    echo [ERROR] Failed to install dependencies.
    pause
    exit /b 1
)
echo [OK] Dependencies ready.
echo.

:: Export env vars for app.py
set PORT=%PORT%
if defined AES_KEY set GODRIVE_AES_KEY_HEX=%AES_KEY%

echo Starting Pipeline Microservice on http://localhost:%PORT%
echo Press Ctrl+C to stop.
echo.
%PYTHON% app.py
