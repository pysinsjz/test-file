# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based data processing suite with multiple utilities for handling log files, user management, and Redis operations. The project consists of several standalone Go programs and shell scripts for processing various data formats and generating commands for database and Redis operations.

**Main Applications:**
1. **main.go** - Multi-function data processor with 5 different modes
2. **generate_redis_commands.go** - Generates Redis delete commands from Excel/CSV files
3. **dedup_unique_uids.go** - Deduplicates user IDs from CSV files
4. **Shell scripts** - Automation pipelines for complex workflows

## Common Commands

### Build and Run Main Application
```bash
go run main.go [function_number]
```

### Function Parameters (main.go)
- `go run main.go 1` - Parse log trace files in ./logs directory → data.csv
- `go run main.go 2` - Lock users from CSV files in ./lock-user-csv directory → lockUser-db_user库.sql + lockUser-redis_db0.txt
- `go run main.go 3` - Parse SQL logs in ./sql-log directory → sql.log (with intelligent deduplication)
- `go run main.go 4` - Split large files in multi-redis directory into 10K line chunks → multi-redis-split/
- `go run main.go 5` - Process KYC review files in kyc-review directory → kyc-YYYY-MM-DD.sql

### Utility Programs
```bash
# Generate Redis delete commands from del-ratio directory
go run generate_redis_commands.go

# Deduplicate UIDs from rm-repeat-uid/uid.csv
go run dedup_unique_uids.go

# Run complete Redis command generation pipeline
./run_del_ratio_pipeline.sh
```

### Dependencies
```bash
go mod tidy  # Install dependencies (primarily github.com/xuri/excelize/v2)
```

## Directory Structure

```
.
├── main.go                           # Multi-function data processor
├── generate_redis_commands.go        # Redis command generator
├── dedup_unique_uids.go              # UID deduplication utility
├── go.mod                            # Go module file
├── *.sh                             # Shell automation scripts
├── logs/                            # Input: Log trace files
├── lock-user-csv/                   # Input: User locking CSV files
├── sql-log/                         # Input: SQL log files
├── kyc-review/                      # Input: KYC Excel/CSV files
├── del-ratio/                       # Input: Excel/CSV for Redis deletion
├── multi-redis/                     # Input: Large files to split
├── multi-redis-split/               # Output: Split Redis command files
├── rm-repeat-uid/                   # Input/Output: UID deduplication
├── data.csv                         # Output: Parsed log trace data
├── lockUser-db_user库.sql           # Output: User locking SQL commands
├── lockUser-redis_db0.txt           # Output: User Redis delete commands
├── sql.log                          # Output: Deduplicated SQL statements
├── kyc-YYYY-MM-DD.sql               # Output: KYC approval SQL commands
└── redis_delete_commands.txt        # Output: Generated Redis commands
```

## Code Architecture

### Main Application (main.go)
Single-file architecture with multiple processing modes:

- **Signal Handling** - Graceful shutdown with resource cleanup
- **File Handle Management** - Automatic tracking and cleanup of open files
- **Processing Functions:**
  - `LogTaceParser()` - Extracts structured data from log files using fixed-position parsing
  - `lockUser()` - Generates SQL and Redis commands for user account locking
  - `sqlLogParser()` - Extracts and deduplicates SQL statements with intelligent key generation
  - `splitMultiRedisFile()` - Splits large command files into manageable chunks
  - `kycReviewProcessor()` - Processes KYC approval data from Excel/CSV files

### Utility Programs
- **generate_redis_commands.go** - Processes Excel/CSV files to generate Redis delete commands for turnover data
- **dedup_unique_uids.go** - Simple deduplication utility for user ID lists

### Key Features
- **Large File Handling** - 1MB buffer sizes for processing large log files
- **Format Support** - CSV and Excel (.xlsx) file processing
- **Error Resilience** - Continues processing despite individual file errors
- **Progress Reporting** - Detailed logging and progress indicators
- **Resource Management** - Automatic cleanup of file handles and memory