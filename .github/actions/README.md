# GitHub Actions pinning policy

External GitHub Actions used by workflows or local composite actions are pinned to
full commit SHAs. Keep the upstream version in a trailing comment so Dependabot,
Renovate, or a manual audit can identify the intended release.

When updating an action, resolve the release tag to its current commit and review
the upstream changelog before replacing the pinned SHA.
