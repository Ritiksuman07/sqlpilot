package connect

import (
	"strconv"

	"github.com/charmbracelet/huh"

	"github.com/ritiksuman07/sqlpilot/internal/config"
)

func RunWizard() (config.Profile, error) {
	var profile config.Profile
	var port string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Driver").
				Options(
					huh.NewOption("Postgres", "postgres"),
					huh.NewOption("MySQL", "mysql"),
					huh.NewOption("SQLite", "sqlite"),
					huh.NewOption("DuckDB", "duckdb"),
				).
				Value(&profile.Driver),
			huh.NewInput().Title("Profile name").Value(&profile.Name),
			huh.NewInput().Title("Host").Value(&profile.Host),
			huh.NewInput().Title("Port").Value(&port),
			huh.NewInput().Title("Database").Value(&profile.Database),
			huh.NewInput().Title("Username").Value(&profile.Username),
		),
	)

	if err := form.Run(); err != nil {
		return config.Profile{}, err
	}

	if port != "" {
		if v, err := strconv.Atoi(port); err == nil {
			profile.Port = v
		}
	}

	return profile, nil
}
