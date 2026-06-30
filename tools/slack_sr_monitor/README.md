## Slack SR Monitor Tool
The Slack SR Monitor tool is designed to monitor TRON Super Representatives (SRs) and notify a Slack channel after every maintenance period.
It automatically tracks vote changes and detects replacements in the top 27 SR positions, providing a clear and formatted report.

This directory also includes `slack_sr_guard`, a lightweight early-warning tool that polls the live SR list every minute and alerts Slack as soon as the current Top 27 membership changes.

### Build and Run
To run either tool, choose native Go execution or Docker deployment.

#### Native Go Execution
Make sure you have Go 1.25+ installed.
```shell
# enter the directory
cd tools/slack_sr_monitor
# install dependencies
go mod tidy
# run both tools by default
go run .

# run only the maintenance-period monitor
go run . monitor

# run only the early-warning guard
go run . guard
```

#### Docker Deployment
We provide a Docker-based deployment for easier management in production environments.
```shell
# build and start both tools
docker compose up -d --build

# build and start only one tool
docker compose up -d --build slack-sr-monitor
docker compose up -d --build slack-sr-guard

# check logs
docker logs -f slack-sr-monitor
docker logs -f slack-sr-guard
```

### Configuration
All configurations are managed via environment variables or a `.env` file in the project root. Please refer to [.env.example](./.env.example) as an example.

- `SLACK_WEBHOOK`: The Slack Incoming Webhook URL used to send notifications.
- `SLACK_SR_MONITOR_WEBHOOK`: Optional Slack webhook URL for `slack_sr_monitor`. Falls back to `SLACK_WEBHOOK` if unset.
- `SLACK_SR_GUARD_WEBHOOK`: Optional Slack webhook URL for `slack_sr_guard`. Falls back to `SLACK_WEBHOOK` if unset.
- `TRON_NODE`: The TRON node HTTP API endpoint (e.g., `http://https://api.trongrid.io`). Default is Trongrid.
- `TRON_NODES`: Optional comma-separated TRON node HTTP API endpoints. If set, the tools try each node in order for every TRON API request and only return an error after all nodes fail. This takes precedence over `TRON_NODE`.

### State Persistence

Both tools persist lightweight JSON state under `logs/`, which is already mounted by Docker Compose:

- `logs/sr_monitor_state.json`: previous vote counts and Top 27 snapshot for maintenance-period reports.
- `logs/sr_guard_state.json`: previous Top 27 snapshot and SR name cache for one-minute guard checks.

If these files are deleted, the next run starts with a fresh baseline and recreates them after the first successful fetch.

### Key Features

#### SR vote monitor
Use `/wallet/getpaginatednowwitnesslist` to get the top **28** real-time votes, also the SR address and URL.

#### Dynamic Scheduling
Instead of a fixed interval, the tool queries `/wallet/getnextmaintenancetime` to calculate the exact wait time. It triggers the report **1 minute** after each maintenance period begins to ensure data consistency.

#### Parallel Data Acquisition
The tool uses Go routines to fetch `account_name` for all 28 witnesses in parallel from the `/wallet/getaccount` interface, significantly reducing the collection time.

#### Vote Change Tracking
The tool persists a snapshot of the previous period's votes. It calculates the `Change` for each SR:
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

#### Top 27 Early Warning Guard
The `slack_sr_guard` process calls `/wallet/getpaginatednowwitnesslist` every minute with `limit=28`, keeps a persisted Top 27 snapshot, and sends a Slack alert only when an SR enters or leaves the Top 27 before the next maintenance period.

The alert includes the entered SRs, left SRs, current 27th/28th vote gap, and a Top 28 vote table.

### Notifications

This monitor only support java-tron node v4.8.1+, because of the API it used.
