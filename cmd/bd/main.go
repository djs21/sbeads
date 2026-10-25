package main

import (
	"flag"
	"fmt"
	"os"

	"sbeads/backend"
	"sbeads/internal/domain"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("sbeads - Minimal issue tracker")
		fmt.Println("Usage: sbeads <command> [options]")
		fmt.Println("Commands: create, ready, claim, close, show, dep")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "create":
		HandleCreate()
	case "ready":
		HandleReady()
	case "claim":
		HandleClaim()
	case "close":
		HandleClose()
	case "show":
		HandleShow()
	case "dep":
		HandleDep()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func HandleCreate() {
	var title, issueType, description string
	var priority int
	var blocked int

	flag.StringVar(&title, "title", "", "Issue title")
	flag.StringVar(&issueType, "issue-type", "", "Issue type")
	flag.IntVar(&priority, "priority", 0, "Priority (0-10)")
	flag.StringVar(&description, "description", "", "Description")
	flag.IntVar(&blocked, "blocked", 0, "Is blocked (0 or 1)")
	flag.CommandLine.Parse(os.Args[2:])

	if title == "" {
		fmt.Println("Error: --title is required")
		os.Exit(1)
	}

	storage, err := backend.NewStorage(".sbeads")
	if err != nil {
		fmt.Printf("Error creating storage: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()

	err = backend.CreateTables(".sbeads")
	if err != nil {
		fmt.Printf("Error creating tables: %v\n", err)
		os.Exit(1)
	}

	service := domain.NewServiceDomain(storage)
	issue, err := service.Create(title, issueType, priority, description, 0)
	if err != nil {
		fmt.Printf("Error creating issue: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created issue: %s (%s)\n", issue.ID, issue.Title)
}

func HandleReady() {
	storage, err := backend.NewStorage(".sbeads")
	if err != nil {
		fmt.Printf("Error creating storage: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()

	service := domain.NewServiceDomain(storage)
	issues, err := service.Ready()
	if err != nil {
		fmt.Printf("Error listing ready issues: %v\n", err)
		os.Exit(1)
	}

	if len(issues) == 0 {
		fmt.Println("No ready issues")
		return
	}

	for _, issue := range issues {
		fmt.Printf("%s: %s\n", issue.ID, issue.Title)
	}
}

func HandleClaim() {
	var id, assignee, agent string
	flag.StringVar(&id, "issue-id", "", "Issue ID")
	flag.StringVar(&assignee, "assignee", "", "Assignee")
	flag.StringVar(&agent, "agent", "me", "Agent name")
	flag.Parse()

	if id == "" {
		fmt.Println("Error: --issue-id is required")
		os.Exit(1)
	}

	storage, err := backend.NewStorage(".sbeads")
	if err != nil {
		fmt.Printf("Error creating storage: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()

	service := domain.NewServiceDomain(storage)
	issue, err := service.Claim(id, assignee, agent)
	if err != nil {
		fmt.Printf("Error claiming issue: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Claimed issue: %s (%s) by %s\n", issue.ID, issue.Title, agent)
}

func HandleClose() {
	var id string
	flag.StringVar(&id, "issue-id", "", "Issue ID")
	flag.Parse()

	if id == "" {
		fmt.Println("Error: --issue-id is required")
		os.Exit(1)
	}

	storage, err := backend.NewStorage(".sbeads")
	if err != nil {
		fmt.Printf("Error creating storage: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()

	service := domain.NewServiceDomain(storage)
	issue, err := service.Close(id)
	if err != nil {
		fmt.Printf("Error closing issue: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Closed issue: %s (%s)\n", issue.ID, issue.Title)
}

func HandleShow() {
	var id string
	flag.StringVar(&id, "issue-id", "", "Issue ID")
	flag.Parse()

	if id == "" {
		fmt.Println("Error: --issue-id is required")
		os.Exit(1)
	}

	storage, err := backend.NewStorage(".sbeads")
	if err != nil {
		fmt.Printf("Error creating storage: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()

	issue, err := storage.GetIssue(id)
	if err != nil {
		fmt.Printf("Error: Issue not found: %s\n", id)
		os.Exit(1)
	}

	fmt.Printf("Issue: %s\n", issue.ID)
	fmt.Printf("Title: %s\n", issue.Title)
	fmt.Printf("Type: %s\n", issue.IssueType)
	fmt.Printf("Priority: %d\n", issue.Priority)
	fmt.Printf("Description: %s\n", issue.Description)
	fmt.Printf("Blocked: %v\n", issue.IsBlocked)
	fmt.Printf("Closed: %v\n", issue.IsClosed)
	fmt.Printf("Claimed: %v\n", issue.IsClaimed)
}

func HandleDep() {
	subcommand := flag.Arg(1)

	switch subcommand {
	case "add":
		HandleDepAdd()
	case "list":
		HandleDepList()
	default:
		fmt.Printf("Unknown dep command: %s\n", subcommand)
		os.Exit(1)
	}
}

func HandleDepAdd() {
	var childID, parentID string
	flag.StringVar(&childID, "child", "", "Child issue ID")
	flag.StringVar(&parentID, "parent", "", "Parent issue ID")
	flag.Parse()

	if childID == "" || parentID == "" {
		fmt.Println("Error: --child and --parent are required")
		os.Exit(1)
	}

	storage, err := backend.NewStorage(".sbeads")
	if err != nil {
		fmt.Printf("Error creating storage: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()

	service := domain.NewServiceDomain(storage)
	err = service.AddDependency(childID, parentID)
	if err != nil {
		fmt.Printf("Error adding dependency: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Added dependency: %s -> %s\n", childID, parentID)
}

func HandleDepList() {
	var childID string
	flag.StringVar(&childID, "issue-id", "", "Issue ID")
	flag.Parse()

	if childID == "" {
		fmt.Println("Error: --issue-id is required")
		os.Exit(1)
	}

	storage, err := backend.NewStorage(".sbeads")
	if err != nil {
		fmt.Printf("Error creating storage: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()

	service := domain.NewServiceDomain(storage)
	deps, err := service.GetDependencies(childID)
	if err != nil {
		fmt.Printf("Error listing dependencies: %v\n", err)
		os.Exit(1)
	}

	if len(deps) == 0 {
		fmt.Println("No dependencies")
		return
	}

	for _, dep := range deps {
		fmt.Printf("  %s\n", dep)
	}
}
