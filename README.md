## sre

A command line tool to help SREs (Site Reliability Engineers) do their job.

## Subcommands

| Name            | Description              | Status | Notes   |
|-----------------|--------------------------|--------|---------|
| `oncall`        | Print oncall information using PagerDuty's API | Mostly complete | Can show oncalls, escalation policies, schedules, and users |
| `omg`           | Print a user-defined first-response template | Done | The template uses Go's `text/template` package and can show links, images, and bold/italic text |
| `tools`         | Print a user-defined list of team tools | Done | It is just a reference for tools available to the team, no installation is performed |
| `schedule`      | Print information about an oncall schedule, given its PagerDuty schedule ID |"
| `incidents`     | Print incidents using PagerDuty's API | Basic implementation | Currently printing all the incidents that PagerDuty reports |
| `notifications` | Print notifications using PagerDuty's API | Basic implementation | Currently just printing all notifications reported by PagerDuty |
| `vpn`           | Connect to user-defined VPNs | Not implemented yet | Planning to support only Cisco AnyConnect through OpenConnect |
