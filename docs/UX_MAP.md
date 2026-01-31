# UX Map — Dashboard Routes & Page Archetypes

Notion-like, content-first UI. All dashboard routes use a shared **page frame** (PageTemplate): title, description, actions, optional filter row, and content. Empty/loading/error states are consistent per archetype.

## Page archetypes

| Archetype | Use case | Empty | Loading | Error |
|-----------|----------|-------|---------|-------|
| **dashboard** | Overview: metrics, charts, activity | "No data yet" — run search/query | Spinner + "Loading…" | ErrorState + retry |
| **search** | Search/explore: semantic, warehouse, catalog, KG | "Nothing to show" — try search/catalog | Spinner + "Loading…" | ErrorState + retry |
| **builder** | Visual builder: workflows, notion blocks | "Start from scratch" — add nodes/blocks | Spinner + "Loading…" | ErrorState + retry |
| **list-detail** | Admin/list: users, API keys, integrations, settings | "No items yet" — create first item | Spinner + "Loading…" | ErrorState + retry |

## Route inventory

| Route | Archetype | PageTemplate | Notes |
|-------|-----------|--------------|-------|
| `/` | dashboard | ✓ (via layout) | Metrics, charts, ActivityFeed, QuickActions |
| `/analytics` | dashboard | ✓ | Analytics overview |
| `/observability` | dashboard | ✓ | Monitoring, traces |
| `/compliance` | dashboard | ✓ | Policies, compliance |
| `/quality` | dashboard | ✓ | Data quality |
| `/billing` | dashboard | ✓ | Billing |
| `/semantic` | search | ✓ | Search & Chat, Notes tabs |
| `/warehouse` | search | ✓ | Query, history, insights |
| `/catalog` | search | ✓ | Dataset list/detail |
| `/knowledge-graph` | search | ✓ | KG explore |
| `/data-sources` | search | ✓ | Connectors |
| `/metrics` | search | ✓ | Business metrics |
| `/workflows` | list-detail | ✓ | Workflow list |
| `/workflows/builder` | builder | ✓ | WorkflowBuilder (lazy) |
| `/notion-ui` | builder | ✓ | Blocks, databases |
| `/users` | list-detail | ✓ | User list, Add user |
| `/api-keys` | list-detail | ✓ | API key list, Create |
| `/integrations` | list-detail | ✓ | Integrations, webhooks |
| `/settings` | list-detail | ✓ | Org, preferences |
| `/audit` | list-detail | ✓ | Audit logs |
| `/alerts` | list-detail | ✓ | Alerts |
| `/lineage` | list-detail | ✓ | Lineage |
| `/versioning` | list-detail | ✓ | Versioning |
| `/support` | list-detail | ✓ | Support |
| `/features` | list-detail | ✓ | Features |
| `/why-neuronip` | list-detail | ✓ | Why NeuronIP |
| `/agents` | list-detail | ✓ | Agent Hub |
| `/models` | list-detail | ✓ | Models |

## Page frame (PageTemplate)

- **Title** — `text-2xl` / `text-3xl`, bold.
- **Description** — optional, `text-muted-foreground` below title.
- **Actions** — primary actions (e.g. "Add User", "Create API Key") in header row.
- **Filter row** — optional search/filters below header.
- **Content** — main area. Use `status`: `idle` | `loading` | `empty` | `error` for consistent states.

## Interaction

- **Motion**: Subtle, purposeful (variants in `lib/animations/variants`). Use `duration-fast` / `duration-base` for micro-interactions.
- **Inline edit**: Use `InlineEditable` for rename-on-double-click (e.g. list labels). Enter/blur saves; Escape cancels.
- **Keyboard-first**: ⌘K command palette, ⌘B sidebar, ⌘G dashboard, ⌘? shortcuts. Focus-visible rings, skip link.

## Performance

- **React Query**: Single `QueryClient` in `Providers`; no duplication in `DashboardLayout`.
- **Lazy-load**: `WorkflowBuilder` (ReactFlow), `LazyBlockEditor` (TipTap), `LazyLineChart` / `LazyBarChart` etc. via `next/dynamic`.
- **Virtualize**: `DataTable` uses `@tanstack/react-virtual` when `virtualize` and `data.length >= virtualizeThreshold` (default 50). Set `virtualize` and `virtualizeThreshold` for large lists.
- **Bundle analysis**: `npm run build:analyze` (`ANALYZE=true next build`).

## Accessibility & quality gates

- **Focus & keyboard**: `focus-visible` rings on interactive elements (globals.css fallback + component-level). Skip link to `#main-content`; main has `id="main-content"` and `tabIndex={-1}` for programmatic focus.
- **ARIA**: Form inputs use `aria-invalid`, `aria-describedby`; errors use `role="alert"`. Breadcrumbs `aria-label="Breadcrumb"`, `aria-current="page"`. Buttons/controls have `aria-label` where needed.
- **Loading / empty / error**: `PageTemplate` + `PageStates` per archetype (`PageStateLoading`, `PageStateEmpty`, `PageStateError`). Use `status` and overrides (`emptyTitle`, `errorMessage`, `onRetry`) consistently.

## Quality gates per archetype

- **Dashboard**: loading skeleton (metric cards + chart placeholders); empty "No data yet"; error + retry.
- **Search**: loading skeleton (bars + result area); empty "Nothing to show"; error + retry.
- **Builder**: loading spinner; empty "Start from scratch"; error + retry.
- **List-detail**: loading skeleton (list rows); empty "No items yet" + primary CTA; error + retry.
