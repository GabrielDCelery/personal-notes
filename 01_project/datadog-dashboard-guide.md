# Datadog Dashboard Guide

## Navigating Alerts

- **Monitors > Triggered Monitors** — currently firing alerts
- **Monitors > Manage Monitors** — all monitors (OK, alerting, no data)
- Add a **Monitor Summary** widget to a dashboard to show alerts inline

---

## Dragging Widgets

1. Enter **edit mode** (pencil icon at the top)
2. Click and hold the **top edge/header** of the widget
3. Drag to new position and release

---

## Visualizing Lambda Errors on a Dashboard

### Find Logs First

Go to **Logs > Search** and filter:

```
source:lambda status:error @aws.function_name:service-ice-api-*
```

### Recommended Widgets

| Widget      | Purpose                               |
| ----------- | ------------------------------------- |
| Query Value | Current error count                   |
| Timeseries  | Error trend over time                 |
| Top List    | Which lambdas/quotes are failing most |
| Log Stream  | Raw log details                       |

### Group Errors by Lambda

- Widget: **Top List** or **Timeseries**
- Query: `source:lambda status:error service:services-*`
- **Group by**: `@aws.function_name` or `@log.group` (if function name isn't parsed)

If `aws.function_name` isn't available, install the **Datadog Lambda Layer** on your Lambdas.

---

## Template Variables (Dynamic Filtering)

Add a template variable at the top of the dashboard:

- **Name**: `env`
- **Tag**: `env`
- **Default**: `production` or `*`

Use `$env` in widget queries:

```
sum:ice_api.lambda.errors{env:$env} by {aws_function_name}
```

For Lambda name:

- **Name**: `lambda`
- **Tag/Attribute**: `@aws.function_name`
- **Default**: `service-ice-api-*`

---

## Creating Log-Based Metrics

Go to **Logs > Generate Metrics** → **+ New Metric**

### Naming Convention

Use dots, no dashes: `ice_api.lambdaname.errors`

### Metric for All ICE API Lambda Errors

- **Filter**: `source:lambda status:error @aws.function_name:service-ice-api-*`
- **Name**: `ice_api.lambda.errors`
- **Count**: count
- **Group by**: `@aws.function_name`, `env`

### Metric for Log Level Ratio (single lambda)

Create 3 separate metrics:

| Metric                            | Filter                                                                       |
| --------------------------------- | ---------------------------------------------------------------------------- |
| `ice_api.specificname.logs.info`  | `source:lambda status:info @aws.function_name:service-ice-api-specificname`  |
| `ice_api.specificname.logs.warn`  | `source:lambda status:warn @aws.function_name:service-ice-api-specificname`  |
| `ice_api.specificname.logs.error` | `source:lambda status:error @aws.function_name:service-ice-api-specificname` |

All group by: `env`

Then use formulas on the dashboard widget:

- `a` = info metric, `b` = warn metric, `c` = error metric
- Formula: `a / (a + b + c) * 100` for percentage breakdown
- Or use **stacked bars** display for visual ratio

**Shortcut**: skip metrics entirely and use Log Analytics directly:

- Data source: **Logs**, group by `status`, display as stacked bars

---

## Metric Query Syntax

```
sum:metric_name{filter} by {group}
```

- `{}` — filters (narrows data), e.g. `{env:production}`
- `by {}` — groups (splits into series), e.g. `by {aws_function_name}`

Examples:

```
sum:ice_api.lambda.errors{env:production}
sum:ice_api.lambda.errors{env:$env} by {aws_function_name}
sum:aws.lambda.invocations{functionname:service-ice-api-specificname,env:$env}
```

---

## Lambda Invocation Count

Use the built-in AWS metric — no setup needed if AWS integration is configured:

| Metric                   | What it shows          |
| ------------------------ | ---------------------- |
| `aws.lambda.invocations` | Total invocation count |
| `aws.lambda.errors`      | AWS-level errors       |
| `aws.lambda.duration`    | Execution time         |
| `aws.lambda.throttles`   | Throttled invocations  |

Query example:

```
sum:aws.lambda.invocations{functionname:service-ice-api-specificname,env:$env}
```

Check availability under **Metrics > Summary**.

---

## Adding a Metric to a Dashboard

### From Metrics Explorer

1. **Metrics > Explorer** — build and preview your query
2. Click the **export icon** (top right) → **Export to Dashboard**
3. Select your dashboard → **Export**

### Directly on the Dashboard (faster)

1. Open dashboard → **+ Add Widgets**
2. Select widget type (e.g. Timeseries)
3. Type metric name in query field
4. Add filters and save

---

## Filters vs. Log-Based Metrics

|               | Filters on Dashboard | Log-Based Metrics              |
| ------------- | -------------------- | ------------------------------ |
| Setup speed   | Fast                 | Requires metric creation       |
| Alerting      | No                   | Yes                            |
| Log retention | Limited              | Long-term                      |
| Best for      | Exploring, debugging | Permanent dashboards, alerting |

**Recommendation**: start with filters on the dashboard to validate, then convert to a log-based metric for long-term use and alerting.
