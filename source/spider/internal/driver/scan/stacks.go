package scan

const (
	stackGo     = "go"
	stackBun    = "bun"
	stackNode   = "node"
	stackPython = "python"
)

type marked struct {
	name    string
	manager string
}

var markers = []struct {
	file  string
	stack marked
}{
	{"go.mod", marked{stackGo, ""}},
	{"bun.lock", marked{stackBun, ""}},
	{"bun.lockb", marked{stackBun, ""}},
	{"bunfig.toml", marked{stackBun, ""}},
	{"pnpm-lock.yaml", marked{stackNode, "pnpm"}},
	{"yarn.lock", marked{stackNode, "yarn"}},
	{"package-lock.json", marked{stackNode, "npm"}},
	{"pyproject.toml", marked{stackPython, ""}},
	{"requirements.txt", marked{stackPython, ""}},
}

var runnerCommands = map[string][]string{
	stackGo:     {"go", "test", "{pkg}", "-run", "{target}"},
	stackBun:    {"bun", "test", "--cwd", "{pkg}", "{target}"},
	stackPython: {"python", "-m", "pytest", "{target}"},
	"pnpm": {
		"pnpm", "--filter", "{pkg}", "test", "--", "{target}",
	},
	"yarn": {"yarn", "workspace", "{pkg}", "test", "{target}"},
	"npm": {
		"npm", "test", "--workspace", "{pkg}", "--", "{target}",
	},
}

func detect(names map[string]bool, atRoot bool) []marked {
	var stacks []marked
	seen := map[string]bool{}

	for _, marker := range markers {
		if names[marker.file] && !seen[marker.stack.name] {
			seen[marker.stack.name] = true
			stacks = append(stacks, marker.stack)
		}
	}

	javascript := seen[stackBun] || seen[stackNode]
	if atRoot && names["package.json"] && !javascript {
		stacks = append(stacks, marked{stackBun, ""})
	}

	return stacks
}

func runnersFor(stacks []string, managers map[string]string) []Runner {
	runners := []Runner{}

	for _, stack := range stacks {
		key := stack
		if stack == stackNode {
			key = managers[stack]
		}

		runners = append(runners, Runner{
			Name:  "test-" + stack,
			Stack: stack,
			Run:   append([]string{}, runnerCommands[key]...),
		})
	}

	return runners
}
