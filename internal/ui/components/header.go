package components

import (
	"fmt"

	"nuther/internal/config"
	"nuther/internal/ui/styles"
)

const (
	AppName        = "Nuther"
	AppVersion     = "0.1.0"
	AppDescription = "S.M.A.R.T. Disk Health Monitor"
)

// Logo ASCII art
const logo = `
  ███╗   ██╗██╗   ██╗████████╗██╗  ██╗███████╗██████╗
  ████╗  ██║██║   ██║╚══██╔══╝██║  ██║██╔════╝██╔══██╗
  ██╔██╗ ██║██║   ██║   ██║   ███████║█████╗  ██████╔╝
  ██║╚██╗██║██║   ██║   ██║   ██╔══██║██╔══╝  ██╔══██╗
  ██║ ╚████║╚██████╔╝   ██║   ██║  ██║███████╗██║  ██║
  ╚═╝  ╚═══╝ ╚═════╝    ╚═╝   ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝`

// RenderHeader renders the application header with logo
func RenderHeader(s *styles.Styles, cfg *config.Config) string {
	if !cfg.Display.ShowLogo {
		return s.Logo.Render(fmt.Sprintf("  %s v%s - %s", AppName, AppVersion, AppDescription))
	}

	subtitle := fmt.Sprintf("  %s v%s • %s", AppName, AppVersion, AppDescription)

	return s.Logo.Render(logo) + "\n" +
		s.Subtitle.Render(subtitle) + "\n"
}
