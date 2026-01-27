/*
 * Copyright 2021 American Express
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing
 * permissions and limitations under the License.
 */

package file

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/americanexpress/earlybird/v4/pkg/wildcard"
)

func Test_ExtendedGitIgnoreSamples(t *testing.T) {
	// Define test cases for each language/sample
	type testCase struct {
		filePath      string // Path relative to searchDir
		shouldIgnore  bool
	}

	samples := []struct {
		name       string
		ignoreFile string
		cases      []testCase
	}{
		{
			name:       "Go",
			ignoreFile: "test_data/gitignore_samples/Go.gitignore",
			cases: []testCase{
				// Ignored patterns
				{"myprogram.exe", true},
				{"test.exe~", true},
				{"pkg.dll", true},
				{"build/main.test", true},
				{"coverage.out", true},
				{"profile.cov", true},
				{"go.work", true},
				{".env", true},
				{"config/.env", true}, // .env can be anywhere (depends on matcher root, assuming root for now)
				
				// Not ignored
				{"main.go", false},
				{"go.mod", false},
				{"readme.md", false},
				{"pkg/main.go", false},
			},
		},
		{
			name:       "Python",
			ignoreFile: "test_data/gitignore_samples/Python.gitignore",
			cases: []testCase{
				// Ignored patterns
				{"__pycache__/cache.pyc", true},
				{"src/__pycache__/cache.pyc", true},
				{"module.so", true},
				{"build/lib/pkg", true},
				{"dist/package-1.0.tar.gz", true},
				{".env", true},
				{".venv/bin/activate", true},
				{".idea/workspace.xml", false}, // Python.gitignore doesn't ignore .idea by default (usually global)
				{"htmlcov/index.html", true},
				
				// Not ignored
				{"main.py", false},
				{"setup.py", false},
				{"requirements.txt", false},
				{"src/module.py", false},
			},
		},
		{
			name:       "Node",
			ignoreFile: "test_data/gitignore_samples/Node.gitignore",
			cases: []testCase{
				// Ignored patterns
				{"node_modules/package.json", true},
				{"logs/debug.log", true},
				{"npm-debug.log", true},
				{"coverage/lcov.info", true},
				{".env", true},
				{".env.local", true},
				{"dist/app.js", true},
				{".DS_Store", false}, // standard macOS ignore not in Node.gitignore usually (but good to check it's not falsely positive)
				
				// Negation checks (previously unsupported)
				{".env.example", false}, // Explicitly un-ignored: !.env.example
				
				// Not ignored
				{"package.json", false},
				{"src/index.js", false},
				{"public/index.html", false},
			},
		},
		{
			name:       "Java",
			ignoreFile: "test_data/gitignore_samples/Java.gitignore",
			cases: []testCase{
				// Ignored patterns
				{"Main.class", true},
				{"build/classes/Shape.class", true},
				{"app.jar", true},
				{"lib/dependency.war", true},
				{"server.log", true},
				{"hs_err_pid1234.log", true},
				
				// Not ignored
				{"Main.java", false},
				{"gradlew", false},
				{"pom.xml", false},
			},
		},
	}

	for _, sample := range samples {
		t.Run(sample.name, func(t *testing.T) {
			// Load ignore patterns from the sample file
			patterns, err := loadIgnorePatterns(sample.ignoreFile)
			if err != nil {
				t.Fatalf("Failed to load ignore patterns from %s: %v", sample.ignoreFile, err)
			}
			
			for _, tc := range sample.cases {
				// Test if the file path matches any ignore pattern
				// Using the same logic as isIgnoredFile but adapted for testing
				got := matchesAnyPattern(tc.filePath, patterns)
				if got != tc.shouldIgnore {
					t.Errorf("[%s] File '%s': expected ignore=%v, got=%v", sample.name, tc.filePath, tc.shouldIgnore, got)
				}
			}
		})
	}
}

// loadIgnorePatterns loads ignore patterns from a file, similar to getIgnorePatterns but simplified for testing
func loadIgnorePatterns(filePath string) ([]string, error) {
	var patterns []string
	
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}
	
	file, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Ignore comment lines (starting with #) and empty lines
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}
	
	return patterns, scanner.Err()
}

// matchesAnyPattern checks if a file path matches any of the ignore patterns
func matchesAnyPattern(filePath string, patterns []string) bool {
	// Add leading slash to match the behavior of isIgnoredFile
	if !strings.HasPrefix(filePath, "/") {
		filePath = "/" + filePath
	}
	
	for _, pattern := range patterns {
		if wildcard.PatternMatch(filePath, pattern) {
			return true
		}
	}
	return false
}