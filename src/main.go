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

func main() {

	// Initialize flags and flag values
	pathPtr := flag.String("path", "./", "Path to your application code.")
	modelPtr := flag.String("model", "gemini-3-flash-preview", "Choose a Google Gemini AI model.")
	flag.Parse()

	targetDirectory := *pathPtr
	modelName := *modelPtr

	//Read template files
	fmt.Println("Scanning...")
	fmt.Println("- Template files")
	systemInstruction, err := os.ReadFile("docs/templates/system_instruction.md")
	if err != nil {
		log.Fatal(err)
	}

	outputTemplate, err := os.ReadFile("docs/templates/output_template.md")
	if err != nil {
		log.Fatal(err)
	}

	// Read source code
	fmt.Println("- Source code")
	codeContext, err := scanFiles(targetDirectory)
	if err != nil {
		log.Fatal(err)
	}

	// Read Architectural Decision Records (ADRs)
	fmt.Println("- Architectural Decision Records (ADRs)")
	adrPath := filepath.Join(targetDirectory, "docs", "adr")
	adrContext := collectDesignDecisions(adrPath)

	// ---------------------------------------------------------
	// Build the prompt
	// ---------------------------------------------------------
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

	// D. Koden (Fakta)
	sb.WriteString("Here is the source code:\n")
	sb.WriteString(codeContext)

	fullPrompt := sb.String()

	// fmt.Println(fullPrompt)

	// ---------------------------------------------------------
	// 6. Send to AI
	// ---------------------------------------------------------
	// Initialize Gemini client
	// The client gets the API key from the environment variable `GEMINI_API_KEY`.
	fmt.Println("Initializing Gemini client")
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Generating content with: %s\n", modelName)
	response, err := client.Models.GenerateContent(
		ctx,
		modelName,
		genai.Text(fullPrompt),
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(response.Text())
	// ---------------------------------------------------------
	// 7. Save the result
	// ---------------------------------------------------------
	fmt.Println("Saving content response")
	outputDir := filepath.Join(targetDirectory, "docs")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal(err)
	}

	outputFile := filepath.Join(outputDir, "AI_GENERATED.md")

	fmt.Println("Creating file")
	err = os.WriteFile(outputFile, []byte(response.Text()), 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Success! Documentation saved to: %s\n", outputFile)
}

// Helper function: Recursive file scanner with .gitignore support
// scanFiles recursively walks the directory tree, respecting .gitignore and whitelists.
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

// Helper function: File scanning a specific directory
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
