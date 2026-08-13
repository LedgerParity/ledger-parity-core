# Contributing to ledger-parity-core

Thank you for your interest in contributing to `ledger-parity-core`!

---

## 1. Drips Wave Guidelines

This repository participates in the **Drips Wave Program (Stellar Wave)**.

### Principles
1. **Focus on Real Impact:** All contributions should improve core reconciliation accuracy, speed up ingestion, or add new discrepancy detection rules.
2. **Clear Context & Scope:** Write descriptive pull request summaries linking to issues or detailing exact problem contexts.
3. **No Superficial Patches:** Do not hide errors by swallowing exceptions or altering test assertions without addressing root causes.

---

## 2. Development Setup

### Requirements
- Go 1.22+

### Workflow
1. Fork and clone the repository.
2. Create a feature branch:
   ```bash
   git checkout -b feature/my-new-feature
   ```
3. Run tests locally before opening a PR:
   ```bash
   go test -v ./...
   go vet ./...
   ```
4. Submit your Pull Request with a clear summary of changes and test verification results.
