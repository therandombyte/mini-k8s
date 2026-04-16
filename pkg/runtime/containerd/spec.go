package containerd

import "strings"

func buildContainerArgs(command, args []string) []string {
	if len(command) == 0 && len(args) == 0 {
		return nil
	}

	if len(command) > 0 {
		out := append([]string{}, command...)
		out = append(out, args...)
		return out
	}

	out := append([]string{}, args...)
	return out
}

func sanitizeID(in string) string {
	replacer := strings.NewReplacer("/", "-", "_", "-", " ", "-")
	return replacer.Replace(in)
}
