---
name: core-structure
description: Derive names from architectural roles, decouple files and methods, and separate capability implementation from business orchestration
---

## Usage

Design the changed area as a small system before naming declarations or distributing code across files. Identify its domain, responsibility owners, dependency direction, and coordination point; reuse existing boundaries where they fit.

Apply [core-scope](core-scope.md) when responsibility boundaries require additional edits.

### Name by relationship and role

When declaring a variable, type, function, or file, derive its name from the domain object or collaborator it belongs to and its architectural role. Decide what it owns before choosing its name. A reader should be able to connect related declarations without tracing every assignment.

- Reuse the domain vocabulary already established by adjacent types and APIs.
- Add the owning concept when a generic name such as `config`, `data`, `manager`, or `utils` would be ambiguous.
- Keep names proportional to scope: a short local may be concise when its relationship is obvious; filenames and exported declarations need enough context to stand alone.
- Prefer names that state responsibility over names that state implementation mechanics.
- Distinguish concepts and lifecycle states when they affect behavior: `orderDraft`, `pricedOrder`, and `savedOrder` convey more than `data`, `result`, and `newData`. Avoid repeating context already clear from the enclosing scope.

### Give each file one reason to change

A file should own one cohesive responsibility, not merely a convenient collection of code. Split a file when its declarations change for different reasons; keep tightly coupled declarations together even if they have different syntactic kinds.

Do not split code into one-declaration files solely to make files smaller.

### Give each method one responsibility

Each function or method performs one domain operation or coordinates one cohesive workflow. Keep its steps at a consistent level of abstraction. Extract independently changing rules or side effects when a method mixes calculation, storage, rendering, or unrelated workflows.

Multiple steps do not imply multiple responsibilities: an order-placement workflow may sequence pricing and persistence without implementing either. Do not split every statement into a helper or invent a class for each operation.

### Make dependencies explicit

- Pass collaborators and required data through parameters or constructors. Avoid hidden global state and reaching through another component into its internals.
- Keep domain rules independent of UI layout and concrete infrastructure. Supply storage or network behavior at the boundary using the project's existing conventions; add an interface only when it provides a useful boundary.
- Let sibling components communicate through explicit contracts or their coordinator. Do not introduce circular imports or mutual control between peers.
- Keep each piece of mutable state with one clear owner. Expose operations or results rather than allowing unrelated components to mutate its internals.

### Separate capability implementation from business orchestration

Distinguish two responsibilities within the changed area:

- **Capability implementation** owns how one focused operation works: calculation rules, validation, persistence, or interaction with an external system. Expose explicit inputs, outputs, and failure contracts without depending on a particular calling workflow.
- **Business orchestration** owns how operations combine to fulfill a use case: their order, business branches, data flow, and workflow-level failure handling. Call capabilities through their contracts rather than reproducing their algorithms or infrastructure details.

Orchestration contains business decisions about the workflow; it is not necessarily a trivial forwarding layer. Keep operation-specific decisions with the capability that implements them. For example, pricing owns discount calculation, while order placement decides when pricing and persistence run.

Keep dependencies directed from orchestration to capability contracts. A capability returns a result or failure to its caller; it should not call back into a specific workflow to decide what happens next.

Choose names from the use case or operation and follow project conventions. This distinction does not require particular suffixes, classes, directories, or exactly two physical layers. A small workflow can be a function; separate independently changing responsibilities without adding wrappers around every call.

Keep dependency construction and wiring at an explicit entry point. Wiring supplies collaborators; business orchestration uses them to execute a use case. Do not confuse object assembly with business sequencing or move operation details into either role.

## Key Points

- Review with concrete change scenarios: a discount formula changes pricing, storage changes persistence, and the order of checkout steps changes orchestration. Unrelated changes spreading across these boundaries signal coupling.
