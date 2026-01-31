# NeuronIP Full-Stack Demo

Single demo script and acceptance checklist to validate NeuronIP end-to-end: warehouse, workload queues, data products, notebooks, decision dashboards, ITSM, knowledge workspace templates, and AI assistant.

## Prerequisites

- NeuronIP API running (e.g. `./run_neuronip.sh`)
- Valid API key in `NEURONIP_API_KEY` or use `test-key-82f13cedd19abec5bdd9ffad70f3f774` for demo
- `curl` for the script

## Run the demo

```bash
export API_BASE="${API_BASE:-http://localhost:8082/api/v1}"
export NEURONIP_API_KEY="${NEURONIP_API_KEY:-test-key-82f13cedd19abec5bdd9ffad70f3f774}"

./demo-all.sh
```

## Files

| File | Purpose |
|------|---------|
| `demo-all.sh` | Single script that exercises warehouse, workload, data products, notebooks, decision dashboards, ITSM, templates, and AI assistant |
| `CHECKLIST.md` | Acceptance checklist and SLOs for all flows |

## Acceptance criteria (summary)

- Warehouse query returns 200 and result shape; workload queue create/list 200; data product create 201.
- Notebook create 201; add cell 201; create run 201.
- Decision dashboard create 201; list 200; record run 201.
- Incident create 201; change create 201; runbook create/list 200.
- List page/database templates 200; AI assistant (RAG) 200 with answer/sources.
