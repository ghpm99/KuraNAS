$ErrorActionPreference = 'Stop'

$originalBranch = git branch --show-current
if ([string]::IsNullOrWhiteSpace($originalBranch)) {
	throw 'Could not determine current branch.'
}

git diff-index --quiet HEAD --
if ($LASTEXITCODE -ne 0) {
	throw 'Working tree has uncommitted changes.'
}

try {
	Write-Host ''
	Write-Host '======== Release Main (fast-forward) ========'

	git fetch origin
	git checkout main
	git pull --ff-only origin main
	git merge --ff-only origin/develop
	git push origin main
}
finally {
	git checkout $originalBranch | Out-Null
}

Write-Host 'Main was fast-forwarded to origin/develop.'
