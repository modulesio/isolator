package naked

import "github.com/modulesio/isolator/installer"

func (m *Manager) Uninstall(params *installer.UninstallParams) error {
	// install folder is getting wiped anyway, nothing
	// in particular to do here.
	return nil
}

