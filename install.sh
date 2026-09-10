#!/usr/bin/env bash
set -e

echo "=== Установщик и сборщик DCCAG ==="
echo ""

# 1. Проверка наличия Go
if ! command -v go &>/dev/null; then
  echo "⚠️  Компилятор Go не найден в системе."
  read -p "Хотите попробовать установить Go автоматически? (y/N): " install_go
  if [[ "$install_go" =~ ^[Yy]$ ]]; then
    if command -v apt-get &>/dev/null; then
      sudo apt-get update && sudo apt-get install -y golang
    elif command -v pacman &>/dev/null; then
      sudo pacman -Syu --noconfirm go
    elif command -v dnf &>/dev/null; then
      sudo dnf install -y golang
    elif command -v brew &>/dev/null; then
      brew install go
    else
      echo "❌ Пакетный менеджер не определен. Установите Go вручную: https://go.dev/dl/"
      exit 1
    fi
  else
    echo "Сборка прервана: требуется Go 1.18+."
    exit 1
  fi
fi

echo "✅ Go обнаружен: $(go version)"
echo ""

# 2. Подгрузка зависимостей
echo "📦 Загрузка зависимостей..."
go mod tidy

# 3. Выбор целевой платформы
echo ""
echo "Выберите целевую систему для сборки:"
echo "1) Текущая система ($(go env GOOS)/$(go env GOARCH))"
echo "2) Linux (x86_64 / amd64)"
echo "3) Linux (ARM64 / Raspberry Pi, Orange Pi)"
echo "4) Windows (64-bit .exe)"
echo "5) macOS (Apple Silicon M-серия, arm64)"
echo "6) macOS (Intel, amd64)"
echo "7) Собрать сразу ВСЕ варианты"
read -p "Ваш выбор [1-7] (по умолчанию 1): " choice

OUTPUT_DIR="build"
mkdir -p "$OUTPUT_DIR"

compile() {
  local target_os=$1
  local target_arch=$2
  local ext=""
  if [ "$target_os" = "windows" ]; then
    ext=".exe"
  fi
  local out_name="dccag-${target_os}-${target_arch}${ext}"

  echo "⚙️  Сборка под ${target_os}/${target_arch} -> ${OUTPUT_DIR}/${out_name}..."
  CGO_ENABLED=0 GOOS="$target_os" GOARCH="$target_arch" go build -ldflags="-s -w" -o "${OUTPUT_DIR}/${out_name}" .
}

case "$choice" in
2)
  compile "linux" "amd64"
  ;;
3)
  compile "linux" "arm64"
  ;;
4)
  compile "windows" "amd64"
  ;;
5)
  compile "darwin" "arm64"
  ;;
6)
  compile "darwin" "amd64"
  ;;
7)
  compile "linux" "amd64"
  compile "linux" "arm64"
  compile "windows" "amd64"
  compile "darwin" "arm64"
  compile "darwin" "amd64"
  ;;
*)
  echo "⚙️  Сборка под текущую ОС..."
  go build -ldflags="-s -w" -o "${OUTPUT_DIR}/dccag" .
  ;;
esac

echo ""
echo "🎉 Сборка успешно завершена! Файлы находятся в папке ./${OUTPUT_DIR}/"
