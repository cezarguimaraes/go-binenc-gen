package main

type (
	Info struct {
		Subdir string `json:"subdir"`
	}

	Package struct {
		Build       string   `json:"build"`
		BuildNumber uint32   `json:"build_number"`
		Depends     []string `json:"depends"`
		License     string   `json:"license"`
		MD5         string   `json:"md5"`
		Name        string   `json:"name"`
		// Noarch      bool   `json: "noarch"`
		Sha256    string `json:"sha256"`
		Size      uint32 `json:"size"`
		Subdir    string `json:"subdir"`
		Timestamp uint64 `json:"timestamp"`
		Version   string `json:"version"`
	}

	PackageConda struct {
		Package

		Constrains    []string `json:"constrains"`
		LegacyBz2Md5  string   `json:"legacy_bz2_md5"`
		LicenseFamily string   `json:"license_family"`
	}
)

// Convert maps to arrays since it isn't supported yet
type CondaRepoData struct {
	Info            Info
	Packages        []Package
	PackagesConda   []PackageConda
	Removed         []string
	RepoDataVersion uint32
}
