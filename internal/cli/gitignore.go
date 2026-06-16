package cli

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// requiredGitignore are the entries Harpo guarantees in a project's .gitignore
// so that local secrets and the .harpo working dir are never committed.
// See MVP spec §11.3.
var requiredGitignore = []string{
	".harpo/",
	".env",
	".env.*",
	"!.env.example",
}

// ensureGitignore makes sure every required entry exists in <root>/.gitignore,
// appending any that are missing. Returns the entries it added.
func ensureGitignore(root string) ([]string, error) {
	path := filepath.Join(root, ".gitignore")

	existing := map[string]bool{}
	if f, err := os.Open(path); err == nil {
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			existing[strings.TrimSpace(sc.Text())] = true
		}
		f.Close()
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	var missing []string
	for _, entry := range requiredGitignore {
		if !existing[entry] {
			missing = append(missing, entry)
		}
	}
	if len(missing) == 0 {
		return nil, nil
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if len(existing) > 0 {
		if _, err := f.WriteString("\n"); err != nil {
			return nil, err
		}
	}
	if _, err := f.WriteString("# Harpo — local secrets, never commit\n"); err != nil {
		return nil, err
	}
	for _, entry := range missing {
		if _, err := f.WriteString(entry + "\n"); err != nil {
			return nil, err
		}
	}
	return missing, nil
}
