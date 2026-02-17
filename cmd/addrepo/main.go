package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"deployment-logs/internal/config"
	"deployment-logs/internal/database"
	"deployment-logs/internal/models"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	db, err := database.Setup(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	r := bufio.NewReader(os.Stdin)

	fmt.Println("📦 Add New Repository Config")
	fmt.Println("----------------------------")

	name := prompt(r, "Name", "")
	repoURL := prompt(r, "Repo URL", "")
	provider := prompt(r, "Provider (github/gitlab)", "github")
	defaultBranch := prompt(r, "Default branch", "main")
	branches := prompt(r, "Branches (comma-separated)", "develop,main")

	fmt.Print("SSH private key (paste, then empty line to finish; empty = none):\n")
	sshKey := readMultiline(r)

	var branchesJSON string
	if branches != "" {
		parts := strings.Split(branches, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		b, _ := json.Marshal(parts)
		branchesJSON = string(b)
	}

	repo := models.RepositoryConfig{
		Name:          name,
		RepoURL:       repoURL,
		Provider:      provider,
		SSHKey:        sshKey,
		DefaultBranch: defaultBranch,
		Branches:      branchesJSON,
	}

	if err := db.Create(&repo).Error; err != nil {
		log.Fatal("Failed to create config:", err)
	}

	fmt.Printf("\n✅ Created repo config #%d: %s\n", repo.ID, repo.Name)
}

func prompt(r *bufio.Reader, label, def string) string {
	if def != "" {
		fmt.Printf("%s [%s]: ", label, def)
	} else {
		fmt.Printf("%s: ", label)
	}
	line, _ := r.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	return line
}

func readMultiline(r *bufio.Reader) string {
	var lines []string
	for {
		line, _ := r.ReadString('\n')
		if strings.TrimSpace(line) == "" {
			break
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "")
}
