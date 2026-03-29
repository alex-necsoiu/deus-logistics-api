#!/bin/bash
echo "🧹 Cleaning repository to reduce size..."

# Build artifacts
rm -rf bin/ dist/ tmp/ build/ 2>/dev/null
find . -name "*.exe" -type f -delete
find . -name "*.dylib" -type f -delete
find . -name "*.so" -type f -delete
find . -name "api" -type f -delete
find . -name "consumer" -type f -delete
find . -name "deus-logistics-api" -type f -delete

# Cache
rm -rf .cache/ vendor/ node_modules/ 2>/dev/null
find . -name "__pycache__" -type d -exec rm -rf {} + 2>/dev/null

# IDE files
rm -rf .vscode/ .idea/ 2>/dev/null

# OS files
find . -name ".DS_Store" -type f -delete
find . -name "Thumbs.db" -type f -delete

# Swap files
find . -name "*.swp" -type f -delete
find . -name "*.swo" -type f -delete

# Logs and test artifacts
find . -name "*.log" -type f -delete
find . -name "*.test" -type f -delete
rm -f coverage.out coverage.html *.prof 2>/dev/null

# Backups
find . -name "*.bak" -type f -delete
find . -name "*.tmp" -type f -delete
find . -name "*~" -type f -delete

# AI scratch files
find . -name "prompt_*.md" -type f -delete
find . -name "temp_*.md" -type f -delete
find . -name "response_*.txt" -type f -delete
find . -name "*.tmp.md" -type f -delete

# Git compression
git gc --aggressive --prune=now
git reflog expire --expire=now --all

echo "✅ Cleanup complete!"
du -sh .