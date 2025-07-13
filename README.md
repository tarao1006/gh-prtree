# gh-prtree

A GitHub CLI extension that generates a tree structure from Pull Requests in a GitHub repository.

## Installation

```bash
gh extension install tarao1006/gh-prtree
```

## Usage

### Basic Usage

```bash
# Generate tree for current repository
gh prtree

# Generate tree for specific repository
gh prtree --repository owner/repo-name
```

### Options

| Option             | Short | Description                                   | Default                        |
| ------------------ | ----- | --------------------------------------------- | ------------------------------ |
| `--repository`     | `-r`  | Repository owner/name                         | Current directory's repository |
| `--author`         | `-a`  | Filter by author (can be used multiple times) | All authors                    |
| `--exclude-drafts` | `-d`  | Exclude draft pull requests                   | `true`                         |
| `--format`         | `-f`  | Output format (mermaid\|json\|graphviz)       | `mermaid`                      |

### Examples

#### Filter by Author

```bash
# Single author
gh prtree --author username

# Multiple authors
gh prtree --author alice --author bob
```

#### Different Output Formats

```bash
# Mermaid format (default)
gh prtree --format mermaid

# JSON format
gh prtree --format json

# GraphViz format
gh prtree --format graphviz
```

#### Include Draft PRs

```bash
gh prtree --exclude-drafts=false
```

## Output Formats

### Mermaid

Generates a [Mermaid](https://mermaid.js.org/) diagram showing the tree structure of Pull Requests.

```
graph TD
    main["main"]:::prNode
    PR123["Feature: Add new component"]:::prNode
    PR456["Fix: Update dependencies"]:::prNode

    PR123 --> main
    PR456 --> main

    click PR123 "https://github.com/owner/repo/pull/123"
    click PR456 "https://github.com/owner/repo/pull/456"
```

### JSON

Provides structured data with nodes and edges for programmatic processing.

```json
{
  "nodes": {
    "main": { "name": "main" },
    "feature-branch": { "name": "feature-branch" }
  },
  "edges": [
    {
      "from": "feature-branch",
      "to": "main",
      "pr": {
        "id": "PR_123",
        "title": "Feature: Add new component",
        "url": "https://github.com/owner/repo/pull/123"
      }
    }
  ],
  "defaultBranch": "main"
}
```

### GraphViz

Generates DOT notation for [GraphViz](https://graphviz.org/).

```
digraph PRGraph {
    rankdir=TB;
    node [shape=box, style=filled];
    edge [fontsize=10];

    "main" [fillcolor=lightblue, fontweight=bold];
    "feature-branch" [fillcolor=lightgray];

    "feature-branch" -> "main" [label="PR: Feature: Add new component\nAuthor: username"];
}
```

## License

MIT License
