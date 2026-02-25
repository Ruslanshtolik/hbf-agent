#!/bin/bash
# Quick Push Script - Easily commit and push changes to GitHub
# Usage: ./scripts/quick-push.sh "Your commit message"

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if commit message is provided
if [ -z "$1" ]; then
    echo -e "${YELLOW}No commit message provided. Using default message.${NC}"
    COMMIT_MSG="Update: $(date '+%Y-%m-%d %H:%M:%S')"
else
    COMMIT_MSG="$1"
fi

echo -e "${BLUE}=== Quick Push to GitHub ===${NC}"
echo ""

# Show current status
echo -e "${BLUE}Current status:${NC}"
git status --short

echo ""
echo -e "${BLUE}Adding all changes...${NC}"
git add .

echo ""
echo -e "${BLUE}Committing with message: ${GREEN}${COMMIT_MSG}${NC}"
git commit -m "$COMMIT_MSG" || {
    echo -e "${YELLOW}No changes to commit${NC}"
    exit 0
}

echo ""
echo -e "${BLUE}Pushing to GitHub...${NC}"
git push origin main

echo ""
echo -e "${GREEN}✓ Successfully pushed to GitHub!${NC}"
echo -e "${GREEN}✓ View at: https://github.com/Ruslanshtolik/hbf-agent${NC}"
