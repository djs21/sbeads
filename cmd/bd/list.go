package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"sbeads/internal/domain"
	"sbeads/internal/types"
)

type issueJSONItem struct {
	ID        string           `json:"id"`
	Title     string           `json:"title"`
	IssueType string           `json:"issue_type"`
	Priority  int              `json:"priority"`
	Status    string           `json:"status"`
	IsBlocked int              `json:"is_blocked"`
	IsClosed  int              `json:"is_closed"`
	IsClaimed int              `json:"is_claimed"`
	Agent     string           `json:"agent,omitempty"`
	Assignee  string           `json:"assignee,omitempty"`
	Children  []*issueJSONItem `json:"children,omitempty"`
}

// NewListCmd creates the list command with filtering and tree view support.
func NewListCmd() *cobra.Command {
	var (
		allFlag     bool
		closedFlag  bool
		blockedFlag bool
		treeFlag    bool
		jsonFlag    bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues with optional filtering and tree hierarchy",
		RunE: func(cmd *cobra.Command, args []string) error {
			appCtx := GetAppContext(cmd.Context())
			if appCtx == nil || appCtx.Storage == nil {
				return errors.New("storage unavailable in context")
			}

			// 1. Fetch all issues
			allIssues, err := appCtx.Storage.ListIssues()
			if err != nil {
				return fmt.Errorf("listing issues: %w", err)
			}

			if len(allIssues) == 0 {
				if jsonFlag {
					fmt.Fprintln(cmd.OutOrStdout(), "[]")
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "No issues found")
				}
				return nil
			}

			issueMap := make(map[string]*types.Issue, len(allIssues))
			for i := range allIssues {
				issueMap[allIssues[i].ID] = &allIssues[i]
			}

			// 2. Fetch dependencies per issue and compute status
			parentsMap := make(map[string][]string, len(allIssues))
			statusMap := make(map[string]domain.Status, len(allIssues))

			for _, iss := range allIssues {
				deps, err := appCtx.Storage.GetDependencies(iss.ID)
				if err != nil {
					return fmt.Errorf("getting dependencies for %s: %w", iss.ID, err)
				}
				parentsMap[iss.ID] = deps

				hasUnclosedDeps := false
				for _, pID := range deps {
					if p, ok := issueMap[pID]; ok && p.IsClosed == 0 {
						hasUnclosedDeps = true
						break
					}
				}
				statusMap[iss.ID] = domain.ComputeStatus(&iss, hasUnclosedDeps)
			}

			// 3. Filter issues
			filteredMap := make(map[string]*types.Issue)
			var filteredList []*types.Issue

			for i := range allIssues {
				iss := &allIssues[i]
				status := statusMap[iss.ID]
				match := false

				if allFlag {
					match = true
				} else if closedFlag || blockedFlag {
					if closedFlag && (status == domain.StatusClosed || iss.IsClosed != 0) {
						match = true
					}
					if blockedFlag && status == domain.StatusBlocked {
						match = true
					}
				} else {
					// default filter: only unclosed issues
					match = (iss.IsClosed == 0)
				}

				if match {
					filteredMap[iss.ID] = iss
					filteredList = append(filteredList, iss)
				}
			}

			if len(filteredList) == 0 {
				if jsonFlag {
					fmt.Fprintln(cmd.OutOrStdout(), "[]")
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "No issues found")
				}
				return nil
			}

			// Deterministic sort: priority DESC, created_at ASC, id ASC
			sortIssues := func(issues []*types.Issue) {
				sort.Slice(issues, func(i, j int) bool {
					if issues[i].Priority != issues[j].Priority {
						return issues[i].Priority > issues[j].Priority
					}
					if !issues[i].CreatedAt.Equal(issues[j].CreatedAt) {
						return issues[i].CreatedAt.Before(issues[j].CreatedAt)
					}
					return issues[i].ID < issues[j].ID
				})
			}

			// 4. Flat list output (when --tree is not passed)
			if !treeFlag {
				sortIssues(filteredList)

				if jsonFlag {
					var items []*issueJSONItem
					for _, iss := range filteredList {
						items = append(items, &issueJSONItem{
							ID:        iss.ID,
							Title:     iss.Title,
							IssueType: iss.IssueType,
							Priority:  iss.Priority,
							Status:    string(statusMap[iss.ID]),
							IsBlocked: iss.IsBlocked,
							IsClosed:  iss.IsClosed,
							IsClaimed: iss.IsClaimed,
							Agent:     iss.Agent,
							Assignee:  iss.Assignee,
						})
					}
					data, err := json.MarshalIndent(items, "", "  ")
					if err != nil {
						return fmt.Errorf("marshaling json: %w", err)
					}
					fmt.Fprintln(cmd.OutOrStdout(), string(data))
					return nil
				}

				for _, iss := range filteredList {
					fmt.Fprintln(cmd.OutOrStdout(), domain.FormatOneLiner(iss, statusMap[iss.ID]))
				}
				return nil
			}

			// 5. Hierarchical tree view (when --tree is passed)
			childrenMap := make(map[string][]*types.Issue)
			hasParentInTree := make(map[string]bool)

			for _, iss := range filteredList {
				for _, pID := range parentsMap[iss.ID] {
					p, ok := filteredMap[pID]
					if ok && strings.EqualFold(p.IssueType, "epic") {
						childrenMap[pID] = append(childrenMap[pID], iss)
						hasParentInTree[iss.ID] = true
						break
					}
				}
			}

			for pID := range childrenMap {
				sortIssues(childrenMap[pID])
			}

			// Root issues: issues without parent, or whose parent is not an epic/root
			var roots []*types.Issue
			for _, iss := range filteredList {
				if !hasParentInTree[iss.ID] {
					roots = append(roots, iss)
				}
			}
			sortIssues(roots)

			// Tree JSON representation
			if jsonFlag {
				visited := make(map[string]bool)
				var buildJSONNode func(iss *types.Issue) *issueJSONItem
				buildJSONNode = func(iss *types.Issue) *issueJSONItem {
					node := &issueJSONItem{
						ID:        iss.ID,
						Title:     iss.Title,
						IssueType: iss.IssueType,
						Priority:  iss.Priority,
						Status:    string(statusMap[iss.ID]),
						IsBlocked: iss.IsBlocked,
						IsClosed:  iss.IsClosed,
						IsClaimed: iss.IsClaimed,
						Agent:     iss.Agent,
						Assignee:  iss.Assignee,
					}
					visited[iss.ID] = true
					for _, child := range childrenMap[iss.ID] {
						if !visited[child.ID] {
							node.Children = append(node.Children, buildJSONNode(child))
						}
					}
					return node
				}

				var rootNodes []*issueJSONItem
				for _, root := range roots {
					rootNodes = append(rootNodes, buildJSONNode(root))
				}
				data, err := json.MarshalIndent(rootNodes, "", "  ")
				if err != nil {
					return fmt.Errorf("marshaling json: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			// Tree text rendering with branch connectors
			visited := make(map[string]bool)
			var renderNode func(node *types.Issue, prefix string, depth int)
			renderNode = func(node *types.Issue, prefix string, depth int) {
				visited[node.ID] = true
				children := childrenMap[node.ID]

				for i, child := range children {
					if visited[child.ID] || depth >= 10 {
						continue
					}
					isLast := (i == len(children)-1)
					branch := "├── "
					nextPrefix := prefix + "│   "
					if isLast {
						branch = "└── "
						nextPrefix = prefix + "    "
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%s%s%s\n", prefix, branch, domain.FormatOneLiner(child, statusMap[child.ID]))
					renderNode(child, nextPrefix, depth+1)
				}
			}

			for _, root := range roots {
				fmt.Fprintln(cmd.OutOrStdout(), domain.FormatOneLiner(root, statusMap[root.ID]))
				renderNode(root, "  ", 1)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Show all issues (including closed)")
	cmd.Flags().BoolVarP(&closedFlag, "closed", "c", false, "Show only closed issues")
	cmd.Flags().BoolVarP(&blockedFlag, "blocked", "b", false, "Show only blocked issues")
	cmd.Flags().BoolVarP(&treeFlag, "tree", "t", false, "Show hierarchical tree view")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output in JSON format")

	return cmd
}
