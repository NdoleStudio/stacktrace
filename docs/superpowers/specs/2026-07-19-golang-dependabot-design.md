# Golang Dependabot Design

## Goal

Configure Dependabot to check the repository's root Go module for dependency
updates on the first day of every month.

## Configuration

Add `.github/dependabot.yml` using Dependabot configuration version 2. Define
one update entry with:

- `package-ecosystem: gomod`
- `directory: /`
- `schedule.interval: cron`
- `schedule.cron: "0 9 1 * *"`

The cron expression schedules checks for 09:00 UTC on the first day of each
month. The configuration will not group updates or impose additional pull
request limits, preserving Dependabot's default pull request behavior.

## Validation

Confirm that the YAML parses successfully and that the configuration contains
the required Dependabot version, Go module ecosystem, root directory, and exact
cron schedule.
