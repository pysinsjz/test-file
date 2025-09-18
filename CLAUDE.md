# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go language log parsing tool with four main functions:
1. Log trace parsing - extracts key information from TXT log files and saves to data.log
2. User locking tool - reads user IDs from CSV files and generates SQL update statements and Redis delete commands
3. SQL log parsing - extracts SQL statements from log files and performs intelligent deduplication
4. KYC review processing - processes Excel/CSV files and generates SQL update statements for KYC records

## Directory Structure

```
.
├── main.go              # Main application with all functionality
├── CLAUDE.md            # This file
├── README.md            # Project documentation
├── test_cleanup.sh      # Test script
├── logs/                # Log files directory (function 1)
├── csv/                 # CSV files directory (function 2)
├── sql-log/             # SQL log files directory (function 3)
├── kyc-review/          # KYC review files directory (function 4)
├── data.log             # Function 1 output file
├── lockUser.sql         # Function 2 SQL output file
├── kyc-YYYY-MM-DD.sql   # Function 4 SQL output file (dated)
└── sql.log              # Function 3 output file
```

## Common Commands

### Build and Run
```bash
go run main.go [function_number]
```

### Run Tests
```bash
./test_cleanup.sh
```

### Function Parameters
- `go run main.go 1` - Parse log files in ./logs directory
- `go run main.go 2` - Lock users from CSV files in ./csv directory
- `go run main.go 3` - Parse SQL logs in ./sql-log directory
- `go run main.go 4` - Split files in multi-redis directory into 10K line chunks
- `go run main.go 5` - Process KYC review files in kyc-review directory

## Code Architecture

The application is structured as a single main.go file with multiple functions:

1. **Main Function** - Entry point that routes to different functionality based on command line arguments
2. **LogTaceParser()** - Handles log trace parsing functionality (function 1)
3. **lockUser()** - Handles user locking functionality (function 2)
4. **sqlLogParser()** - Handles SQL log parsing functionality (function 3)
5. **splitMultiRedisFile()** - Handles file splitting functionality (function 4)
6. **kycReviewProcessor()** - Handles KYC review processing functionality (function 5)
7. **Helper Functions** - Various utility functions for string parsing, SQL key generation, etc.

The code includes resource management features like automatic file handle tracking and signal handling for graceful shutdown.