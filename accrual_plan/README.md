# Accrual Accounting Calendar Feature Plan

This folder contains a series of self-contained, comprehensive implementation phases to extend Monetr with support for accrual accounting.

## Architectural Overview
Accrual accounting allows users to track when they *consume* assets rather than just when they spend cash. For bulk items (e.g. rice, toiletries) or prepaid services (e.g. annual subscriptions), the expense is recognized daily/weekly/monthly over a defined consumption timeframe.

To support both **Linear Depreciation** and **Varying Manual Usage**, we use a two-table data model:
1. `accrual_expenses`: Stores the metadata, total value, and date range (Start to End).
2. `accrual_usage_logs`: Stores manual logs of usage for specific days (e.g. "Used $15 of toiletries on June 10").

### Amortization Algorithm
For any day $d$ in the interval $[\text{StartDate}, \text{EndDate}]$:
1. If there is a manual log in `accrual_usage_logs` on day $d$, the recognized expense is the logged amount.
2. If there is no manual log, the remaining unallocated balance of the accrual expense is distributed evenly across all remaining non-logged days in the interval.
3. If manual logs deplete the total amount before the end date, subsequent non-logged days default to $0$ recognized expense.

---

## Phases of Implementation

Please follow these phases in order, as each builds directly on the previous one. Each document is self-contained with full specification, directory contexts, code layouts, and tests:

1. [Phase 1: DB Schema & Migrations](file:///Users/janeetbajracharya/Desktop/Code/monetr/accrual_plan/phase_1_db.md)
2. [Phase 2: Go Backend Models & Repository Layer](file:///Users/janeetbajracharya/Desktop/Code/monetr/accrual_plan/phase_2_repo.md)
3. [Phase 3: Go Backend REST Controllers & Routes](file:///Users/janeetbajracharya/Desktop/Code/monetr/accrual_plan/phase_3_controllers.md)
4. [Phase 4: Frontend API Models & React Hooks](file:///Users/janeetbajracharya/Desktop/Code/monetr/accrual_plan/phase_4_frontend_api.md)
5. [Phase 5: Interactive Gantt-style Calendar UI](file:///Users/janeetbajracharya/Desktop/Code/monetr/accrual_plan/phase_5_calendar_ui.md)
6. [Phase 6: Drag-and-Drop & Resizing Interactions](file:///Users/janeetbajracharya/Desktop/Code/monetr/accrual_plan/phase_6_drag_resize.md)
7. [Phase 7: Transaction Sidebar Panel & Drag-to-Create](file:///Users/janeetbajracharya/Desktop/Code/monetr/accrual_plan/phase_7_sidebar_drag.md)
8. [Phase 8: Usage Log Entry Modal & Amortization Analytics](file:///Users/janeetbajracharya/Desktop/Code/monetr/accrual_plan/phase_8_usage_analytics.md)

Note for Agents:

After implementing each phase, mention if that phase is done. Make sure a succinct recap needed for later phases are mentioned and inserted such that the next agent can understand the whole context of each previous phase. 