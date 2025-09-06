#!/bin/bash

# update_claude_md.sh - Update CLAUDE.md from iDesign project template
# Usage: ./update_claude_md.sh <repo_user> <repo_name> <branch> [file_path] [local_file]

set -euo pipefail  # Exit on error, undefined variables, and pipe failures

# Check for required parameters
if [[ $# -lt 3 ]]; then
    echo "Error: Missing required parameters"
    echo "Usage: $0 <repo_user> <repo_name> <branch> [file_path] [local_file]"
    echo ""
    echo "Parameters:"
    echo "  repo_user   - GitHub repository owner (required)"
    echo "  repo_name   - GitHub repository name (required)" 
    echo "  branch      - Branch to sync from (required)"
    echo "  file_path   - Path to file in source repo (optional, default: CLAUDE.md)"
    echo "  local_file  - Local destination file (optional, default: CLAUDE.md)"
    echo ""
    echo "Examples:"
    echo "  $0 rknuus idesign_project_template main"
    echo "  $0 rknuus idesign_project_template main CLAUDE.md CLAUDE.md"
    echo "  $0 myorg my_template develop docs/CLAUDE.md CLAUDE.md"
    exit 1
fi

# Configuration from parameters
REPO_USER="$1"
REPO_NAME="$2"
BRANCH="$3"
FILE_PATH="${4:-CLAUDE.md}"       # Default to CLAUDE.md if not provided
LOCAL_FILE="${5:-CLAUDE.md}"      # Default to CLAUDE.md if not provided
BACKUP_FILE="${LOCAL_FILE}.backup.$(date +%Y%m%d_%H%M%S)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# Function to check if required commands exist
check_dependencies() {
    if ! command -v gh &> /dev/null; then
        log_error "GitHub CLI (gh) is required but not found."
        log_error "Please install it:"
        log_error "  - macOS: brew install gh"
        log_error "  - Ubuntu: sudo apt install gh"
        log_error "  - Other: https://cli.github.com/manual/installation"
        exit 1
    fi

    if ! command -v git &> /dev/null; then
        log_error "Git is required but not found. Please install git first."
        exit 1
    fi

    if ! command -v curl &> /dev/null; then
        log_error "curl is required but not found. Please install curl first."
        exit 1
    fi
}

# Function to check if we're in a git repository
check_git_repo() {
    if ! git rev-parse --git-dir &> /dev/null; then
        log_error "Not in a git repository. Please run this script from the project root."
        exit 1
    fi
}

# Function to check and setup GitHub CLI authentication
ensure_gh_auth() {
    if gh auth status &> /dev/null; then
        local gh_user
        gh_user=$(gh api user --jq '.login' 2>/dev/null || echo "unknown")
        log_info "GitHub CLI authenticated as: $gh_user"
        return 0
    else
        log_warn "GitHub CLI not authenticated."
        log_info "Setting up GitHub CLI authentication..."

        if gh auth login --web; then
            log_info "GitHub CLI authentication successful!"
            return 0
        else
            log_error "GitHub CLI authentication failed"
            exit 1
        fi
    fi
}

# Function to backup existing file
backup_existing_file() {
    if [[ -f "$LOCAL_FILE" ]]; then
        log_info "Creating backup: $BACKUP_FILE"
        cp "$LOCAL_FILE" "$BACKUP_FILE" || {
            log_error "Failed to create backup of existing $LOCAL_FILE"
            exit 1
        }
    else
        log_warn "No existing $LOCAL_FILE found to backup"
    fi
}

# Function to download the file using GitHub CLI
download_file() {
    local api_url="https://api.github.com/repos/${REPO_USER}/${REPO_NAME}/contents/${FILE_PATH}?ref=${BRANCH}"
    local auth_header="Authorization: token $(gh auth token)"

    log_info "Downloading $FILE_PATH from GitHub API (branch: $BRANCH)..."

    # Build curl command for API incl. authentication header
    local curl_cmd="curl -sS -L --fail-with-body -H '$auth_header'"

    # Add Accept header for API
    curl_cmd="$curl_cmd -H 'Accept: application/vnd.github.v3.raw'"

    # Download the file
    if eval "$curl_cmd '$api_url'" > "$LOCAL_FILE"; then
        log_info "File downloaded successfully"
    else
        log_error "Failed to download $FILE_PATH from GitHub API"
        restore_backup_and_exit
    fi

    # Verify the file was downloaded and is not empty
    if [[ ! -s "$LOCAL_FILE" ]]; then
        log_error "Downloaded file is empty"
        restore_backup_and_exit
    fi
}

# Function to restore backup and exit on error
restore_backup_and_exit() {
    if [[ -f "$BACKUP_FILE" ]]; then
        log_info "Restoring backup..."
        mv "$BACKUP_FILE" "$LOCAL_FILE"
    fi
    exit 1
}

# Function to verify the downloaded file
verify_file() {
    # Basic verification - check if it looks like a CLAUDE.md file
    if ! grep -q "CLAUDE.md" "$LOCAL_FILE" 2>/dev/null; then
        log_warn "Downloaded file may not be a valid CLAUDE.md (missing expected header)"
    fi

    local line_count
    line_count=$(wc -l < "$LOCAL_FILE")
    log_info "Downloaded file has $line_count lines"
}

# Function to show diff if backup exists
show_changes() {
    if [[ -f "$BACKUP_FILE" ]] && command -v diff &> /dev/null; then
        log_info "Changes made:"
        if ! diff -u "$BACKUP_FILE" "$LOCAL_FILE" 2>/dev/null; then
            log_info "Differences shown above"
        else
            log_info "No differences found"
        fi
    fi
}

# Function to cleanup
cleanup() {
    if [[ -f "$BACKUP_FILE" ]]; then
        read -p "Keep backup file $BACKUP_FILE? (y/N): " -r
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            rm -f "$BACKUP_FILE"
            log_info "Backup file removed"
        else
            log_info "Backup file kept: $BACKUP_FILE"
        fi
    fi
}

# Function to show usage
show_usage() {
    cat << EOF
Usage: $0 <repo_user> <repo_name> <branch> [file_path] [local_file]

Update files from iDesign project template repository using GitHub CLI.

Parameters:
  repo_user   - GitHub repository owner (required)
  repo_name   - GitHub repository name (required) 
  branch      - Branch to sync from (required)
  file_path   - Path to file in source repo (optional, default: CLAUDE.md)
  local_file  - Local destination file (optional, default: CLAUDE.md)

Prerequisites:
- curl must be installed
- git must be installed
- GitHub CLI (gh) must be installed
- Authentication via 'gh auth login'

Examples:
  $0 rknuus idesign_project_template main
  $0 rknuus idesign_project_template main CLAUDE.md CLAUDE.md  
  $0 myorg my_template develop docs/CLAUDE.md CLAUDE.md
  gh auth login   # Setup authentication first if needed

EOF
}

# Main execution
main() {
    # Check for help flag among all arguments
    for arg in "$@"; do
        if [[ "$arg" == "-h" || "$arg" == "--help" ]]; then
            show_usage
            exit 0
        fi
    done

    log_info "Starting ${LOCAL_FILE} update from ${REPO_USER}/${REPO_NAME} (branch: ${BRANCH})..."

    check_dependencies
    check_git_repo
    ensure_gh_auth
    backup_existing_file
    download_file
    verify_file
    show_changes

    log_info "Successfully updated $LOCAL_FILE"
    log_info "Please review the changes and commit if appropriate"

    cleanup
}

# Trap to ensure cleanup on script exit
trap 'log_error "Script interrupted or failed"' ERR

# Run main function
main "$@"

