package main

import (
	"context"
	"flag"
	"fmt"
	ignore "github.com/sabhiram/go-gitignore"
	//"google.golang.org/api/option"
	"google.golang.org/genai"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var pathFlag string
var modelFlag string

func init() {
	flag.StringVar(&pathFlag, "path", "./", "Path to your application code.")
	flag.StringVar(&modelFlag, "model", "gemini-3-flash-preview", "Choose a Google Gemini AI model.")
	flag.Parse()
}

func main() {
	fmt.Println("Scanning...")
	fmt.Println(" - System instructions")
	systemInstruction, err := os.ReadFile("docs/templates/system_instruction.md")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(" - Output template")
	outputTemplate, err := os.ReadFile("docs/templates/output_template.md")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(" - Source code")
	codeContext, err := scanFiles(pathFlag)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(" - Architectural Decision Records (ADRs)")
	adrPath := filepath.Join(pathFlag, "docs", "adr")
	adrContext := collectDesignDecisions(adrPath)

	prompt := buildPrompt(systemInstruction, outputTemplate, adrContext, codeContext)

	fmt.Printf("Initializing genai client and generating content with: %s\n", modelFlag)
	response := initGenai(prompt)

	// Save the result
	fmt.Println("Saving genai response...")
	outputFile, err := saveFile(pathFlag, response)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf(" - Success! Documentation saved to: %s\n", outputFile)
	}
}

// Builds a prompt with strings.Builder that returns a prompt to be used with genai.
func buildPrompt(systemInstruction []byte, outputTemplate []byte, adrContext string, codeContext string) (prompt string) {
	var sb strings.Builder

	// A. System Instruction (The role)
	sb.Write(systemInstruction)
	sb.WriteString("\n---\n")

	// B. Template (The format)
	sb.Write(outputTemplate)
	sb.WriteString("\n---\n")

	// C. Design Decisions (The context)
	sb.WriteString("Here are the Architecture Decision Records (ADR):\n")
	sb.WriteString(adrContext)
	sb.WriteString("\n---\n")

	// D. Source Code
	sb.WriteString("Here is the source code:\n")
	sb.WriteString(codeContext)

	return sb.String()
}

func saveFile(pathFlag string, response string) (string, error) {
	outputDir := filepath.Join(pathFlag, "docs")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal(err)
	}

	outputFile := filepath.Join(outputDir, "AI_GENERATED.md")

	err := os.WriteFile(outputFile, []byte(response), 0644)

	return outputFile, err
}

// Initialize the genai client, and sends a prompt.
// The client gets the API key from the environment variable `GEMINI_API_KEY`.
func initGenai(prompt string) (genaiResponse string) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	response, err := client.Models.GenerateContent(
		ctx,
		modelFlag,
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	return response.Text()
}

// Recursive file scanner with .gitignore support, it recursively walks the directory tree, respecting .gitignore and whitelists.
func scanFiles(rootPath string) (string, error) {
	var sb strings.Builder

	// Attempt to load .gitignore rules
	gitIgnorePath := filepath.Join(rootPath, ".gitignore")
	gitIgnore, _ := ignore.CompileIgnoreFile(gitIgnorePath)

	// Hardcoded directories to always skip
	ignoredDirs := map[string]bool{
		".git":         true,
		"docs":         true, // Avoid recursive loop by ignoring output folder
		"node_modules": true,
		"vendor":       true,
	}

	// Whitelist allowed file extensions
	allowedExts := map[string]bool{
		".go": true, ".tf": true, ".yaml": true,
		".py": true, ".md": true, ".ts": true, ".js": true,
	}

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip files we cannot access
		}

		// Calculate relative path for cleaner logs and .gitignore matching
		relPath, _ := filepath.Rel(rootPath, path)

		// 1. Check hardcoded directory exclusions
		if d.IsDir() && ignoredDirs[d.Name()] {
			return filepath.SkipDir
		}

		// 2. Check .gitignore rules
		if gitIgnore != nil && gitIgnore.MatchesPath(relPath) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip if it is a directory (we only read files)
		if d.IsDir() {
			return nil
		}

		// 3. Check file extension whitelist
		ext := filepath.Ext(path)
		if allowedExts[ext] {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			sb.WriteString("\n--- FILE: " + relPath + " ---\n")
			sb.Write(content)
			sb.WriteString("\n")
		}

		return nil
	})

	return sb.String(), err
}

// Scans the
func collectDesignDecisions(adrPath string) string {
	// Check if folder exists
	if _, err := os.Stat(adrPath); os.IsNotExist(err) {
		return "No ADRs found."
	}

	// Read the folder
	entries, err := os.ReadDir(adrPath)
	if err != nil {
		log.Fatal(err)
	}

	var sb strings.Builder

	// Loop through the files
	for _, entry := range entries {
		if entry.IsDir() { // Skip subdirectories
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md") { // Keep only markdown
			continue
		}

		fullPath := filepath.Join(adrPath, entry.Name())
		content, _ := os.ReadFile(fullPath)

		// Build string
		sb.WriteString("\n--- DECISION RECORD: " + entry.Name() + " ---\n")
		sb.Write(content)
		sb.WriteString("\n")
	}

	return sb.String()
}
