package system

import "fmt"

// DownloadAndInstallPackages downloads packages from URLs using curl and installs them in chroot.
// Supports mirror fallbacks: URLs with the same package name but different mirrors.
func DownloadAndInstallPackages(logPath, chrootPath string, urls ...string) error {
	if len(urls) == 0 {
		return fmt.Errorf("no URLs provided")
	}

	// Group URLs by package (same filename = same package, different mirrors)
	packageMirrors := make(map[string][]string)
	for _, url := range urls {
		var pkgName string
		for i := len(url) - 1; i >= 0; i-- {
			if url[i] == '/' {
				pkgName = url[i+1:]
				break
			}
		}
		packageMirrors[pkgName] = append(packageMirrors[pkgName], url)
	}

	// Build command with mirror fallback per package
	cmd := "mkdir -p /tmp && "
	for i, pkgName := range getSortedKeys(packageMirrors) {
		mirrors := packageMirrors[pkgName]
		cmd += fmt.Sprintf("(curl -L -f -o /tmp/pkg%d.tar.zst '%s'", i, mirrors[0])
		for _, mirror := range mirrors[1:] {
			cmd += fmt.Sprintf(" || curl -L -f -o /tmp/pkg%d.tar.zst '%s'", i, mirror)
		}
		cmd += ") && "
	}
	cmd += "pacman -U --noconfirm /tmp/pkg*.tar.zst && rm -f /tmp/pkg*.tar.zst"

	if err := ChrootExec(logPath, chrootPath, cmd); err != nil {
		cleanupErr := ChrootExec(logPath, chrootPath, "rm -f /tmp/pkg*.tar.zst")
		if cleanupErr != nil {
			return fmt.Errorf("%w (cleanup failed: %v)", err, cleanupErr)
		}
		return err
	}

	return nil
}

// getSortedKeys returns map keys in stable alphabetical order
func getSortedKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}
