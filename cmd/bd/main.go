package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"sbeads/backend"
	"sbeads/internal/domain"
)

func getDBPath() string {
	if p := os.Getenv("SBEADS_DB"); p != "" {
		return p
	}
	return ".sbeads"
}

func main() {
	if err := execute(os.Args); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func execute(args []string) error {
	if len(args) < 2 {
		fmt.Println("sbeads - Minimal issue tracker")
		fmt.Println("Usage: sbeads <command> [options]")
		fmt.Println("Commands: create, ready, claim, close, show, dep")
		return errors.New("command required")
	}

	command := args[1]

	switch command {
	case "create":
		return HandleCreate(args[2:])
	case "ready":
		return HandleReady(args[2:])
	case "claim":
		return HandleClaim(args[2:])
	case "close":
		return HandleClose(args[2:])
	case "show":
		return HandleShow(args[2:])
	case "dep":
		return HandleDep(args[2:])
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func HandleCreate(args []string) error {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	var title, issueType, description string
	var priority, blocked int

	fs.StringVar(&title, "title", "", "Issue title")
	fs.StringVar(&issueType, "issue-type", "", "Issue type")
	fs.IntVar(&priority, "priority", 0, "Priority (0-10)")
	fs.StringVar(&description, "description", "", "Description")
	fs.IntVar(&blocked, "blocked", 0, "Is blocked (0 or 1)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if title == "" && len(fs.Args()) > 0 {
		title = fs.Arg(0)
	}
	if title == "" {
		return errors.New("--title is required")
	}

	dbPath := getDBPath()
	err := backend.CreateTables(dbPath)
	if err != nil {
		return fmt.Errorf("creating tables: %w", err)
	}

	storage, err := backend.NewStorage(dbPath)
	if err != nil {
		return fmt.Errorf("creating storage: %w", err)
	}
	defer storage.Close()

	service := domain.NewServiceDomain(storage)
	issue, err := service.Create(title, issueType, priority, description, blocked)
	if err != nil {
		return fmt.Errorf("creating issue: %w", err)
	}

	fmt.Printf("Created issue: %s (%s)\n", issue.ID, issue.Title)
	return nil
}

func HandleReady(args []string) error {
	dbPath := getDBPath()
	if err := backend.CreateTables(dbPath); err != nil {
		return fmt.Errorf("creating tables: %w", err)
	}
	storage, err := backend.NewStorage(dbPath)
	if err != nil {
		return fmt.Errorf("creating storage: %w", err)
	}
	defer storage.Close()

	service := domain.NewServiceDomain(storage)
	issues, err := service.Ready()
	if err != nil {
		return fmt.Errorf("listing ready issues: %w", err)
	}

	if len(issues) == 0 {
		fmt.Println("No ready issues")
		return nil
	}

	for _, issue := range issues {
		fmt.Printf("%s: %s\n", issue.ID, issue.Title)
	}
	return nil
}

func HandleClaim(args []string) error {
	fs := flag.NewFlagSet("claim", flag.ContinueOnError)
	var id, assignee, agent string
	fs.StringVar(&id, "issue-id", "", "Issue ID")
	fs.StringVar(&assignee, "assignee", "", "Assignee")
	fs.StringVar(&agent, "agent", "me", "Agent name")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if id == "" && len(fs.Args()) > 0 {
		id = fs.Arg(0)
	}
	if id == "" {
		return errors.New("--issue-id is required")
	}

	dbPath := getDBPath()
	if err := backend.CreateTables(dbPath); err != nil {
		return fmt.Errorf("creating tables: %w", err)
	}
	storage, err := backend.NewStorage(dbPath)
	if err != nil {
		return fmt.Errorf("creating storage: %w", err)
	}
	defer storage.Close()

	service := domain.NewServiceDomain(storage)
	issue, err := service.Claim(id, assignee, agent)
	if err != nil {
		return fmt.Errorf("claiming issue: %w", err)
	}

	fmt.Printf("Claimed issue: %s (%s) by %s\n", issue.ID, issue.Title, agent)
	return nil
}

func HandleClose(args []string) error {
	fs := flag.NewFlagSet("close", flag.ContinueOnError)
	var id string
	fs.StringVar(&id, "issue-id", "", "Issue ID")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if id == "" && len(fs.Args()) > 0 {
		id = fs.Arg(0)
	}
	if id == "" {
		return errors.New("--issue-id is required")
	}

	dbPath := getDBPath()
	if err := backend.CreateTables(dbPath); err != nil {
		return fmt.Errorf("creating tables: %w", err)
	}
	storage, err := backend.NewStorage(dbPath)
	if err != nil {
		return fmt.Errorf("creating storage: %w", err)
	}
	defer storage.Close()

	service := domain.NewServiceDomain(storage)
	issue, err := service.Close(id)
	if err != nil {
		return fmt.Errorf("closing issue: %w", err)
	}

	fmt.Printf("Closed issue: %s (%s)\n", issue.ID, issue.Title)
	return nil
}

func HandleShow(args []string) error {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	var id string
	fs.StringVar(&id, "issue-id", "", "Issue ID")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if id == "" && len(fs.Args()) > 0 {
		id = fs.Arg(0)
	}
	if id == "" {
		return errors.New("--issue-id is required")
	}

	dbPath := getDBPath()
	if err := backend.CreateTables(dbPath); err != nil {
		return fmt.Errorf("creating tables: %w", err)
	}
	storage, err := backend.NewStorage(dbPath)
	if err != nil {
		return fmt.Errorf("creating storage: %w", err)
	}
	defer storage.Close()

	issue, err := storage.GetIssue(id)
	if err != nil {
		return fmt.Errorf("issue not found: %s", id)
	}

	fmt.Printf("Issue: %s\n", issue.ID)
	fmt.Printf("Title: %s\n", issue.Title)
	fmt.Printf("Type: %s\n", issue.IssueType)
	fmt.Printf("Priority: %d\n", issue.Priority)
	fmt.Printf("Description: %s\n", issue.Description)
	fmt.Printf("Blocked: %v\n", issue.IsBlocked)
	fmt.Printf("Closed: %v\n", issue.IsClosed)
	fmt.Printf("Claimed: %v\n", issue.IsClaimed)
	return nil
}

func HandleDep(args []string) error {
	if len(args) < 1 {
		fmt.Println("Usage: sbeads dep <add|list> [options]")
		return errors.New("subcommand required")
	}
	subcommand := args[0]

	switch subcommand {
	case "add":
		return HandleDepAdd(args[1:])
	case "list":
		return HandleDepList(args[1:])
	default:
		return fmt.Errorf("unknown dep command: %s", subcommand)
	}
}

func HandleDepAdd(args []string) error {
	fs := flag.NewFlagSet("dep add", flag.ContinueOnError)
	var childID, parentID string
	fs.StringVar(&childID, "child", "", "Child issue ID")
	fs.StringVar(&parentID, "parent", "", "Parent issue ID")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if childID == "" && len(fs.Args()) >= 2 {
		childID = fs.Arg(0)
		parentID = fs.Arg(1)
	}
	if childID == "" || parentID == "" {
		return errors.New("--child and --parent are required")
	}

	dbPath := getDBPath()
	if err := backend.CreateTables(dbPath); err != nil {
		return fmt.Errorf("creating tables: %w", err)
	}
	storage, err := backend.NewStorage(dbPath)
	if err != nil {
		return fmt.Errorf("creating storage: %w", err)
	}
	defer storage.Close()

	service := domain.NewServiceDomain(storage)
	err = service.AddDependency(childID, parentID)
	if err != nil {
		return fmt.Errorf("adding dependency: %w", err)
	}

	fmt.Printf("Added dependency: %s -> %s\n", childID, parentID)
	return nil
}

func HandleDepList(args []string) error {
	fs := flag.NewFlagSet("dep list", flag.ContinueOnError)
	var childID string
	fs.StringVar(&childID, "issue-id", "", "Issue ID")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if childID == "" && len(fs.Args()) > 0 {
		childID = fs.Arg(0)
	}
	if childID == "" {
		return errors.New("--issue-id is required")
	}

	dbPath := getDBPath()
	if err := backend.CreateTables(dbPath); err != nil {
		return fmt.Errorf("creating tables: %w", err)
	}
	storage, err := backend.NewStorage(dbPath)
	if err != nil {
		return fmt.Errorf("creating storage: %w", err)
	}
	defer storage.Close()

	service := domain.NewServiceDomain(storage)
	deps, err := service.GetDependencies(childID)
	if err != nil {
		return fmt.Errorf("listing dependencies: %w", err)
	}

	if len(deps) == 0 {
		fmt.Println("No dependencies")
		return nil
	}

	for _, dep := range deps {
		fmt.Printf("  %s\n", dep)
	}
	return nil
}
