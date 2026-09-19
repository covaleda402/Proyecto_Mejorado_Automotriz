@echo off
chcp 65001 >nul
echo ========================================================
echo   Iniciando Backend (Go Server)
echo ========================================================
cd /d "%~dp0backend"
set "PATH=%LOCALAPPDATA%\Programs\go\bin;%PATH%"
go run ./cmd/server
if %ERRORLEVEL% neq 0 (
    echo.
    echo El backend se detuvo con un error.
    echo Verifica que la base de datos 'workshop' este creada ejecutando setup_database.bat
    pause
)
