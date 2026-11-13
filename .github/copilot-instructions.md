## graphify

For any question about this repo's architecture, structure, components, or how to add/modify/find
code, your first action should be `graphify query "<question>"` when `graphify-out/graph.json`
exists. Use `graphify path "<A>" "<B>"` for relationship questions and `graphify explain "<concept>"`
for focused-concept questions. These return a scoped subgraph, usually much smaller than the full
report or raw grep output.

Triggers: "how do I…", "where is…", "what does … do", "add/modify a <component>",
"explain the architecture", or anything that depends on how files or classes relate.

If `graphify-out/wiki/index.md` exists, use it for broad navigation. Read `graphify-out/GRAPH_REPORT.md`
only for broad architecture review or when query/path/explain do not surface enough context. Only read
source files when (a) modifying/debugging specific code, (b) the graph lacks the needed detail, or
(c) the graph is missing or stale.

Type `/graphify` in Copilot Chat to build or update the graph.

---

> This file defines mandatory rules for AI coding agents working in this repository.

---

## Design

UI work (HTML, CSS, and JavaScript under `internal/server/static/` and `internal/configeditor/web/`, plus any new UI) must follow the design system in `DESIGN.md`: colors, typography, spacing, radii, components, and its do's and don'ts. Read `DESIGN.md` before touching any UI code. If a needed token does not exist, add it to `DESIGN.md` first instead of introducing raw hex values or ad-hoc styles.

---

# General principles

* Correctness is more important than speed.
* Prefer the smallest change that completely solves the requested problem.
* Do not introduce unrelated refactoring.
* Do not rewrite working code unless it significantly improves correctness, maintainability or readability.
* Match the existing coding style of the repository instead of enforcing your own preferences.
* Prefer consistency over perfection.
* Do not introduce new abstractions, patterns or dependencies unless they solve a demonstrated problem.
* Never assume APIs, files, types or project structure exist. Search the repository first.

---

# Before writing code

Before modifying anything:

1. Understand the problem.
2. Identify the root cause.
3. Search for existing implementations.
4. Identify every affected file.
5. Reuse existing solutions whenever possible.
6. Explain your implementation plan before making non-trivial changes.

If requirements are ambiguous, ask instead of guessing.

If multiple reasonable implementations exist, briefly explain the tradeoffs and choose the simplest one unless instructed otherwise.

---

# Code quality

Write code that is easy to maintain.

Prefer:

* readability over cleverness
* explicitness over magic
* simple control flow
* deterministic behavior
* type safety
* small focused functions
* descriptive naming
* early returns instead of deep nesting

Avoid:

* duplicate logic
* unnecessary abstractions
* speculative optimizations
* dead code
* commented-out code

Do not fix unrelated issues while implementing the requested task.

If unrelated problems are discovered, mention them separately instead of modifying them.

---

# Comments

Comments exist to explain *why*, never *what*.

Only write documentation comments for exported/public symbols when required by the language convention (Go doc comments, JSDoc, XML documentation, etc.).

Do not write inline comments that merely explain code.

Inline comments are allowed only when documenting:

* protocol requirements
* RFC requirements
* platform-specific behavior
* compiler/toolchain quirks
* non-obvious business rules
* intentional deviations from otherwise obvious implementations

If code requires explanatory comments, prefer refactoring first.

All code comments must be written in English.

---

# Error handling

Fail loudly.

Never silently ignore errors.

Never suppress exceptions unless explicitly required.

Always preserve useful error information.

Return actionable error messages whenever appropriate.

Validate all external input.

Treat every external source as untrusted.

---

# Dependencies

Prefer existing project dependencies.

Do not introduce new libraries unless they provide clear value that cannot reasonably be achieved with existing dependencies.

If a new dependency is added:

* explain why
* update lockfiles
* verify compatibility with the project's runtime

---

# Tests

If existing tests cover the modified functionality:

* update them when necessary
* extend them when appropriate

Do not ignore failing tests.

If no automated tests exist, mention that the change could not be verified automatically.

---

# Documentation

For every non-trivial change:

Review existing documentation.

This includes, when applicable:

* README.md
* docs/
* CHANGELOG
* API documentation
* OpenAPI / Swagger
* examples/
* public documentation
* docstrings

If behavior, configuration or usage changes, update the relevant documentation within the same change.

If examples become outdated, update them.

If documentation is missing, do not create new documentation unless requested. Instead, mention that documentation should be added.

All documentation must be written in English.

---

# CI/CD verification

Before considering the task complete:

* inspect the project's CI configuration
* identify all relevant quality checks
* execute every check that can be run locally
* report the actual results

This includes, when applicable:

* formatting
* linting
* type checking
* unit tests
* integration tests
* end-to-end tests
* build verification
* security checks
* coverage thresholds

Also verify:

* runtime versions
* lockfiles
* workflow triggers
* required environment variables
* required secrets

If something cannot be verified locally, explicitly explain why instead of assuming success.

---

# Git

Never:

* add AI signatures
* add "Generated by..."
* add "Co-Authored-By"
* mention AI anywhere in commits, documentation or source code

Commit messages must:

* be concise
* describe the logical change
* follow the repository's existing convention
* contain only one logical change per commit

When using GitHub CLI or similar tools with multiline content, always write the content to a temporary file and use:

--body-file

instead of inline arguments.

---

# Completion checklist

Before declaring the task complete, verify:

* the requested functionality works
* affected tests pass (when possible)
* formatting passes
* linting passes
* type checking passes
* build passes (when applicable)
* documentation is updated
* examples are updated (if needed)
* exported APIs remain consistent
* no unrelated files were modified
* no TODO/FIXME comments were introduced
* no dead code remains

If any verification could not be performed, explicitly state why.

Never claim something has been tested unless you actually ran the corresponding command.

Never claim something works unless you have evidence.

When uncertain, say so.


