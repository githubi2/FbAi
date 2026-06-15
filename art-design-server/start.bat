@echo off
cd /d "E:\FbAi\art-design-server"
echo Starting art-design-server...
start "art-design-server" server.exe
echo Server started on http://localhost:9090
echo Health check: http://localhost:9090/api/v1/ping
pause
