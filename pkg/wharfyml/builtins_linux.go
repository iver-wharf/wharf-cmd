package wharfyml

func listOSPossibleVarsFiles() []varFile {
	return []varFile{
		{
			path:   "/etc/iver-wharf/wharf-cmd/" + builtInVarsFile,
			source: varFileSourceConfigDir,
		},
	}
}
