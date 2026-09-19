@echo off
chcp 65001 >nul
echo ========================================================
echo   Iniciando Sistema de Soporte Tecnico Automotriz
echo ========================================================
echo.
echo 1. Levantando Backend en http://localhost:8080...
start "Backend - Taller Automotriz" cmd /k "cd /d "%~dp0backend" && set "PATH=%LOCALAPPDATA%\Programs\go\bin;%PATH%" && go run ./cmd/server"

echo 2. Levantando Frontend en http://localhost:4173...
start "Frontend - Taller Automotriz" cmd /k "cd /d "%~dp0frontend" && npm run dev"

echo.
echo ========================================================
echo Todo en marcha:
echo   - Frontend: http://localhost:4173
echo   - Backend:  http://localhost:8080
echo   - Usuario:  admin
echo   - Clave:    Admin2026*
echo ========================================================
timeout /t 5 >nul
