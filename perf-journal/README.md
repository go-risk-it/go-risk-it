# Performance Optimization Journal

Track every optimization as a narrative with reproducible evidence.

## Current Status

| # | Name | E2E p95 | Throughput | Status |
|---|------|---------|------------|--------|
| 000 | Origin | — | — | pending |

## How to read

Each entry in `entries/` documents one optimization:
- **Hypothesis**: what we expect to improve and why
- **Change**: what was modified
- **Results**: before/after metrics from baselines in `baselines/`
- **Analysis**: what the numbers mean

Baselines in `baselines/` are named `NNN-<slug>-<commit>.json` with auto-incrementing sequence numbers. Any two baselines can be compared:

```bash
cd perf-test
go run ./cmd/loadtest --compare ../perf-journal/baselines/000-origin-abc1234.json \
  --save-baseline --baseline-name <name> --preset light --output text
```

## How to add an entry

1. Make your optimization change
2. Run the load test with `--save-baseline --baseline-name <slug>`
3. Copy `TEMPLATE.md` to `entries/NNN-<slug>.md`
4. Fill in hypothesis, results (from baseline comparison), and analysis
5. Update the status table above
