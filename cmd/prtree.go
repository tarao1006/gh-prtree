package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/shurcooL/githubv4"
)

type Node struct {
	Name string `json:"name"`
}

type Graph struct {
	Nodes         map[string]*Node `json:"nodes"`
	Edges         []Edge           `json:"edges"`
	DefaultBranch string           `json:"defaultBranch"`
}

type Edge struct {
	From string      `json:"from"`
	To   string      `json:"to"`
	PR   PullRequest `json:"pr"`
}

func buildGraph(repository Repository, pullRequests []PullRequest) Graph {
	graph := Graph{
		Nodes:         make(map[string]*Node),
		Edges:         []Edge{},
		DefaultBranch: repository.DefaultBranchRef.Name,
	}

	graph.Nodes[repository.DefaultBranchRef.Name] = &Node{
		Name: repository.DefaultBranchRef.Name,
	}

	for _, pr := range pullRequests {
		if graph.Nodes[pr.BaseRefName] == nil {
			graph.Nodes[pr.BaseRefName] = &Node{
				Name: pr.BaseRefName,
			}
		}
		if graph.Nodes[pr.HeadRefName] == nil {
			graph.Nodes[pr.HeadRefName] = &Node{
				Name: pr.HeadRefName,
			}
		}
		graph.Edges = append(graph.Edges, Edge{
			From: pr.HeadRefName,
			To:   pr.BaseRefName,
			PR:   pr,
		})
	}

	return graph
}

var (
	branchNodeColor       = "#e1f5fe"
	branchNodeStrokeColor = "#0277bd"

	prNodeColor       = "#ffffff"
	prNodeStrokeColor = "#000000"

	approvedNodeColor       = "#e8f5e8"
	approvedNodeStrokeColor = "#2e7d32"
)

func generateMermaid(graph Graph) string {
	var builder strings.Builder

	builder.WriteString("graph TD\n")

	builder.WriteString(fmt.Sprintf("    %s[\"%s\"]:::branchNode\n", graph.DefaultBranch, graph.DefaultBranch))
	for _, edge := range graph.Edges {
		class := "prNode"
		if edge.PR.IsApproved {
			class = "approvedNode"
		}
		builder.WriteString(fmt.Sprintf("    %s[\"%s\"]:::%s\n", edge.PR.ID, edge.PR.Title, class))
	}
	builder.WriteString("\n")

	for _, edge := range graph.Edges {
		if edge.To == graph.DefaultBranch {
			builder.WriteString(fmt.Sprintf("    %s --> %s\n", edge.PR.ID, graph.DefaultBranch))
		}
	}
	for _, edge1 := range graph.Edges {
		for _, edge2 := range graph.Edges {
			if edge1.PR.ID == edge2.PR.ID {
				continue
			}
			if edge1.From == edge2.To && edge1.To == edge2.From {
				continue
			}

			if edge2.To == edge1.From {
				builder.WriteString(fmt.Sprintf("    %s --> %s\n", edge2.PR.ID, edge1.PR.ID))
			}
		}
	}
	builder.WriteString("\n")

	for _, edge := range graph.Edges {
		builder.WriteString(fmt.Sprintf("    click %s \"%s\"\n", edge.PR.ID, edge.PR.URL))
	}

	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("    classDef branchNode fill:%s,stroke:%s,stroke-width:2px\n", branchNodeColor, branchNodeStrokeColor))
	builder.WriteString(fmt.Sprintf("    classDef prNode fill:%s,stroke:%s,stroke-width:2px\n", prNodeColor, prNodeStrokeColor))
	builder.WriteString(fmt.Sprintf("    classDef approvedNode fill:%s,stroke:%s,stroke-width:2px\n", approvedNodeColor, approvedNodeStrokeColor))

	return builder.String()
}

func generateJSON(graph Graph) string {
	out, _ := json.Marshal(graph)
	return string(out)
}

func generateGraphViz(graph Graph) string {
	var builder strings.Builder

	builder.WriteString("digraph PRGraph {\n")
	builder.WriteString("    rankdir=TB;\n")
	builder.WriteString("    node [shape=box, style=filled];\n")
	builder.WriteString("\n")

	builder.WriteString(fmt.Sprintf("    \"%s\" [fillcolor=\"%s\", color=\"%s\", fontweight=bold];\n", graph.DefaultBranch, branchNodeColor, branchNodeStrokeColor))

	for _, edge := range graph.Edges {
		title := edge.PR.Title
		title = strings.ReplaceAll(title, "\"", "\\\"")
		title = strings.ReplaceAll(title, "\n", "\\n")

		color := prNodeStrokeColor
		fillColor := prNodeColor
		if edge.PR.IsApproved {
			color = approvedNodeStrokeColor
			fillColor = approvedNodeColor
		}

		builder.WriteString(fmt.Sprintf("    \"%s\" [fillcolor=\"%s\", color=\"%s\", label=\"%s\", URL=\"%s\"];\n",
			edge.PR.ID, fillColor, color, title, edge.PR.URL))
	}
	builder.WriteString("\n")

	for _, edge := range graph.Edges {
		if edge.To == graph.DefaultBranch {
			builder.WriteString(fmt.Sprintf("    \"%s\" -> \"%s\";\n", edge.PR.ID, graph.DefaultBranch))
		}
	}

	for _, edge1 := range graph.Edges {
		for _, edge2 := range graph.Edges {
			if edge1.PR.ID == edge2.PR.ID {
				continue
			}
			if edge1.From == edge2.To && edge1.To == edge2.From {
				continue
			}

			if edge2.To == edge1.From {
				builder.WriteString(fmt.Sprintf("    \"%s\" -> \"%s\";\n", edge2.PR.ID, edge1.PR.ID))
			}
		}
	}

	builder.WriteString("}\n")

	return builder.String()
}

type Format string

func (f *Format) Set(value string) error {
	switch value {
	case "json":
		*f = FormatJSON
	case "mermaid":
		*f = FormatMermaid
	case "graphviz":
		*f = FormatGraphViz
	}
	return nil
}

func (f *Format) String() string {
	return string(*f)
}

func (f *Format) Type() string {
	return "Format"
}

const (
	FormatJSON     Format = "json"
	FormatMermaid  Format = "mermaid"
	FormatGraphViz Format = "graphviz"
)

func printGraph(graph Graph, format Format) {
	var fn func(Graph) string
	switch format {
	case FormatJSON:
		fn = generateJSON
	case FormatMermaid:
		fn = generateMermaid
	case FormatGraphViz:
		fn = generateGraphViz
	default:
		fn = generateMermaid
	}
	fmt.Print(fn(graph))
}

func ExecutePRTree(options Options) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		log.Fatal(err)
	}

	if options.Format == "" {
		options.Format = FormatMermaid
	}

	repository := GetRepository(client, options.Repository)
	pullRequests := GetPullRequests(
		client,
		repository.Owner.Login,
		repository.Name,
		[]githubv4.PullRequestState{githubv4.PullRequestStateOpen},
		options.Authors,
		options.ExcludeDrafts,
		options.Format,
	)

	graph := buildGraph(repository, pullRequests)

	printGraph(graph, options.Format)
}
