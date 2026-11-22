package cmd

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yvv4git/macos-space/internal/application"
)

const bytesInUnit = 1024

func SetupAnalyseCommand() *cobra.Command {
	command := &cobra.Command{
		Use:          "analyse",
		Short:        "Show disk usage for configured directories",
		SilenceUsage: true,
		Example: `  # Run with default config
  macos-space analyse

  # Run with explicit config path
  macos-space analyse --config config.toml
  macos-space analyse -c config.toml

  # Scan a specific directory
  macos-space analyse --path /Users/vladimireliseev/Downloads
  macos-space analyse -p /Users/vladimireliseev/Downloads

  # Walk directories recursively
  macos-space analyse --recursive
  macos-space analyse -r

  # Limit recursive depth (0 means unlimited)
  macos-space analyse --depth 2
  macos-space analyse -d 2

  # Include files in output
  macos-space analyse --files
  macos-space analyse -f

  # Combine flags
  macos-space analyse --recursive --depth 2
  macos-space analyse -r -d 2
  macos-space analyse -p ~/Downloads -r -d 1
  macos-space analyse --recursive --files
  macos-space analyse -r -f
  macos-space analyse --path /Users/vladimireliseev/Downloads --recursive
  macos-space analyse -p /Users/vladimireliseev/Downloads -r
  macos-space analyse --path /Users/vladimireliseev/Downloads --recursive --depth 2
  macos-space analyse -p /Users/vladimireliseev/Downloads -r -d 2
  macos-space analyse --path /Users/vladimireliseev/Downloads --recursive --depth 1 --files
  macos-space analyse -p /Users/vladimireliseev/Downloads -r -d 1 -f
  macos-space analyse --config config.toml --path /Users/vladimireliseev/Downloads --recursive --depth 2 --files
  macos-space analyse -c config.toml -p /Users/vladimireliseev/Downloads -r -d 2 -f`,
		RunE: runAnalyse,
	}

	command.Flags().BoolP("recursive", "r", false, "walk directories recursively")
	command.Flags().IntP("depth", "d", 0, "limit recursive depth (0 means unlimited)")
	command.Flags().StringP("path", "p", "", "scan a specific directory instead of configured directories")
	command.Flags().BoolP("files", "f", false, "include files in output")

	return command
}

func runAnalyse(command *cobra.Command, _ []string) error {
	configPath, err := command.Flags().GetString("config")
	if err != nil {
		return fmt.Errorf("get config flag: %w", err)
	}

	recursive, err := command.Flags().GetBool("recursive")
	if err != nil {
		return fmt.Errorf("get recursive flag: %w", err)
	}

	depth, err := command.Flags().GetInt("depth")
	if err != nil {
		return fmt.Errorf("get depth flag: %w", err)
	}

	path, err := command.Flags().GetString("path")
	if err != nil {
		return fmt.Errorf("get path flag: %w", err)
	}

	includeFiles, err := command.Flags().GetBool("files")
	if err != nil {
		return fmt.Errorf("get files flag: %w", err)
	}

	app, err := application.NewAnalyseApplication(command.Context(), application.AnalyseOptions{
		ConfigPath: configPath,
	})
	if err != nil {
		return err
	}
	defer app.Shutdown()

	info := app.RunInfo(recursive, depth, path, includeFiles)
	printAnalyseInfo(command, info)

	result, err := app.Run(recursive, depth, path, includeFiles)
	if err != nil {
		return err
	}

	printAnalyseResult(result)

	return nil
}

type analyseNode struct {
	usage    application.DirectoryUsage
	children []*analyseNode
}

func printAnalyseInfo(command *cobra.Command, info *application.AnalyseRunInfo) {
	if info == nil {
		return
	}

	command.PrintErrln("Settings:")
	command.PrintErrf("  directories: %s\n", strings.Join(info.Directories, ", "))
	command.PrintErrf("  recursive: %t\n", info.Recursive)
	command.PrintErrf("  depth: %d\n", info.Depth)
	command.PrintErrf("  timeout: %s\n", info.DirectoryTimeout)
	command.PrintErrf("  files: %t\n", info.IncludeFiles)
	command.PrintErrln()
}

func printAnalyseResult(result *application.AnalyseResult) {
	roots := buildAnalyseTree(result.Directories)
	sizeWidth := analyseTreeSizeWidth(roots)
	for index, root := range roots {
		printAnalyseNode(root, "", index == len(roots)-1, true, sizeWidth)
	}
}

func buildAnalyseTree(usages []application.DirectoryUsage) []*analyseNode {
	nodes := make(map[string]*analyseNode, len(usages))
	sortedUsages := append([]application.DirectoryUsage(nil), usages...)
	sort.Slice(sortedUsages, func(i, j int) bool {
		return len(sortedUsages[i].Path) < len(sortedUsages[j].Path)
	})

	for _, usage := range sortedUsages {
		nodes[usage.Path] = &analyseNode{usage: usage}
	}

	roots := make([]*analyseNode, 0)
	for _, usage := range sortedUsages {
		node := nodes[usage.Path]
		parentPath := parentAnalysePath(usage.Path, nodes)
		if parentPath == "" {
			roots = append(roots, node)
			continue
		}

		parent := nodes[parentPath]
		parent.children = append(parent.children, node)
	}

	sortAnalyseNodes(roots)
	return roots
}

func parentAnalysePath(path string, nodes map[string]*analyseNode) string {
	current := filepath.Dir(path)
	for current != "." && current != string(filepath.Separator) {
		if _, ok := nodes[current]; ok {
			return current
		}

		next := filepath.Dir(current)
		if next == current {
			break
		}
		current = next
	}

	if _, ok := nodes[current]; ok {
		return current
	}

	return ""
}

func sortAnalyseNodes(nodes []*analyseNode) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].usage.SizeBytes == nodes[j].usage.SizeBytes {
			return nodes[i].usage.Path < nodes[j].usage.Path
		}

		return nodes[i].usage.SizeBytes > nodes[j].usage.SizeBytes
	})

	for _, node := range nodes {
		sortAnalyseNodes(node.children)
	}
}

func analyseTreeSizeWidth(nodes []*analyseNode) int {
	width := 0
	var walk func([]*analyseNode)
	walk = func(children []*analyseNode) {
		for _, node := range children {
			formattedSize := formatBytes(node.usage.SizeBytes)
			if len(formattedSize) > width {
				width = len(formattedSize)
			}
			walk(node.children)
		}
	}
	walk(nodes)

	return width
}

func printAnalyseNode(node *analyseNode, prefix string, isLast bool, isRoot bool, sizeWidth int) {
	label := node.usage.Path
	connector := ""
	if !isRoot {
		label = filepath.Base(node.usage.Path)
		connector = "├── "
		if isLast {
			connector = "└── "
		}
	}

	suffix := ""
	if node.usage.IsDir {
		suffix = "/"
	}

	fmt.Printf("%*s  %s%s%s%s\n", sizeWidth, formatBytes(node.usage.SizeBytes), prefix, connector, label, suffix)

	childPrefix := prefix
	if !isRoot {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
	}

	for index, child := range node.children {
		printAnalyseNode(child, childPrefix, index == len(node.children)-1, false, sizeWidth)
	}
}

func formatBytes(sizeBytes int64) string {
	if sizeBytes < bytesInUnit {
		return fmt.Sprintf("%d B", sizeBytes)
	}

	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	value := float64(sizeBytes)
	unit := "B"

	for _, currentUnit := range units {
		value /= bytesInUnit
		unit = currentUnit
		if value < bytesInUnit {
			break
		}
	}

	return fmt.Sprintf("%.1f %s", value, unit)
}
