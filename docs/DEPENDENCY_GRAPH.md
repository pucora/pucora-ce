# Enable GitHub Dependency Graph (for dependency-review CI)

The [Dependency Review](https://docs.github.com/en/code-security/supply-chain-security/understanding-your-software-supply-chain/about-dependency-review) workflow needs the dependency graph enabled on this repository.

## One-time setup (org admin)

1. Open **Settings → Code security and analysis** for [velonetics-ce](https://github.com/velonetics/velonetics-ce/settings/security_analysis).
2. Enable **Dependency graph**.
3. (Optional) Enable **Dependabot alerts** and **Dependabot security updates**.

After the graph is populated (first push to `main` with `go.mod`), PRs will get dependency-review results. Until then, the workflow is marked `continue-on-error: true` so it does not block merges.

## Optional: submit snapshots from CI

If the org allows `dependency-graph: write`, add a job using [dependency-submission-action](https://github.com/marketplace/actions/go-dependency-submission) to `.github/workflows/go.yml` to populate the graph automatically on each push to `main`.
