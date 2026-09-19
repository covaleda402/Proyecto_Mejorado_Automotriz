@echo off
chcp 65001 >nul
echo ========================================================
echo   Configuración de la Base de Datos Workshop (MySQL)
echo ========================================================
echo.
echo Este script creará la base de datos 'workshop', el usuario,
echo las tablas y cargará los datos de prueba (administrador y técnicos).
echo.
echo Por favor ingresa la contraseña de tu usuario 'root' de MySQL:
set /p MYSQL_ROOT_PASS="Contraseña de root: "

echo.
echo Ejecutando scripts SQL...
"C:\Program Files\MySQL\MySQL Server 8.0\bin\mysql.exe" -u root -p"%MYSQL_ROOT_PASS%" < "%~dp0database\init_complete_db.sql"

if %ERRORLEVEL% equ 0 (
    echo.
    echo [EXITO] Base de datos 'workshop' inicializada correctamente.
) else (
    echo.
    echo [ERROR] Hubo un error al ejecutar el script. Verifica tu contraseña de root.
)

echo.
pause
