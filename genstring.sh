#!/bin/bash

# Check if python3-venv is installed
install_python3_venv() {
    echo "🔍 Checking if python3-venv is installed..."
    if ! python3 -m venv --help >/dev/null 2>&1; then
        echo "⚠️ python3-venv not found. Installing..."
        if [[ "$OSTYPE" == "linux-gnu"* ]]; then
            sudo apt update && sudo apt install -y python3-venv
        elif [[ "$OSTYPE" == "darwin"* ]]; then
            # Check if Homebrew is installed
            if ! command -v brew &>/dev/null; then
                echo "⚠️ Homebrew not found. Installing Homebrew..."
                /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
            fi
            echo "🍺 Installing python3-venv via Homebrew..."
            brew install python3
        else
            echo "❌ Unsupported operating system. Please install python3-venv manually."
            exit 1
        fi
    else
        echo "✅  python3-venv is already installed."
    fi
}

# Dependency installation
setup_environment() {
    echo "🔧 Setting up virtual environment..."
    python3 -m venv .venv
    source .venv/bin/activate

    echo "📦 Installing dependencies..."
    pip3 install --upgrade pip
    pip3 install telethon pyrogram TgCrypto
}

# Clear the terminal
clear_screen() {
    echo "🧹 Clearing the terminal..."
    clear
}

# Run the Python script
run_script() {
    echo "🚀 Running genstring.py..."
    python3 genstring.py
}

# Main flow execution
install_python3_venv
setup_environment
clear_screen
run_script
