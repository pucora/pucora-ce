# Dependency graph (optional CI polish)

## What is it?

GitHub can scan your `go.mod` and build a list of every library Velonetics depends on. That list is called the **dependency graph**.

The **Dependency Review** workflow (runs on pull requests) compares that list against known security advisories. If a PR adds a vulnerable package, GitHub can warn you in the PR.

Right now that workflow is set to `continue-on-error: true`, which means: *if it cannot run, ignore the failure and still allow merges.* That is intentional — the graph is not set up yet, so the check would always fail.

## Do you need to do anything?

**No, not for WebSocket or releases to work.** This is optional supply-chain hygiene.

If you want PR dependency warnings later:

1. Go to [velonetics-ce → Settings → Code security and analysis](https://github.com/velonetics/velonetics-ce/settings/security_analysis).
2. Turn **Dependency graph** to **Enabled**.
3. Wait until GitHub has indexed `go.mod` (usually after the next push to `main`).
4. Open `.github/workflows/dependency_review.yml` and delete the line `continue-on-error: true` so failed reviews block merges.

> The velonetics org does not allow CI to submit the graph automatically (`dependency-graph: write` is blocked). An admin must flip the setting in the GitHub UI.

## Docker Hub (niteesh20)

Images are published to **[niteesh20/velonetics](https://hub.docker.com/r/niteesh20/velonetics)**.

Add these secrets under [velonetics-ce → Settings → Secrets → Actions](https://github.com/velonetics/velonetics-ce/settings/secrets/actions):

| Secret | Value |
|--------|-------|
| `DOCKER_USERNAME` | `niteesh20` |
| `DOCKER_PASSWORD` | A Docker Hub access token (not your account password) |

Create a token at [Docker Hub → Account Settings → Security](https://hub.docker.com/settings/security).

On each GitHub release, the **Handle Release** workflow builds and pushes `niteesh20/velonetics:$TAG`.

Optional secrets for signed `.deb`/`.rpm` packages: `PGP_SIGNING_KEY`, `PGP_PASSPHRASE`.
