# Enable GitHub Dependency Graph (for dependency-review CI)

The [Dependency Review](https://docs.github.com/en/code-security/supply-chain-security/understanding-your-software-supply-chain/about-dependency-review) workflow needs the dependency graph enabled on this repository.

## One-time setup (org admin)

1. Open **Settings → Code security and analysis** for [velonetics-ce](https://github.com/velonetics/velonetics-ce/settings/security_analysis).
2. Enable **Dependency graph**.
3. (Optional) Enable **Dependabot alerts** and **Dependabot security updates**.

After the graph is populated, PRs will get dependency-review results. Until then, the workflow is marked `continue-on-error: true` so it does not block merges.

## Automatic snapshot submission

`.github/workflows/dependency-submission.yml` submits Go dependency snapshots on each push to `main` using [go-dependency-submission-action](https://github.com/marketplace/actions/go-dependency-submission). The job uses `continue-on-error: true` if the org blocks `dependency-graph: write`.

Once dependency review reports succeed on PRs, remove `continue-on-error: true` from `.github/workflows/dependency_review.yml`.

## Required secrets for release Docker publish

| Secret | Purpose |
|--------|---------|
| `DOCKER_USERNAME` | Docker Hub login (e.g. `niteesh20`) |
| `DOCKER_PASSWORD` | Docker Hub token |
| `PGP_SIGNING_KEY` | Optional — signed `.deb`/`.rpm` artifacts |
| `PGP_PASSPHRASE` | Optional — GPG passphrase |

Without Docker secrets, the Handle Release workflow still completes: builder/deb/rpm jobs are skipped and only the optional `ce-docker` job runs when credentials exist.
