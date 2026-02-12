package main

import (
	"context"
	"flag"
	"fmt"
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
	systemInstruction, err := os.ReadFile("docs/templates/system_instruction.md")
	if err != nil {
		log.Fatal(err)
	}

	outputTemplate, err := os.ReadFile("docs/templates/output_template.md")
	if err != nil {
		log.Fatal(err)
	}

	// Read source code
	codeContext, err := scanFiles(targetDirectory)
	if err != nil {
		log.Fatal(err)
	}

	// Read Architectural Decision Records (ADRs)
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
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

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
	outputDir := filepath.Join(targetDirectory, "docs")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal(err)
	}

	outputFile := filepath.Join(outputDir, "AI_GENERATED.md")

	err = os.WriteFile(outputFile, []byte(response.Text()), 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Success! Documentation saved to: %s\n", outputFile)
}

// Helper function: Recursive file scanner
func scanFiles(rootPath string) (string, error) {
	var sb strings.Builder

	// Blacklist directories
	ignoreMap := map[string]bool{
		".git": true, "node_modules": true, "docs": true,
	}

	// Whitelist file extensions
	allowedExtMap := map[string]bool{
		".go": true, ".tf": true, ".yaml": true, ".py": true, ".md": true,
	}

	// Walk through directories and filter them
	err := filepath.WalkDir(rootPath, func(path string, dir os.DirEntry, err error) error {

		if err != nil {
			return nil
		}

		// Check if it is a directory and filter out blacklist
		if dir.IsDir() {
			if ignoreMap[dir.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)

		// Check if file extension is in our whitelist
		if allowedExtMap[ext] {

			// Read files
			content, err := os.ReadFile(path)
			if err != nil {
				// Skip file if we're not able to read
				return nil
			}

			// Format the header so the AI knows what file it is and add it to built string
			// "--- FILE: src/main.go ---"
			sb.WriteString("\n--- FILE: " + path + " ---\n")

			// Add content from read files to built string
			sb.Write(content)

			// Add new line for cleaner look
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
		return "No ADRs found." // Helt ok, vi bare returnerer tomt.
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
