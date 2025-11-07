# Release Instructions

## Setup

1. **Install goreleaser**
   ```bash
   brew install goreleaser
   ```

2. **Create GitHub Personal Access Token**
   - Go to GitHub Settings > Developer settings > Personal access tokens > Tokens (classic)
   - Generate new token with `repo` scope
   - Export it: `export GITHUB_TOKEN=your_token_here`

3. **Create Homebrew Tap Repository**
   ```bash
   # Create a new repository on GitHub named: homebrew-apimug
   # Clone it locally
   git clone https://github.com/doganarif/homebrew-apimug.git
   cd homebrew-apimug
   mkdir Formula
   git add .
   git commit -m "Initial commit"
   git push
   ```

## Release Process

1. **Tag a new version**
   ```bash
   git tag -a v0.1.0 -m "First release"
   git push origin v0.1.0
   ```

2. **Run goreleaser**
   ```bash
   goreleaser release --clean
   ```

   Or test without publishing:
   ```bash
   goreleaser release --snapshot --clean
   ```

3. **Verify the release**
   - Check GitHub Releases page
   - Check homebrew-apimug repository for updated Formula

## Install via Homebrew

Once published, users can install with:

```bash
brew tap doganarif/apimug
brew install apimug
```

## Notes

- goreleaser handles everything: building, archiving, checksums, changelog
- Formula is automatically created and pushed to homebrew-apimug
- Binaries for macOS, Linux, Windows (amd64 & arm64) are built
- GitHub Releases page is automatically created with artifacts
