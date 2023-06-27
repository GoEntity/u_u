@echo off
cd E:\Git\GoEntity_Github

"C:\Go\bin\go.exe" run main.go

"C:\Git\bin\git.exe" add .
"C:\Git\bin\git.exe" commit -m "auto update"
"C:\Git\bin\git.exe" push origin main