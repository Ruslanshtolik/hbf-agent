@echo off
REM Quick Push Script - Easily commit and push changes to GitHub
REM Usage: scripts\quick-push.bat "Your commit message"

setlocal enabledelayedexpansion

REM Check if commit message is provided
if "%~1"=="" (
    echo No commit message provided. Using default message.
    for /f "tokens=1-4 delims=/ " %%a in ('date /t') do (set mydate=%%c-%%a-%%b)
    for /f "tokens=1-2 delims=: " %%a in ('time /t') do (set mytime=%%a:%%b)
    set COMMIT_MSG=Update: !mydate! !mytime!
) else (
    set COMMIT_MSG=%~1
)

echo === Quick Push to GitHub ===
echo.

echo Current status:
git status --short

echo.
echo Adding all changes...
git add .

echo.
echo Committing with message: !COMMIT_MSG!
git commit -m "!COMMIT_MSG!"
if errorlevel 1 (
    echo No changes to commit
    exit /b 0
)

echo.
echo Pushing to GitHub...
git push origin main

echo.
echo Successfully pushed to GitHub!
echo View at: https://github.com/Ruslanshtolik/hbf-agent
