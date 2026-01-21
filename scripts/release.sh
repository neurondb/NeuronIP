#!/bin/bash
# Release Script
# Automates version tagging, changelog generation, and release

set -e

VERSION_TYPE="${1:-patch}" # patch, minor, major

# Get current version
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
CURRENT_VERSION=${CURRENT_VERSION#v}

# Calculate new version
IFS='.' read -ra VERSION_PARTS <<< "$CURRENT_VERSION"
MAJOR=${VERSION_PARTS[0]}
MINOR=${VERSION_PARTS[1]}
PATCH=${VERSION_PARTS[2]}

case $VERSION_TYPE in
  major)
    MAJOR=$((MAJOR + 1))
    MINOR=0
    PATCH=0
    ;;
  minor)
    MINOR=$((MINOR + 1))
    PATCH=0
    ;;
  patch)
    PATCH=$((PATCH + 1))
    ;;
  *)
    echo "Invalid version type: $VERSION_TYPE (use: major, minor, patch)"
    exit 1
    ;;
esac

NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"

echo "Current version: v${CURRENT_VERSION}"
echo "New version: ${NEW_VERSION}"
read -p "Continue? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  exit 1
fi

# Update CHANGELOG.md
echo "Updating CHANGELOG.md..."
# This would add a new section to CHANGELOG.md
# For now, placeholder

# Create git tag
echo "Creating git tag: ${NEW_VERSION}"
git tag -a "${NEW_VERSION}" -m "Release ${NEW_VERSION}"

# Push tag
echo "Pushing tag to remote..."
git push origin "${NEW_VERSION}"

echo ""
echo "Release ${NEW_VERSION} created successfully!"
echo "GitHub Actions will build and publish the release."
