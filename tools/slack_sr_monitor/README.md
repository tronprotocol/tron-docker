## Slack SR Monitor Tool
The Slack SR Monitor tool is designed to monitor TRON Super Representatives (SRs) and notify a Slack channel after every maintenance period.
It automatically tracks vote changes and detects replacements in the top 27 SR positions, providing a clear and formatted report.

### Build and Run the monitor
To run the monitor tool, you can choose between native Go execution or Docker deployment.

#### Native Go Execution
Make sure you have Go 1.25+ installed.
```shell
# enter the directory
cd tools/slack_sr_monitor
# install dependencies
go mod tidy
# run the tool
go run main.go
```

#### Docker Deployment
We provide a Docker-based deployment for easier management in production environments.
```shell
# build and start the container
docker-compose up -d --build
# check logs
docker logs -f slack-sr-monitor
```

### Configuration
All configurations are managed via environment variables or a `.env` file in the project root. Please refer to [.env.example](./.env.example) as an example.

- `SLACK_WEBHOOK`: The Slack Incoming Webhook URL used to send notifications.
- `TRON_NODE`: The TRON node HTTP API endpoint (e.g., `http://https://api.trongrid.io`). Default is Trongrid.

### Key Features

#### SR vote monitor
Use `/wallet/getpaginatednowwitnesslist` to get the top **28** real-time votes, also the SR address and URL.

#### Dynamic Scheduling
Instead of a fixed interval, the tool queries `/wallet/getnextmaintenancetime` to calculate the exact wait time. It triggers the report **1 minute** after each maintenance period begins to ensure data consistency.

#### Parallel Data Acquisition
The tool uses Go routines to fetch `account_name` for all 28 witnesses in parallel from the `/wallet/getaccount` interface, significantly reducing the collection time.

#### Vote Change Tracking
The tool maintains an in-memory snapshot of the previous period's votes. It calculates the `Change` for each SR:
```text
*1. Poloniex*
Current: `3,228,089,488`  Change: `+89,488`
```

#### Top 27 Replacement Detection
After each report, it compares the current Top 27 list with the previous one and highlights any changes:
```text
SR Replacement Detected:
>:inbox_tray: *Entered:* New_SR_Name
>:outbox_tray: *Left:* Old_SR_Name
```
If no changes occur, it displays `Top 27 SRs remain unchanged.`

### Notifications

This monitor only support java-tron node v4.8.1+, because of the API it used.
