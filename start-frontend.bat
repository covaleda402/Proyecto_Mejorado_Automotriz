@echo off
chcp 65001 >nul
echo Iniciando Frontend (React + Vite)...
cd /d "%~dp0frontend"
npm run dev
pause
