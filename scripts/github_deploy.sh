#!/bin/bash

###########################################
# GitHub Actions Deployment Script
#
# This script handles the deployment of the
# launchpad application to the production server.
###########################################

set -euo pipefail  # Exit on error, undefined vars, pipe failures
IFS=$'\n\t'        # Better word splitting

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Configuration
readonly PROJECT_DIR="${HOME}/launchpad"
readonly LOG_FILE="${PROJECT_DIR}/deploy.log"

# Logging functions
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $*" | tee -a "${LOG_FILE}"
}

log_error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR:${NC} $*" | tee -a "${LOG_FILE}" >&2
}

log_warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING:${NC} $*" | tee -a "${LOG_FILE}"
}

log_info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO:${NC} $*" | tee -a "${LOG_FILE}"
}

# Error handler
error_exit() {
    log_error "$1"
    log_error "Deployment failed"
    exit 1
}

# Cleanup function
cleanup() {
    log_info "Performing cleanup..."
    # Remove old images (keep last 3)
    docker image prune -a -f --filter "until=72h" || log_warn "Failed to prune old images"
}

# Main deployment function
deploy() {
    log "=========================================="
    log "Starting deployment process..."
    log "=========================================="

    # Change to project directory
    cd "${PROJECT_DIR}" || error_exit "Failed to change to project directory"

    # Check current git status
    log_info "Current branch: $(git branch --show-current)"
    log_info "Current commit: $(git rev-parse --short HEAD)"

    git fetch --all
    git reset --hard origin/main

    log_info "New commit: $(git rev-parse --short HEAD)"

    # Stop the application container
    log "Stopping application container..."
    if ! docker compose stop app; then
        log_warn "Failed to stop app container gracefully"
    fi

    # Build new image
    log "Building new Docker image..."
    if ! docker compose build --no-cache app; then
        error_exit "Failed to build Docker image"
    fi

    # Start the application
    log "Starting application..."
    if ! docker compose up -d app; then
        error_exit "Failed to start application"
    fi

    # Cleanup old resources
    cleanup

    log "=========================================="
    log "Deployment completed successfully!"
    log "=========================================="
    log_info "Logs: docker compose logs -f app"
}

# Trap errors
trap 'error_exit "Script interrupted or failed at line $LINENO"' ERR INT TERM

# Main execution
main() {
    # Ensure log directory exists
    mkdir -p "$(dirname "${LOG_FILE}")"

    # Run deployment
    deploy

    exit 0
}

# Execute main function
main "$@"
