@echo off

REM dirrr
cd E:\Git\GoEntity_Github

REM setup git config
git config user.name "GoEntity"
git config user.email "goentity13@gmail.com"

REM run Go script
go run main.go

REM add, commit, and push changes
git add -A
for /f "tokens=2 delims==" %%I in ('wmic os get localdatetime /format:list') do set datetime=%%I
git commit -m "Latest data: %datetime%" || exit /b 0
git push
