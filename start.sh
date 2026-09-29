#!/bin/sh

# Проверяем наличие пакета tzdata
if ! apk info -e tzdata > /dev/null 2>&1; then
    echo "tzdata не найден. Устанавливаем..."
    apk update
    apk add --no-cache tzdata
    echo "tzdata установлен."
fi
export TZ="Asia/Yekaterinburg"
# Переходим в директорию и запускаем приложение
cd /opt/task-manager || { echo "Директория /opt/task-manager не найдена"; exit 1; }
echo "Запускаем cmd/main_linux..."
exec ./main_linux
