# Gemini Docs

## Usage
### GitHub Actions
```yaml
name: Update Documentation

on:
  push:
    branches: [ main ]
    paths:
      - 'src/**'     
      - 'terraform/**'
      - '!docs/**'    

jobs:
  workflow-gemini-docs:
    uses: vetle-dev/workflow-gemini-docs/.github/workflows/main.yml@main
    with:
      model: 'gemini-1.5-pro'
    secrets:
      GEMINI_API_KEY: ${{ secrets.GEMINI_API_KEY }}
```

### Binary
```bash
go run src/main.go -h
  -model string
        Choose a Google Gemini AI model. (default "gemini-3-flash-preview")
  -path string
        Path to your application code. (default "./")
```